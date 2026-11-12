package queue

import (
	"context"
	"time"
)

// RunRotationReconciler restores pending work after queue loss or a hard restart.
func (p *Processor) RunRotationReconciler(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		jobs, err := p.repo.ActiveRotations(runCtx)
		if err == nil {
			for _, j := range jobs {
				if e := p.enqueuer.EnqueueTranscode(TranscodePayload{RotationJobID: j.ID.String()}); e != nil {
					p.logger.Warn("rotation recovery enqueue failed", "job_id", j.ID, "error", e)
				}
			}
		} else {
			p.logger.Warn("rotation reconciliation failed", "error", err)
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
