package queue

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type telegramProcessorFakeIngestion struct {
	sourceID               uuid.UUID
	mediaID                uuid.UUID
	syncCalls              int
	processCalls           int
	reconcileDownloadCalls int
	reconcileCalls         int
	err                    error
}

func (s *telegramProcessorFakeIngestion) SyncSource(context.Context, uuid.UUID) error {
	s.syncCalls++
	return s.err
}

func (s *telegramProcessorFakeIngestion) ProcessMedia(context.Context, uuid.UUID) error {
	s.processCalls++
	return s.err
}

func (s *telegramProcessorFakeIngestion) ReconcileDownloads(context.Context, int) error {
	s.reconcileDownloadCalls++
	return s.err
}

func (s *telegramProcessorFakeIngestion) ReconcileTranscodes(context.Context, int) error {
	s.reconcileCalls++
	return s.err
}

func TestTelegramProcessorDispatchesSourceDownloadAndReconcileTasks(t *testing.T) {
	service := &telegramProcessorFakeIngestion{}
	processor := NewTelegramProcessor(service, nil)
	sourceID := uuid.New()
	mediaID := uuid.New()

	if err := processor.HandleSourceSync(context.Background(), asynq.NewTask(TypeTelegramSourceSync, mustTelegramJSON(t, TelegramSourceTaskPayload{SourceID: sourceID.String()}))); err != nil {
		t.Fatalf("HandleSourceSync() error = %v", err)
	}
	if err := processor.HandleDownload(context.Background(), asynq.NewTask(TypeTelegramDownload, mustTelegramJSON(t, TelegramMediaTaskPayload{MediaID: mediaID.String()}))); err != nil {
		t.Fatalf("HandleDownload() error = %v", err)
	}
	if err := processor.HandleReconcile(context.Background(), asynq.NewTask(TypeTelegramReconcile, []byte(`{}`))); err != nil {
		t.Fatalf("HandleReconcile() error = %v", err)
	}
	if service.syncCalls != 1 || service.processCalls != 1 || service.reconcileDownloadCalls != 1 || service.reconcileCalls != 1 {
		t.Fatalf("processor calls = sync:%d process:%d download-reconcile:%d transcode-reconcile:%d", service.syncCalls, service.processCalls, service.reconcileDownloadCalls, service.reconcileCalls)
	}
}

func TestTelegramProcessorMarksInvalidPayloadAsNonRetryable(t *testing.T) {
	processor := NewTelegramProcessor(&telegramProcessorFakeIngestion{}, nil)
	err := processor.HandleDownload(context.Background(), asynq.NewTask(TypeTelegramDownload, []byte(`{"media_id":"bad"}`)))
	if err == nil {
		t.Fatal("HandleDownload() error = nil, want invalid UUID")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("error = %v, want asynq.SkipRetry", err)
	}
}

func TestTelegramProcessorPropagatesServiceError(t *testing.T) {
	wantErr := errors.New("Telegram FloodWait")
	service := &telegramProcessorFakeIngestion{err: wantErr}
	processor := NewTelegramProcessor(service, nil)
	mediaID := uuid.New()
	err := processor.HandleDownload(context.Background(), asynq.NewTask(TypeTelegramDownload, mustTelegramJSON(t, TelegramMediaTaskPayload{MediaID: mediaID.String()})))
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want service error", err)
	}
}

func TestTelegramProcessorRegisterAcceptsNilMux(t *testing.T) {
	processor := NewTelegramProcessor(&telegramProcessorFakeIngestion{}, nil)
	processor.Register(nil)
}

type telegramBlockingIngestion struct {
	entered chan struct{}
	release chan struct{}
	active  atomic.Int32
	maximum atomic.Int32
}

func (s *telegramBlockingIngestion) SyncSource(context.Context, uuid.UUID) error {
	return nil
}

func (s *telegramBlockingIngestion) ProcessMedia(context.Context, uuid.UUID) error {
	active := s.active.Add(1)
	for {
		maximum := s.maximum.Load()
		if active <= maximum || s.maximum.CompareAndSwap(maximum, active) {
			break
		}
	}
	s.entered <- struct{}{}
	<-s.release
	s.active.Add(-1)
	return nil
}

func (s *telegramBlockingIngestion) ReconcileDownloads(context.Context, int) error {
	return nil
}

func (s *telegramBlockingIngestion) ReconcileTranscodes(context.Context, int) error {
	return nil
}

func TestTelegramProcessorLimitsConcurrentDownloads(t *testing.T) {
	service := &telegramBlockingIngestion{
		entered: make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	processor := NewTelegramProcessor(service, nil, 1)
	task := func(id uuid.UUID) *asynq.Task {
		return asynq.NewTask(TypeTelegramDownload, mustTelegramJSON(t, TelegramMediaTaskPayload{MediaID: id.String()}))
	}

	var group sync.WaitGroup
	group.Add(2)
	for range 2 {
		go func() {
			defer group.Done()
			if err := processor.HandleDownload(context.Background(), task(uuid.New())); err != nil {
				t.Errorf("HandleDownload() error = %v", err)
			}
		}()
	}
	select {
	case <-service.entered:
	case <-time.After(time.Second):
		t.Fatal("first download did not enter service")
	}
	if got := service.maximum.Load(); got != 1 {
		t.Fatalf("maximum concurrent downloads = %d, want 1", got)
	}
	close(service.release)
	group.Wait()
}

func mustTelegramJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return raw
}
