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

	w.logger.Info("ai worker: processing task", plumelog.Fields{"id": task.ID, "type": task.TaskType})

	err = w.svc.ProcessTask(ctx, task)
	if err != nil {
		w.logger.Error("ai worker: task failed", plumelog.Fields{"id": task.ID, "err": err.Error()})
		retry := task.RetryCount < w.maxRetries
		if ferr := w.svc.repo.FailTask(task.ID, err.Error(), retry); ferr != nil {
			w.logger.Error("ai worker: fail task record", plumelog.Fields{"err": ferr.Error()})
		}
		return
	}

	w.logger.Info("ai worker: task completed", plumelog.Fields{"id": task.ID})
}
