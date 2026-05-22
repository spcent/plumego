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
		lease:   nopLoopLease{},
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
		settings := cfg.Runtime.kubeSyncLoopSettings()
		settings.Lease = l.lease
		startManagedLoop(loopCtx, &wg, settings, l.reportRuntimeError, func(ctx context.Context) error {
			if watcher, ok := syncer.(inventoryWatcher); ok {
				_, err := watcher.SyncWatch(ctx)
				return err
			}
			_, err := syncer.SyncOnce(ctx)
			return err
		})
	}

	if cfg.Runtime.StatusSweepEnabled {
		settings := cfg.Runtime.statusSweepLoopSettings()
		settings.Lease = l.lease
		startManagedLoop(loopCtx, &wg, settings, l.reportRuntimeError, func(loopRunCtx context.Context) error {
			return l.SweepWorkerStatuses(loopRunCtx, time.Now().UTC())
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

func startManagedLoop(ctx context.Context, wg *sync.WaitGroup, settings loopExecutionSettings, report func(string, error), fn func(context.Context) error) {
	normalized := normalizeLoopExecutionSettings(settings)
	resultCh := make(chan loopExecutionResult, 1)
	timer := time.NewTimer(0)
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	resetLoopTimer(timer, 0)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer timer.Stop()

		running := false
		backoff := normalized.FailureBackoff
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-resultCh:
				running = false
				if result.timedOut {
					report(normalized.Name, context.DeadlineExceeded)
					resetLoopTimer(timer, backoff)
					backoff = nextLoopBackoff(backoff, normalized)
					continue
				}
				if result.err != nil {
					report(normalized.Name, result.err)
					resetLoopTimer(timer, backoff)
					backoff = nextLoopBackoff(backoff, normalized)
					continue
				}
				backoff = normalized.FailureBackoff
				resetLoopTimer(timer, normalized.Interval)
			case <-timer.C:
				if running {
					resetLoopTimer(timer, normalized.Interval)
					continue
				}
				release, acquired, err := normalized.Lease.TryAcquire(ctx, normalized.Name)
				if err != nil {
					report(normalized.Name, err)
					resetLoopTimer(timer, backoff)
					backoff = nextLoopBackoff(backoff, normalized)
					continue
				}
				if !acquired {
					resetLoopTimer(timer, normalized.Interval)
					continue
				}
				running = true
				go runLoopIteration(ctx, normalized, fn, release, resultCh)
			}
		}
	}()
}

func runLoopIteration(parent context.Context, settings loopExecutionSettings, fn func(context.Context) error, release func(), resultCh chan<- loopExecutionResult) {
	runCtx := parent
	cancel := func() {}
	if settings.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(parent, settings.Timeout)
	}
	defer cancel()
	if release != nil {
		defer release()
	}
	err := fn(runCtx)
	result := loopExecutionResult{err: err}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		result.timedOut = true
		if result.err == nil || errors.Is(result.err, context.Canceled) {
			result.err = context.DeadlineExceeded
		}
	}
	select {
	case resultCh <- result:
	case <-parent.Done():
	}
}

func normalizeLoopExecutionSettings(settings loopExecutionSettings) loopExecutionSettings {
	normalized := settings
	if normalized.Interval <= 0 {
		normalized.Interval = 30 * time.Second
	}
	if normalized.Timeout <= 0 {
		normalized.Timeout = defaultLoopTimeout
	}
	if normalized.FailureBackoff <= 0 {
		normalized.FailureBackoff = defaultLoopFailureBackoff
	}
	if normalized.MaxFailureBackoff < normalized.FailureBackoff {
		normalized.MaxFailureBackoff = maxDuration(defaultLoopMaxBackoff, normalized.FailureBackoff)
	}
	if normalized.Lease == nil {
		normalized.Lease = nopLoopLease{}
	}
	return normalized
}

func nextLoopBackoff(current time.Duration, settings loopExecutionSettings) time.Duration {
	if current <= 0 {
		return settings.FailureBackoff
	}
	next := current * 2
	if next > settings.MaxFailureBackoff {
		next = settings.MaxFailureBackoff
	}
	if next < settings.FailureBackoff {
		next = settings.FailureBackoff
	}
	return next
}

func resetLoopTimer(timer *time.Timer, delay time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(delay)
}

func maxDuration(left time.Duration, right time.Duration) time.Duration {
	if left >= right {
		return left
	}
	return right
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
