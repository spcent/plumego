package app

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"workerfleet/internal/domain"
)

type recordingRuntimeErrorObserver struct {
	mu         sync.Mutex
	operations []string
	errs       []error
}

func (o *recordingRuntimeErrorObserver) ObserveRuntimeError(operation string, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.operations = append(o.operations, operation)
	o.errs = append(o.errs, err)
}

func (o *recordingRuntimeErrorObserver) snapshot() ([]string, []error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]string(nil), o.operations...), append([]error(nil), o.errs...)
}

func TestSweepWorkerStatusesMarksExpiredHeartbeatOffline(t *testing.T) {
	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})

	previous := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a"},
		Runtime: domain.WorkerRuntime{
			ProcessAlive:    true,
			AcceptingTasks:  true,
			LastHeartbeatAt: now.Add(-2 * time.Minute),
			LastSeenAt:      now.Add(-2 * time.Minute),
		},
		Status:              domain.WorkerStatusOnline,
		StatusReason:        "ready",
		LastStatusChangedAt: now.Add(-3 * time.Minute),
	}
	if err := runtime.shell.loops.store.UpsertWorkerSnapshot(context.Background(), previous); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}

	if err := runtime.SweepWorkerStatuses(context.Background(), now); err != nil {
		t.Fatalf("sweep: %v", err)
	}

	current, ok, err := runtime.shell.loops.store.GetWorkerSnapshot(context.Background(), "worker-1")
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if !ok {
		t.Fatalf("snapshot not found")
	}
	if current.Status != domain.WorkerStatusOffline {
		t.Fatalf("status = %q, want offline", current.Status)
	}
	if current.StatusReason != "heartbeat_expired" {
		t.Fatalf("status reason = %q, want heartbeat_expired", current.StatusReason)
	}
	events, err := runtime.shell.loops.store.ListWorkerEvents(context.Background(), "worker-1")
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 || events[0].Type != domain.EventWorkerOffline {
		t.Fatalf("events = %#v, want one worker_offline event", events)
	}
}

func TestStartLoopsDisabledIsSafe(t *testing.T) {
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})

	stop, err := runtime.StartLoops(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("start loops: %v", err)
	}
	stop()
}

