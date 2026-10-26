package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeArchiveWorkspace = "archive:workspace"

// EnqueueArchiveWorkspace wakes persistent work; identical queued/running wakeups coalesce.
func (e *Enqueuer) EnqueueArchiveWorkspace(id uuid.UUID) error {
	raw, err := json.Marshal(ArchiveExtractPayload{BatchID: id.String()})
	if err != nil {
		return err
	}
	_, err = e.client.Enqueue(asynq.NewTask(TypeArchiveWorkspace, raw), asynq.Queue(archiveExtractQueue), asynq.Unique(45*time.Second), asynq.MaxRetry(3), asynq.Timeout(12*time.Hour))
	if errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	return err
}

// HandleArchiveWorkspace executes persisted snapshots independently of an HTTP request.
func (p *Processor) HandleArchiveWorkspace(ctx context.Context, t *asynq.Task) error {
	svc, ok := p.archiveImport.(interface {
		ProcessArchiveWorkspace(context.Context, uuid.UUID) error
	})
	if !ok {
		return fmt.Errorf("archive workspace service unavailable")
	}
	var payload ArchiveExtractPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	id, err := uuid.Parse(payload.BatchID)
	if err != nil {
		return err
	}
	return svc.ProcessArchiveWorkspace(ctx, id)
}
