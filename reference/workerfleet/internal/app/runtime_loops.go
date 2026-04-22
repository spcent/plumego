package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"workerfleet/internal/domain"
	"workerfleet/internal/platform/kube"
)

func (r *Runtime) StartLoops(ctx context.Context, cfg Config) (func(), error) {
	if r == nil {
		return func() {}, nil
	}
	loopCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup

	if cfg.Runtime.KubeSyncEnabled {
		client, err := kube.NewClient(kube.Config{
			APIHost:         cfg.Kube.APIHost,
			BearerToken:     cfg.Kube.BearerToken,
			Namespace:       cfg.Kube.Namespace,
			LabelSelector:   cfg.Kube.LabelSelector,
			WorkerContainer: cfg.Kube.WorkerContainer,
		})
		if err != nil {
			cancel()
			return nil, fmt.Errorf("create kubernetes client: %w", err)
		}
		syncer := kube.NewInventorySync(client, r.store, cfg.Kube.WorkerContainer, r.policy, kube.WithMetricsObserver(r.metrics))
		startLoop(loopCtx, &wg, cfg.Runtime.KubeSyncInterval, func(ctx context.Context) {
			_, _ = syncer.SyncOnce(ctx)
		})
	}

	if cfg.Runtime.StatusSweepEnabled {
		startLoop(loopCtx, &wg, cfg.Runtime.StatusSweepInterval, func(context.Context) {
			_ = r.SweepWorkerStatuses(time.Now().UTC())
		})
	}

	return func() {
		cancel()
		wg.Wait()
	}, nil
}

func (r *Runtime) SweepWorkerStatuses(now time.Time) error {
	if r == nil || r.store == nil {
		return fmt.Errorf("workerfleet store is not configured")
	}
	started := time.Now()
	snapshots, err := r.store.ListCurrentWorkerSnapshots()
	if err != nil {
		return err
	}
	for _, previous := range snapshots {
		current := previous
		status, reason := domain.EvaluateWorkerStatus(current, now, r.policy)
		if current.Status == status && current.StatusReason == reason {
			continue
		}
		current.Status = status
		current.StatusReason = reason
		if previous.Status != status {
			current.LastStatusChangedAt = now
		}
		if err := r.store.UpsertWorkerSnapshot(current); err != nil {
			return err
		}
		if previous.Status != status {
			if err := r.appendStatusEvent(previous, current, now); err != nil {
				return err
			}
		}
		if r.metrics != nil {
			r.metrics.ObserveWorkerSnapshot(previous, current)
		}
	}
	if r.metrics != nil {
		r.metrics.ObserveWorkerReportApplied("status_sweep", time.Since(started))
	}
	return nil
}

func (r *Runtime) appendStatusEvent(previous domain.WorkerSnapshot, current domain.WorkerSnapshot, now time.Time) error {
	eventType, ok := statusEventType(current.Status)
	if !ok {
		return nil
	}
	return r.store.AppendWorkerEvent(domain.DomainEvent{
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
