package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// CloudDownloadPayload identifies one cloud download task to advance.
type CloudDownloadPayload struct {
	TaskID int64 `json:"task_id"`
}

// CloudDownload advances one cloud download task.
type CloudDownload interface {
	Process(ctx context.Context, taskID int64) error
	// MarkFailed records a terminal failure once asynq has exhausted retries.
	MarkFailed(ctx context.Context, taskID int64, message string) error
}

// SetCloudDownloadService wires the cloud download processor.
func (p *Processor) SetCloudDownloadService(svc CloudDownload) {
	p.cloudDownload = svc
}

// CloudDownloadTaskID is the asynq uniqueness key for one task attempt.
// retry_count participates so a manual retry gets a fresh ID while a still
// in-flight or archived attempt keeps its own key.
func CloudDownloadTaskID(taskID int64, retryCount int) string {
	return fmt.Sprintf("cloud:download:%d:%d", taskID, retryCount)
}

// EnqueueCloudDownload schedules a cloud download task on the dedicated cloud
// queue. The queue has a low weight so long downloads cannot occupy every
// worker slot needed by transcodes and archive extraction.
func (e *Enqueuer) EnqueueCloudDownload(taskID int64, retryCount int) error {
	payload, err := json.Marshal(CloudDownloadPayload{TaskID: taskID})
	if err != nil {
		return fmt.Errorf("marshal cloud download payload: %w", err)
	}
	_, err = e.client.Enqueue(
		asynq.NewTask(TypeCloudDownload, payload),
		buildCloudDownloadTaskOptions(e.cloudQueue, CloudDownloadTaskID(taskID, retryCount))...,
	)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		// Already pending, active or scheduled for this attempt; that is
		// exactly the deduplication the reconciler needs.
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue cloud download task: %w", err)
	}
	return nil
}

func buildCloudDownloadTaskOptions(queue, taskID string) []asynq.Option {
	return []asynq.Option{
		asynq.Queue(queue),
		asynq.TaskID(taskID),
		asynq.MaxRetry(2),
		asynq.Timeout(6 * time.Hour),
	}
}

// HandleCloudDownload runs one cloud download task. On the final asynq attempt
// the row is moved to failed so a permanent queue failure cannot leave the UI
// stuck in a non-terminal status forever.
func (p *Processor) HandleCloudDownload(ctx context.Context, task *asynq.Task) error {
	var payload CloudDownloadPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal cloud download payload: %w", err)
	}
	if p.cloudDownload == nil {
		return fmt.Errorf("cloud download service is not configured")
	}
	err := p.cloudDownload.Process(ctx, payload.TaskID)
	if err == nil {
		return nil
	}
	retryCount, hasRetryCount := asynq.GetRetryCount(ctx)
	maxRetry, hasMaxRetry := asynq.GetMaxRetry(ctx)
	if hasRetryCount && hasMaxRetry && retryCount >= maxRetry {
		markCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()
		if markErr := p.cloudDownload.MarkFailed(markCtx, payload.TaskID, err.Error()); markErr != nil {
			if p.logger != nil {
				p.logger.Error("mark cloud download failed", "task_id", payload.TaskID, "error", markErr)
			}
		}
		return nil
	}
	return err
}
