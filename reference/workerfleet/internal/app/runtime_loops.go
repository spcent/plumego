package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"workerfleet/internal/domain"
	"workerfleet/internal/platform/kube"
	workerfleetmetrics "workerfleet/internal/platform/metrics"
)

var errWorkerfleetStoreNotConfigured = errors.New("workerfleet store is not configured")

func NewLoopRunner(store runtimeStore, policy domain.StatusPolicy, metrics *workerfleetmetrics.Observer, errors RuntimeErrorObserver) *LoopRunner {
	runner := &LoopRunner{
		store:   store,
		policy:  policy,
		metrics: metrics,
		errors:  errors,
	}
	runner.inventorySyncerFn = runner.newInventorySyncer
	return runner
}

func (r *Runtime) StartLoops(ctx context.Context, cfg Config) (func(), error) {
	if r == nil || r.shell.loops == nil {
		return func() {}, nil
	}
	return r.shell.loops.Start(ctx, cfg)
}

func (l *LoopRunner) Start(ctx context.Context, cfg Config) (func(), error) {
	if l == nil {
		return func() {}, nil
	}
	loopCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup

	if cfg.Runtime.KubeSyncEnabled {
		syncer, err := l.inventorySyncerFn(cfg)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("create kubernetes client: %w", err)
		}
		startLoop(loopCtx, &wg, cfg.Runtime.KubeSyncInterval, func(ctx context.Context) {
			_, err := syncer.SyncOnce(ctx)
			if err != nil {
				l.reportRuntimeError("kube_sync", err)
			}
		})
	}

	if cfg.Runtime.StatusSweepEnabled {
		startLoop(loopCtx, &wg, cfg.Runtime.StatusSweepInterval, func(context.Context) {
			err := l.SweepWorkerStatuses(ctx, time.Now().UTC())
			if err != nil {
				l.reportRuntimeError("status_sweep", err)
			}
		})
	}

	return func() {
		cancel()
		wg.Wait()
	}, nil
}

func (r *Runtime) SweepWorkerStatuses(ctx context.Context, now time.Time) error {
	if r == nil || r.shell.loops == nil {
		return errWorkerfleetStoreNotConfigured
	}
	return r.shell.loops.SweepWorkerStatuses(ctx, now)
}

func (l *LoopRunner) SweepWorkerStatuses(ctx context.Context, now time.Time) error {
	if l == nil || l.store == nil {
		return errWorkerfleetStoreNotConfigured
	}
	started := time.Now()
	snapshots, err := l.store.ListCurrentWorkerSnapshots(ctx)
	if err != nil {
		return err
	}
	for _, previous := range snapshots {
		current := previous
		status, reason := domain.EvaluateWorkerStatus(current, now, l.policy)
		if current.Status == status && current.StatusReason == reason {
			if l.metrics != nil {
				l.metrics.ObserveWorkerSnapshot(previous, current)
			}
			continue
		}
		current.Status = status
		current.StatusReason = reason
		if previous.Status != status {
			current.LastStatusChangedAt = now
		}
		if err := l.store.UpsertWorkerSnapshot(ctx, current); err != nil {
			return err
		}
		if previous.Status != status {
			if err := l.appendStatusEvent(ctx, previous, current, now); err != nil {
				return err
			}
		}
		if l.metrics != nil {
			l.metrics.ObserveWorkerSnapshot(previous, current)
		}
	}
	if l.metrics != nil {
		l.metrics.ObserveWorkerReportApplied("status_sweep", time.Since(started))
	}
	return nil
}

func (l *LoopRunner) appendStatusEvent(ctx context.Context, previous domain.WorkerSnapshot, current domain.WorkerSnapshot, now time.Time) error {
	eventType, ok := statusEventType(current.Status)
	if !ok {
		return nil
	}
	return l.store.AppendWorkerEvent(ctx, domain.DomainEvent{
		Type:       eventType,
		OccurredAt: now,
		WorkerID:   current.Identity.WorkerID,
		Reason:     current.StatusReason,
		Attributes: map[string]string{
			"from_status": string(previous.Status),
			"to_status":   string(current.Status),
		},
	})
}

func (l *LoopRunner) reportRuntimeError(operation string, err error) {
	if l == nil || l.errors == nil || err == nil {
		return
	}
	l.errors.ObserveRuntimeError(operation, err)
}

func (l *LoopRunner) newInventorySyncer(cfg Config) (inventorySyncer, error) {
	client, err := kube.NewClient(kube.Config{
		APIHost:         cfg.Kube.APIHost,
		BearerToken:     cfg.Kube.BearerToken,
		Namespace:       cfg.Kube.Namespace,
		LabelSelector:   cfg.Kube.LabelSelector,
		WorkerContainer: cfg.Kube.WorkerContainer,
	})
	if err != nil {
		return nil, err
	}
	return kube.NewInventorySync(client, l.store, cfg.Kube.WorkerContainer, l.policy, kube.WithMetricsObserver(l.metrics)), nil
}

func statusEventType(status domain.WorkerStatus) (domain.EventType, bool) {
	switch status {
	case domain.WorkerStatusOnline:
		return domain.EventWorkerOnline, true
	case domain.WorkerStatusDegraded:
		return domain.EventWorkerDegraded, true
	case domain.WorkerStatusOffline:
		return domain.EventWorkerOffline, true
	default:
		return "", false
	}
}

func startLoop(ctx context.Context, wg *sync.WaitGroup, interval time.Duration, fn func(context.Context)) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn(ctx)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fn(ctx)
			}
		}
	}()
}
