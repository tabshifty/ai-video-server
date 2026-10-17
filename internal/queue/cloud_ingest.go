package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"video-server/internal/models"
)

// CloudIngestPayload identifies one explicit cloud download import. The
// import form travels with the payload so the worker can create the target
// without persisting archive passwords in the database.
type CloudIngestPayload struct {
	ImportID int64                           `json:"import_id"`
	Attempt  int                             `json:"attempt"`
	Input    models.CloudDownloadImportInput `json:"input"`
}

// CloudDownloadImport advances one explicit cloud download import.
type CloudDownloadImport interface {
	Process(ctx context.Context, importID int64, attempt int, input models.CloudDownloadImportInput) error
	// MarkFailed records a terminal failure once asynq has exhausted retries.
	MarkFailed(ctx context.Context, importID int64, attempt int, message string) error
}

// SetCloudDownloadImportService wires the cloud ingest processor.
func (p *Processor) SetCloudDownloadImportService(svc CloudDownloadImport) {
	p.cloudImport = svc
}

// CloudIngestTaskID is the asynq uniqueness key for one import attempt.
func CloudIngestTaskID(importID int64, attempt int) string {
	return fmt.Sprintf("cloud:ingest:%d:%d", importID, attempt)
}

// EnqueueCloudIngest schedules an explicit import on the dedicated cloud
// queue. Downloads and imports share the low-weight queue.
func (e *Enqueuer) EnqueueCloudIngest(importID int64, attempt int, input models.CloudDownloadImportInput) error {
	payload, err := json.Marshal(CloudIngestPayload{
		ImportID: importID,
		Attempt:  attempt,
		Input:    input,
	})
	if err != nil {
		return fmt.Errorf("marshal cloud ingest payload: %w", err)
	}
	_, err = e.client.Enqueue(
		asynq.NewTask(TypeCloudIngest, payload),
		buildCloudIngestTaskOptions(e.cloudQueue, CloudIngestTaskID(importID, attempt))...,
	)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue cloud ingest task: %w", err)
	}
	return nil
}

func buildCloudIngestTaskOptions(queue, taskID string) []asynq.Option {
	return []asynq.Option{
		asynq.Queue(queue),
		asynq.TaskID(taskID),
		asynq.MaxRetry(2),
		asynq.Timeout(2 * time.Hour),
	}
}

// HandleCloudIngest runs one import task. On the final asynq attempt the row
// is moved to failed so a permanent queue failure cannot leave it active.
func (p *Processor) HandleCloudIngest(ctx context.Context, task *asynq.Task) error {
	var payload CloudIngestPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal cloud ingest payload: %w", err)
	}
	if p.cloudImport == nil {
		return fmt.Errorf("cloud download import service is not configured")
	}
	err := p.cloudImport.Process(ctx, payload.ImportID, payload.Attempt, payload.Input)
	if err == nil {
		return nil
	}
	retryCount, hasRetryCount := asynq.GetRetryCount(ctx)
	maxRetry, hasMaxRetry := asynq.GetMaxRetry(ctx)
	if hasRetryCount && hasMaxRetry && retryCount >= maxRetry {
		markCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()
		if markErr := p.cloudImport.MarkFailed(markCtx, payload.ImportID, payload.Attempt, err.Error()); markErr != nil {
			if p.logger != nil {
				p.logger.Error("mark cloud download import failed", "import_id", payload.ImportID, "error", markErr)
			}
		}
		return nil
	}
	return err
}