func TestStartLoopsStopsStatusSweeper(t *testing.T) {
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})
	cfg := DefaultConfig()
	cfg.Runtime.StatusSweepEnabled = true
	cfg.Runtime.StatusSweepInterval = time.Millisecond

	stop, err := runtime.StartLoops(context.Background(), cfg)
	if err != nil {
		t.Fatalf("start loops: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	stop()
}

func TestStartLoopsReportsStatusSweepErrors(t *testing.T) {
	observer := &recordingRuntimeErrorObserver{}
	runtime := &Runtime{
		shell: runtimeShell{
			loops: &LoopRunner{errors: observer},
		},
	}
	cfg := DefaultConfig()
	cfg.Runtime.StatusSweepEnabled = true
	cfg.Runtime.StatusSweepInterval = time.Millisecond

	stop, err := runtime.StartLoops(context.Background(), cfg)
	if err != nil {
		t.Fatalf("start loops: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	stop()

	operations, errs := observer.snapshot()
	if len(operations) == 0 {
		t.Fatalf("runtime errors were not reported")
	}
	if operations[0] != "status_sweep" {
		t.Fatalf("operation = %q, want status_sweep", operations[0])
	}
	if !errors.Is(errs[0], errWorkerfleetStoreNotConfigured) {
		t.Fatalf("error = %v, want store not configured", errs[0])
	}
}

func TestLoopRunnerUsesInjectedRuntimeDependencies(t *testing.T) {
	observer := &recordingRuntimeErrorObserver{}
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})

	called := 0
	runtime.shell.loops.errors = observer
	runtime.shell.loops.inventorySyncerFn = func(Config) (inventorySyncer, error) {
		called++
		return inventorySyncerFunc(func(context.Context) (string, error) {
			return "", errors.New("sync failed")
		}), nil
	}

	cfg := DefaultConfig()
	cfg.Runtime.KubeSyncEnabled = true
	cfg.Runtime.KubeSyncInterval = time.Millisecond

	stop, err := runtime.StartLoops(context.Background(), cfg)
	if err != nil {
		t.Fatalf("start loops: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	stop()

	if called == 0 {
		t.Fatalf("expected injected inventory syncer to be used")
	}
	operations, errs := observer.snapshot()
	if len(operations) == 0 || operations[0] != "kube_sync" {
		t.Fatalf("reported operations = %#v, want kube_sync", operations)
	}
	if errs[0] == nil || errs[0].Error() != "sync failed" {
		t.Fatalf("error = %v, want sync failed", errs[0])
	}
}

type inventorySyncerFunc func(context.Context) (string, error)

func (fn inventorySyncerFunc) SyncOnce(ctx context.Context) (string, error) {
	return fn(ctx)
}

func TestLoopRunnerPreventsOverlappingExecutions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var current int32
	var maxConcurrent int32
	started := make(chan struct{}, 1)
	release := make(chan struct{})

	startManagedLoop(ctx, &wg, loopExecutionSettings{
		Name:              "status_sweep",
		Interval:          2 * time.Millisecond,
		Timeout:           200 * time.Millisecond,
		FailureBackoff:    5 * time.Millisecond,
		MaxFailureBackoff: 10 * time.Millisecond,
	}, func(string, error) {}, func(ctx context.Context) error {
		active := atomic.AddInt32(&current, 1)
		updateMaxConcurrent(&maxConcurrent, active)
		select {
		case started <- struct{}{}:
		default:
		}
		select {
		case <-release:
		case <-ctx.Done():
		}
		atomic.AddInt32(&current, -1)
		return nil
	})

	select {
	case <-started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("loop did not start")
	}
	time.Sleep(20 * time.Millisecond)
	close(release)
	cancel()
	wg.Wait()

	if got := atomic.LoadInt32(&maxConcurrent); got != 1 {
		t.Fatalf("max concurrent = %d, want 1", got)
	}
}

func TestLoopRunnerSkipsWorkWhenLeaseNotAcquired(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var calls int32
	startManagedLoop(ctx, &wg, loopExecutionSettings{
		Name:              "status_sweep",
		Interval:          time.Millisecond,
		Timeout:           20 * time.Millisecond,
		FailureBackoff:    time.Millisecond,
		MaxFailureBackoff: time.Millisecond,
		Lease:             denyingLoopLease{},
	}, func(string, error) {}, func(context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	time.Sleep(20 * time.Millisecond)
	cancel()
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("loop calls = %d, want 0 when lease is not acquired", got)
	}
}

type denyingLoopLease struct{}

func (denyingLoopLease) TryAcquire(context.Context, string) (func(), bool, error) {
	return nil, false, nil
}

func TestLoopRunnerCancelsSlowIterationOnTimeout(t *testing.T) {
	observer := &recordingRuntimeErrorObserver{}
	runner := &LoopRunner{errors: observer}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	canceled := make(chan struct{}, 1)
	startManagedLoop(ctx, &wg, loopExecutionSettings{
		Name:              "status_sweep",
		Interval:          time.Hour,
		Timeout:           15 * time.Millisecond,
		FailureBackoff:    5 * time.Millisecond,
		MaxFailureBackoff: 10 * time.Millisecond,
	}, runner.reportRuntimeError, func(ctx context.Context) error {
		<-ctx.Done()
		select {
		case canceled <- struct{}{}:
		default:
		}
		return ctx.Err()
	})

	select {
	case <-canceled:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("loop timeout did not cancel iteration")
	}
	cancel()
	wg.Wait()

	operations, errs := observer.snapshot()
	if len(operations) == 0 || operations[0] != "status_sweep" {
		t.Fatalf("reported operations = %#v, want status_sweep", operations)
	}
	if !errors.Is(errs[0], context.DeadlineExceeded) {
		t.Fatalf("error = %v, want context deadline exceeded", errs[0])
	}
}

func TestLoopRunnerBacksOffAfterFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	attempts := make([]time.Time, 0, 2)
	done := make(chan struct{})

	startManagedLoop(ctx, &wg, loopExecutionSettings{
		Name:              "kube_sync",
		Interval:          2 * time.Millisecond,
		Timeout:           50 * time.Millisecond,
		FailureBackoff:    25 * time.Millisecond,
		MaxFailureBackoff: 25 * time.Millisecond,
	}, func(string, error) {}, func(context.Context) error {
		mu.Lock()
		attempts = append(attempts, time.Now())
		count := len(attempts)
		mu.Unlock()
		if count >= 2 {
			select {
			case done <- struct{}{}:
			default:
			}
		}
		return errors.New("boom")
	})

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected two loop attempts")
	}
	cancel()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(attempts) < 2 {
		t.Fatalf("attempt count = %d, want at least 2", len(attempts))
	}
	gap := attempts[1].Sub(attempts[0])
	if gap < 20*time.Millisecond {
		t.Fatalf("attempt gap = %v, want at least 20ms backoff", gap)
	}
}

func TestRuntimeLoopSettingsRemainIndependent(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Runtime.KubeSyncInterval = 11 * time.Second
	cfg.Runtime.StatusSweepInterval = 12 * time.Second
	cfg.Runtime.AlertEvaluationInterval = 13 * time.Second

	kubeSettings := cfg.Runtime.kubeSyncLoopSettings()
	statusSettings := cfg.Runtime.statusSweepLoopSettings()
	alertSettings := cfg.Runtime.alertEvaluationLoopSettings()

	if kubeSettings.Name != "kube_sync" || kubeSettings.Interval != 11*time.Second {
		t.Fatalf("kube settings = %#v", kubeSettings)
	}
	if statusSettings.Name != "status_sweep" || statusSettings.Interval != 12*time.Second {
		t.Fatalf("status settings = %#v", statusSettings)
	}
	if alertSettings.Name != "alert_evaluate" || alertSettings.Interval != 13*time.Second {
		t.Fatalf("alert settings = %#v", alertSettings)
	}
}

func updateMaxConcurrent(target *int32, value int32) {
	for {
		current := atomic.LoadInt32(target)
		if value <= current {
			return
		}
		if atomic.CompareAndSwapInt32(target, current, value) {
			return
		}
	}
}
