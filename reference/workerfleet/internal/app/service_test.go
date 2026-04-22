package app

import (
	"context"
	"testing"
	"time"

	"workerfleet/internal/domain"
	"workerfleet/internal/handler"
	platformstore "workerfleet/internal/platform/store"
	"workerfleet/internal/platform/store/memory"
)

func TestServiceListWorkersRespectsFilters(t *testing.T) {
	store := memory.NewStore()
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
	store := memory.NewStore()
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
	store := memory.NewStore()
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
	store := memory.NewStore()
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

func TestServiceCaseTimelineReturnsStoredStepHistory(t *testing.T) {
	store := memory.NewStore()
	service := NewService(nil, store)
	now := time.Date(2026, 4, 19, 17, 20, 0, 0, time.UTC)

	if err := store.AppendCaseStepHistory(platformstore.CaseStepHistoryRecord{
		TaskID:     "case-1",
		WorkerID:   "worker-1",
		ExecPlanID: "plan-1",
		NodeName:   "node-a",
		PodName:    "pod-a",
		Step:       "download_bundle",
		StepName:   "download bundle",
		Status:     domain.CaseStepStatusSucceeded,
		Result:     "succeeded",
		Attempt:    1,
		StartedAt:  now.Add(-time.Minute),
		FinishedAt: now,
		ObservedAt: now,
		EventType:  domain.EventTaskStepFinished,
	}); err != nil {
		t.Fatalf("append case step history: %v", err)
	}

	timeline, err := service.CaseTimeline(context.Background(), "case-1")
	if err != nil {
		t.Fatalf("case timeline: %v", err)
	}
	if timeline.TaskID != "case-1" {
		t.Fatalf("task_id = %q, want case-1", timeline.TaskID)
	}
	if len(timeline.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(timeline.Items))
	}
	if timeline.Items[0].Step != "download_bundle" || timeline.Items[0].Result != "succeeded" {
		t.Fatalf("unexpected timeline item %#v", timeline.Items[0])
	}
}

func TestServiceExecPlanCaseDrilldownFiltersAndPaginates(t *testing.T) {
	store := memory.NewStore()
	service := NewService(nil, store)
	now := time.Date(2026, 4, 19, 17, 25, 0, 0, time.UTC)

	records := []platformstore.CaseStepHistoryRecord{
		{TaskID: "case-1", WorkerID: "worker-1", ExecPlanID: "plan-1", NodeName: "node-a", PodName: "pod-a", Step: "simulate", ObservedAt: now},
		{TaskID: "case-2", WorkerID: "worker-2", ExecPlanID: "plan-1", NodeName: "node-a", PodName: "pod-b", Step: "simulate", ObservedAt: now.Add(time.Second)},
		{TaskID: "case-3", WorkerID: "worker-3", ExecPlanID: "plan-1", NodeName: "node-b", PodName: "pod-c", Step: "simulate", ObservedAt: now.Add(2 * time.Second)},
	}
	for _, record := range records {
		if err := store.AppendCaseStepHistory(record); err != nil {
			t.Fatalf("append case step history: %v", err)
		}
	}

	result, err := service.ExecPlanCaseDrilldown(context.Background(), ExecPlanCaseDrilldownQuery{
		ExecPlanID: "plan-1",
		NodeName:   "node-a",
		Step:       "simulate",
		Page:       2,
		PageSize:   1,
	})
	if err != nil {
		t.Fatalf("exec plan drilldown: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("total = %d, want 2", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(result.Items))
	}
	if result.Items[0].TaskID != "case-2" {
		t.Fatalf("task_id = %q, want case-2", result.Items[0].TaskID)
	}
}

func seedWorkerSnapshot(t *testing.T, store *memory.Store, snapshot domain.WorkerSnapshot) {
	t.Helper()
	if err := store.UpsertWorkerSnapshot(snapshot); err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}
}
