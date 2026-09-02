package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"video-server/internal/repository"
	"video-server/internal/services"
)

// TelegramProcessor connects Telegram Asynq tasks to the ingestion service.
type TelegramProcessor struct {
	ingestion      services.TelegramIngestion
	logger         *slog.Logger
	reconcileLimit int
	downloadSlots  chan struct{}
}

// NewTelegramProcessor creates a processor for Telegram control and download tasks.
func NewTelegramProcessor(ingestion services.TelegramIngestion, logger *slog.Logger, downloadConcurrency ...int) *TelegramProcessor {
	concurrency := 1
	if len(downloadConcurrency) > 0 && downloadConcurrency[0] > 0 {
		concurrency = downloadConcurrency[0]
	}
	return &TelegramProcessor{
		ingestion:      ingestion,
		logger:         logger,
		reconcileLimit: 100,
		downloadSlots:  make(chan struct{}, concurrency),
	}
}

// Register installs all Telegram task handlers on an Asynq multiplexer.
func (p *TelegramProcessor) Register(mux *asynq.ServeMux) {
	if p == nil || mux == nil {
		return
	}
	mux.HandleFunc(TypeTelegramSourceSync, p.HandleSourceSync)
	mux.HandleFunc(TypeTelegramDownload, p.HandleDownload)
	mux.HandleFunc(TypeTelegramReconcile, p.HandleReconcile)
}

// HandleSourceSync processes a source synchronization task.
func (p *TelegramProcessor) HandleSourceSync(ctx context.Context, task *asynq.Task) error {
	payload, err := decodeTelegramSourcePayload(task)
	if err != nil {
		return nonRetryableTelegramError(err)
	}
	sourceID, err := uuid.Parse(payload.SourceID)
	if err != nil || sourceID == uuid.Nil {
		if err == nil {
			err = errors.New("Telegram source ID is empty")
		}
		return nonRetryableTelegramError(fmt.Errorf("invalid Telegram source ID: %w", err))
	}
	if p == nil || p.ingestion == nil {
		return errors.New("Telegram ingestion processor unavailable")
	}
	err = p.ingestion.SyncSource(ctx, sourceID)
	return normalizeTelegramTaskError(err)
}

// HandleDownload processes one realtime or historical Telegram download task.
func (p *TelegramProcessor) HandleDownload(ctx context.Context, task *asynq.Task) error {
	payload, err := decodeTelegramMediaPayload(task)
	if err != nil {
		return nonRetryableTelegramError(err)
	}
	mediaID, err := uuid.Parse(payload.MediaID)
	if err != nil || mediaID == uuid.Nil {
		if err == nil {
			err = errors.New("Telegram media ID is empty")
		}
		return nonRetryableTelegramError(fmt.Errorf("invalid Telegram media ID: %w", err))
	}
	if p == nil || p.ingestion == nil {
		return errors.New("Telegram ingestion processor unavailable")
	}
	if p.downloadSlots != nil {
		select {
		case p.downloadSlots <- struct{}{}:
			defer func() { <-p.downloadSlots }()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	err = p.ingestion.ProcessMedia(ctx, mediaID)
	return normalizeTelegramTaskError(err)
}

// HandleReconcile repairs Telegram transcode tasks that were not enqueued.
func (p *TelegramProcessor) HandleReconcile(ctx context.Context, task *asynq.Task) error {
	if task == nil {
		return nonRetryableTelegramError(errors.New("Telegram reconcile task is nil"))
	}
	var payload map[string]any
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nonRetryableTelegramError(fmt.Errorf("decode Telegram reconcile payload: %w", err))
	}
	if p == nil || p.ingestion == nil {
		return errors.New("Telegram ingestion processor unavailable")
	}
	limit := p.reconcileLimit
	if limit <= 0 {
		limit = 100
	}
	var errs []error
	if err := p.ingestion.ReconcileDownloads(ctx, limit); err != nil {
		errs = append(errs, err)
	}
	if err := p.ingestion.ReconcileTranscodes(ctx, limit); err != nil {
		errs = append(errs, err)
	}
	return normalizeTelegramTaskError(errors.Join(errs...))
}

func decodeTelegramSourcePayload(task *asynq.Task) (TelegramSourceTaskPayload, error) {
	if task == nil {
		return TelegramSourceTaskPayload{}, errors.New("Telegram source task is nil")
	}
	var payload TelegramSourceTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return TelegramSourceTaskPayload{}, fmt.Errorf("decode Telegram source payload: %w", err)
	}
	if payload.SourceID == "" {
		return TelegramSourceTaskPayload{}, errors.New("Telegram source payload is missing source_id")
	}
	return payload, nil
}

func decodeTelegramMediaPayload(task *asynq.Task) (TelegramMediaTaskPayload, error) {
	if task == nil {
		return TelegramMediaTaskPayload{}, errors.New("Telegram media task is nil")
	}
	var payload TelegramMediaTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return TelegramMediaTaskPayload{}, fmt.Errorf("decode Telegram media payload: %w", err)
	}
	if payload.MediaID == "" {
		return TelegramMediaTaskPayload{}, errors.New("Telegram media payload is missing media_id")
	}
	return payload, nil
}

func normalizeTelegramTaskError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrTelegramSourceNotFound) || errors.Is(err, repository.ErrTelegramMediaNotFound) {
		return nonRetryableTelegramError(err)
	}
	return err
}

func nonRetryableTelegramError(err error) error {
	if err == nil {
		return asynq.SkipRetry
	}
	return fmt.Errorf("%w: %w", asynq.SkipRetry, err)
}
