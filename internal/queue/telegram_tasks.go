package queue

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	// TypeTelegramSourceSync resolves a configured Telegram chat and starts its synchronization.
	TypeTelegramSourceSync = "telegram:source-sync"
	// TypeTelegramDownload processes one discovered Telegram video message.
	TypeTelegramDownload = "telegram:download"
	// TypeTelegramReconcile repairs missed Telegram processing and transcode tasks.
	TypeTelegramReconcile = "telegram:reconcile"

	telegramTaskTimeout = 6 * time.Hour
)

// TelegramMediaTaskPayload carries one Telegram media identifier.
type TelegramMediaTaskPayload struct {
	MediaID string `json:"media_id"`
}

// TelegramSourceTaskPayload carries one Telegram source identifier.
type TelegramSourceTaskPayload struct {
	SourceID string `json:"source_id"`
}

// TelegramTaskEnqueuer schedules Telegram control and media work.
type TelegramTaskEnqueuer struct {
	client        *asynq.Client
	realtimeQueue string
	backfillQueue string
	controlQueue  string
}

// NewTelegramTaskEnqueuer creates an enqueuer for the three Telegram queues.
func NewTelegramTaskEnqueuer(redisAddr, redisPassword, realtimeQueue, backfillQueue, controlQueue string) *TelegramTaskEnqueuer {
	return &TelegramTaskEnqueuer{
		client:        asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword}),
		realtimeQueue: defaultTelegramQueue(realtimeQueue, "telegram-realtime"),
		backfillQueue: defaultTelegramQueue(backfillQueue, "telegram-backfill"),
		controlQueue:  defaultTelegramQueue(controlQueue, "telegram-control"),
	}
}

// Close releases the underlying Redis client.
func (e *TelegramTaskEnqueuer) Close() error {
	if e == nil || e.client == nil {
		return nil
	}
	return e.client.Close()
}

// EnqueueSourceSync schedules source resolution and synchronization.
func (e *TelegramTaskEnqueuer) EnqueueSourceSync(sourceID uuid.UUID) error {
	if sourceID == uuid.Nil {
		return fmt.Errorf("enqueue Telegram source sync: missing source id")
	}
	return e.enqueue(TypeTelegramSourceSync, TelegramSourceTaskPayload{SourceID: sourceID.String()}, e.controlQueue)
}

// EnqueueRealtime schedules a high-priority media download.
func (e *TelegramTaskEnqueuer) EnqueueRealtime(mediaID uuid.UUID) error {
	if mediaID == uuid.Nil {
		return fmt.Errorf("enqueue Telegram realtime media: missing media id")
	}
	return e.enqueue(TypeTelegramDownload, TelegramMediaTaskPayload{MediaID: mediaID.String()}, e.realtimeQueue)
}

// EnqueueBackfill schedules a background media download.
func (e *TelegramTaskEnqueuer) EnqueueBackfill(mediaID uuid.UUID) error {
	if mediaID == uuid.Nil {
		return fmt.Errorf("enqueue Telegram backfill media: missing media id")
	}
	return e.enqueue(TypeTelegramDownload, TelegramMediaTaskPayload{MediaID: mediaID.String()}, e.backfillQueue)
}

// EnqueueReconcile schedules recovery of persisted Telegram work.
func (e *TelegramTaskEnqueuer) EnqueueReconcile() error {
	return e.enqueue(TypeTelegramReconcile, struct{}{}, e.controlQueue)
}

func (e *TelegramTaskEnqueuer) enqueue(taskType string, payloadValue any, queueName string) error {
	if e == nil || e.client == nil {
		return fmt.Errorf("enqueue Telegram task: client unavailable")
	}
	payload, err := json.Marshal(payloadValue)
	if err != nil {
		return fmt.Errorf("marshal Telegram task payload: %w", err)
	}
	if _, err := e.client.Enqueue(
		asynq.NewTask(taskType, payload),
		buildTelegramTaskOptions(queueName)...,
	); err != nil {
		return fmt.Errorf("enqueue Telegram task: %w", err)
	}
	return nil
}

func buildTelegramTaskOptions(queueName string, timeoutOverride ...time.Duration) []asynq.Option {
	timeout := telegramTaskTimeout
	if len(timeoutOverride) > 0 && timeoutOverride[0] > 0 {
		timeout = timeoutOverride[0]
	}
	return []asynq.Option{
		asynq.MaxRetry(3),
		asynq.ProcessIn(2 * time.Second),
		asynq.Queue(strings.TrimSpace(queueName)),
		asynq.Timeout(timeout),
	}
}

func defaultTelegramQueue(value, fallback string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	return fallback
}
