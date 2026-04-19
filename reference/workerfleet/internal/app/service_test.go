package app

import (
	"context"
	"testing"
	"time"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
	"github.com/spcent/plumego/reference/workerfleet/internal/handler"
	platformstore "github.com/spcent/plumego/reference/workerfleet/internal/platform/store"
)

func TestServiceListWorkersRespectsFilters(t *testing.T) {
	store := platformstore.NewMemoryStore()
	service := NewService(nil, store)
	now := time.Date(2026, 4, 19, 17, 0, 0, 0, time.UTC)

	seedWorkerSnapshot(t, store, domain.WorkerSnapshot{
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
	})
	seedWorkerSnapshot(t, store, domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{
			WorkerID:  "worker-2",
			Namespace: "sim",
			NodeName:  "node-b",
		},
		Runtime: domain.WorkerRuntime{
			ProcessAlive:   true,
			AcceptingTasks: true,
			LastSeenAt:     now,
		},
		Status: domain.WorkerStatusOnline,
	})
	seedWorkerSnapshot(t, store, domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{
			WorkerID:  "worker-3",
			Namespace: "batch",
			NodeName:  "node-a",
		},
		Runtime: domain.WorkerRuntime{
			ProcessAlive:   true,
			AcceptingTasks: true,
			LastSeenAt:     now,
		},
		Status: domain.WorkerStatusDegraded,
	})

	accepting := false
	result, err := service.ListWorkers(context.Background(), handler.WorkerListQuery{
		Status:         domain.WorkerStatusOnline,
		Namespace:      "sim",
		NodeName:       "node-a",
		TaskType:       "simulation",
		AcceptingTasks: &accepting,
		Page:           1,
		PageSize:       10,
	})
	if err != nil {
		t.Fatalf("list workers: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("total = %d, want 1", result.Total)
	}
	if len(result.Items) != 1 || result.Items[0].WorkerID != "worker-1" {
		t.Fatalf("unexpected items %#v", result.Items)
	}
}

func TestServiceListWorkersPagination(t *testing.T) {
	store := platformstore.NewMemoryStore()
	service := NewService(nil, store)

	for _, workerID := range []domain.WorkerID{"worker-1", "worker-2", "worker-3"} {
		seedWorkerSnapshot(t, store, domain.WorkerSnapshot{
			Identity: domain.WorkerIdentity{WorkerID: workerID},
			Status:   domain.WorkerStatusOnline,
		})
	}

	result, err := service.ListWorkers(context.Background(), handler.WorkerListQuery{
		Page:     2,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("list workers: %v", err)
	}
	if result.Total != 3 {
		t.Fatalf("total = %d, want 3", result.Total)
	}
	if len(result.Items) != 1 || result.Items[0].WorkerID != "worker-2" {
		t.Fatalf("unexpected page items %#v", result.Items)
	}
}

func TestServiceListAlertsEmptyState(t *testing.T) {
	store := platformstore.NewMemoryStore()
	service := NewService(nil, store)

	result, err := service.ListAlerts(context.Background(), handler.AlertListQuery{
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if result.Total != 0 || len(result.Items) != 0 {
		t.Fatalf("unexpected empty state %#v", result)
	}
}

func TestServiceGetTaskFallsBackToHistory(t *testing.T) {
	store := platformstore.NewMemoryStore()
	service := NewService(nil, store)
	now := time.Date(2026, 4, 19, 17, 10, 0, 0, time.UTC)

	if err := store.AppendTaskHistory(domain.TaskHistoryRecord{
		TaskID:        "task-1",
		WorkerID:      "worker-1",
		TaskType:      "simulation",
		Phase:         domain.TaskPhaseSucceeded,
		PhaseName:     "finished",
		Status:        "finished",
		StartedAt:     now.Add(-5 * time.Minute),
		EndedAt:       now,
		LastUpdatedAt: now,
	}); err != nil {
		t.Fatalf("append task history: %v", err)
	}

	task, err := service.GetTask(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if task.Status != "finished" {
		t.Fatalf("status = %q, want finished", task.Status)
	}
	if task.TaskID != "task-1" {
		t.Fatalf("task_id = %q, want task-1", task.TaskID)
	}
}

func seedWorkerSnapshot(t *testing.T, store *platformstore.MemoryStore, snapshot domain.WorkerSnapshot) {
	t.Helper()
	if err := store.UpsertWorkerSnapshot(snapshot); err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}
}
