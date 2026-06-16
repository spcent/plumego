package ai

import (
	"context"
	"time"

	plumelog "github.com/spcent/plumego/log"
)

// Worker polls the ai_tasks queue and processes pending tasks.
type Worker struct {
	svc        *Service
	logger     plumelog.StructuredLogger
	maxRetries int
}

func NewWorker(svc *Service, logger plumelog.StructuredLogger, maxRetries int) *Worker {
	return &Worker{svc: svc, logger: logger, maxRetries: maxRetries}
}

// Run starts the worker loop; it returns when ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processOne(ctx)
		}
	}
}

func (w *Worker) processOne(ctx context.Context) {
	task, err := w.svc.repo.ClaimPendingTask()
	if err != nil {
		w.logger.Error("ai worker: claim task", plumelog.Fields{"err": err.Error()})
		return
	}
	if task == nil {
		return
	}

	w.logger.Info("ai worker: processing task", plumelog.Fields{"id": task.ID, "type": task.TaskType, "retry": task.RetryCount})

	err = w.svc.ProcessTask(ctx, task)
	if err != nil {
		retry := task.RetryCount < w.maxRetries
		w.logger.Error("ai worker: task failed", plumelog.Fields{
			"id":          task.ID,
			"err":         err.Error(),
			"retry_count": task.RetryCount,
			"will_retry":  retry,
		})
		if ferr := w.svc.repo.FailTask(task.ID, err.Error(), retry); ferr != nil {
			w.logger.Error("ai worker: fail task record", plumelog.Fields{"err": ferr.Error()})
		}
		if retry {
			// Exponential backoff: 2^retryCount seconds, capped at 60s.
			backoff := time.Duration(1<<uint(task.RetryCount)) * time.Second
			if backoff > 60*time.Second {
				backoff = 60 * time.Second
			}
			w.logger.Info("ai worker: backing off before retry", plumelog.Fields{
				"id":              task.ID,
				"backoff_seconds": backoff.Seconds(),
			})
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
		} else {
			w.logger.Warn("ai worker: task moved to dead letter", plumelog.Fields{
				"id":   task.ID,
				"type": task.TaskType,
			})
		}
		return
	}

	w.logger.Info("ai worker: task completed", plumelog.Fields{"id": task.ID})
}
