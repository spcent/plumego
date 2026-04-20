package mongo

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func TestStoreCurrentStateIntegration(t *testing.T) {
	uri := os.Getenv("WORKERFLEET_MONGO_TEST_URI")
	if strings.TrimSpace(uri) == "" {
		t.Skip("set WORKERFLEET_MONGO_TEST_URI to run MongoDB integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Fatalf("disconnect mongo: %v", err)
		}
	}()

	db := client.Database(fmt.Sprintf("workerfleet_test_%d", time.Now().UnixNano()))
	defer func() {
		if err := db.Drop(context.Background()); err != nil {
			t.Fatalf("drop test database: %v", err)
		}
	}()

	if err := EnsureIndexes(ctx, db); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	store, err := NewStore(db, WithOperationTimeout(5*time.Second), WithClock(func() time.Time {
		return time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	}))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	now := time.Date(2026, 4, 20, 9, 59, 0, 0, time.UTC)
	if err := store.UpsertWorkerSnapshot(domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{
			WorkerID:  "worker-1",
			Namespace: "sim",
			NodeName:  "node-a",
		},
		Runtime: domain.WorkerRuntime{
			ProcessAlive:   true,
			AcceptingTasks: false,
			LastSeenAt:     now,
		},
		Status: domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     domain.TaskPhaseRunning,
			PhaseName: "running",
			UpdatedAt: now,
		}},
	}); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}

	accepting := false
	filtered, err := store.ListWorkerSnapshots(platformstore.WorkerSnapshotFilter{
		Status:         domain.WorkerStatusOnline,
		Namespace:      "sim",
		NodeName:       "node-a",
		TaskType:       "simulation",
		AcceptingTasks: &accepting,
	})
	if err != nil {
		t.Fatalf("list snapshots: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Identity.WorkerID != "worker-1" {
		t.Fatalf("unexpected filtered snapshots %#v", filtered)
	}

	current, ok, err := store.GetTask("task-1")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if !ok || current.WorkerID != "worker-1" {
		t.Fatalf("current task = %#v ok=%v, want worker-1", current, ok)
	}

	counts := store.FleetCounts()
	if counts.TotalWorkers != 1 || counts.OnlineWorkers != 1 || counts.ActiveTaskCount != 1 {
		t.Fatalf("unexpected fleet counts %#v", counts)
	}

	if err := store.ReplaceActiveTasks("worker-1", []domain.ActiveTask{{
		TaskID:    "task-2",
		TaskType:  "simulation",
		Phase:     domain.TaskPhaseFinalizing,
		PhaseName: "finalizing",
		UpdatedAt: now.Add(time.Minute),
	}}); err != nil {
		t.Fatalf("replace active tasks: %v", err)
	}
	if _, ok, err := store.GetTask("task-1"); err != nil || ok {
		t.Fatalf("old task should be removed, ok=%v err=%v", ok, err)
	}
	if _, ok, err := store.GetTask("task-2"); err != nil || !ok {
		t.Fatalf("new task should be indexed, ok=%v err=%v", ok, err)
	}
}

func TestStoreHistoryAndAlertIntegration(t *testing.T) {
	uri := os.Getenv("WORKERFLEET_MONGO_TEST_URI")
	if strings.TrimSpace(uri) == "" {
		t.Skip("set WORKERFLEET_MONGO_TEST_URI to run MongoDB integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Fatalf("disconnect mongo: %v", err)
		}
	}()

	db := client.Database(fmt.Sprintf("workerfleet_history_test_%d", time.Now().UnixNano()))
	defer func() {
		if err := db.Drop(context.Background()); err != nil {
			t.Fatalf("drop test database: %v", err)
		}
	}()

	if err := EnsureIndexes(ctx, db); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	now := time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC)
	store, err := NewStore(db, WithOperationTimeout(5*time.Second), WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	if err := store.AppendTaskHistory(platformstore.TaskHistoryRecord{
		TaskID:        "task-1",
		WorkerID:      "worker-1",
		TaskType:      "simulation",
		Phase:         domain.TaskPhaseSucceeded,
		PhaseName:     "finished",
		Status:        "finished",
		LastUpdatedAt: now,
		Metadata:      map[string]string{"scene": "A"},
	}); err != nil {
		t.Fatalf("append task history: %v", err)
	}
	latest, ok, err := store.LatestTask("task-1")
	if err != nil {
		t.Fatalf("latest task: %v", err)
	}
	if !ok || latest.Status != "finished" || latest.Metadata["scene"] != "A" {
		t.Fatalf("latest task = %#v ok=%v", latest, ok)
	}

	if err := store.AppendWorkerEvent(domain.DomainEvent{
		Type:       domain.EventWorkerOffline,
		WorkerID:   "worker-1",
		OccurredAt: now,
		Reason:     "no_heartbeat",
		Attributes: map[string]string{"source": "test"},
	}); err != nil {
		t.Fatalf("append worker event: %v", err)
	}
	events, err := store.ListWorkerEvents("worker-1")
	if err != nil {
		t.Fatalf("list worker events: %v", err)
	}
	if len(events) != 1 || events[0].Attributes["source"] != "test" {
		t.Fatalf("unexpected events %#v", events)
	}

	if err := store.AppendAlert(platformstore.AlertRecord{
		DedupeKey:   "worker_offline:worker-1",
		AlertType:   domain.AlertWorkerOffline,
		WorkerID:    "worker-1",
		Status:      domain.AlertStatusFiring,
		Severity:    "critical",
		Message:     "worker offline",
		TriggeredAt: now,
	}); err != nil {
		t.Fatalf("append firing alert: %v", err)
	}
	if err := store.AppendAlert(platformstore.AlertRecord{
		DedupeKey:   "worker_offline:worker-1",
		AlertType:   domain.AlertWorkerOffline,
		WorkerID:    "worker-1",
		Status:      domain.AlertStatusResolved,
		Severity:    "critical",
		Message:     "worker recovered",
		TriggeredAt: now,
		ResolvedAt:  now.Add(time.Minute),
	}); err != nil {
		t.Fatalf("append resolved alert: %v", err)
	}
	alerts, err := store.ListAlerts(platformstore.AlertFilter{
		WorkerID:  "worker-1",
		AlertType: domain.AlertWorkerOffline,
		Status:    domain.AlertStatusResolved,
	})
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(alerts) != 1 || alerts[0].Status != domain.AlertStatusResolved {
		t.Fatalf("unexpected resolved alerts %#v", alerts)
	}
}
