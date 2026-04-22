package app

import (
	"context"
	"testing"
	"time"

	"workerfleet/internal/domain"
)

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
	if err := runtime.store.UpsertWorkerSnapshot(previous); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}

	if err := runtime.SweepWorkerStatuses(now); err != nil {
		t.Fatalf("sweep: %v", err)
	}

	current, ok, err := runtime.store.GetWorkerSnapshot("worker-1")
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
	events, err := runtime.store.ListWorkerEvents("worker-1")
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
