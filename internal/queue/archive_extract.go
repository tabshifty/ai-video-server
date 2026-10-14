package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type ArchiveExtractPayload struct {
	BatchID  string `json:"batch_id"`
	Password string `json:"password,omitempty"`
}

func buildArchiveExtractTaskOptions(queue string) []asynq.Option {
	return []asynq.Option{
		asynq.MaxRetry(0),
		asynq.ProcessIn(1 * time.Second),
		asynq.Queue(queue),
		asynq.Timeout(12 * time.Hour),
	}
}

func (e *Enqueuer) EnqueueArchiveExtract(payloadIn ArchiveExtractPayload) error {
	payload, err := json.Marshal(payloadIn)
	if err != nil {
		return fmt.Errorf("marshal archive extract payload: %w", err)
	}
	_, err = e.client.Enqueue(
		asynq.NewTask(TypeArchiveExtract, payload),
		buildArchiveExtractTaskOptions(e.queue)...,
	)
	if err != nil {
		return fmt.Errorf("enqueue archive extract task: %w", err)
	}
	return nil
}

func (p *Processor) HandleArchiveExtract(ctx context.Context, task *asynq.Task) error {
	if p.archiveImport == nil {
		return fmt.Errorf("archive extract service unavailable")
	}
	var payload ArchiveExtractPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal archive extract payload: %w", err)
	}
	batchID, err := uuid.Parse(payload.BatchID)
	if err != nil {
		return fmt.Errorf("invalid archive batch id: %w", err)
	}
	if err := p.archiveImport.ProcessExtractTask(ctx, batchID, payload.Password); err != nil {
		p.logger.Error("archive extract task failed", "batch_id", batchID.String(), "error", err)
		return err
	}
	p.logger.Info("archive extract task completed", "batch_id", batchID.String())
	return nil
}
