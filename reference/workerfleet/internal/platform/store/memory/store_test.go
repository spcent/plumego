package memory

import (
	"testing"
	"time"

	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func TestStoreUpsertWorkerSnapshotCopiesState(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	snapshot := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{
			WorkerID:  "worker-1",
			Namespace: "sim",
			NodeName:  "node-a",
		},
		Runtime: domain.WorkerRuntime{
			ProcessAlive:    true,
			AcceptingTasks:  true,
			LastHeartbeatAt: now,
			LastSeenAt:      now,
		},
		Status: domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     domain.TaskPhaseRunning,
			PhaseName: "running",
			UpdatedAt: now,
			Metadata: map[string]string{
				"job": "A",
			},
		}},
	}

	if err := store.UpsertWorkerSnapshot(snapshot); err != nil {
		t.Fatalf("upsert snapshot: %v", err)
	}

	snapshot.ActiveTasks[0].Metadata["job"] = "mutated"
	got, ok, err := store.GetWorkerSnapshot("worker-1")
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if !ok {
		t.Fatalf("expected snapshot to exist")
	}
	if got.ActiveTasks[0].Metadata["job"] != "A" {
		t.Fatalf("stored snapshot metadata mutated: %#v", got.ActiveTasks[0].Metadata)
	}
}

func TestStoreReplaceActiveTasksRebuildsTaskIndex(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 19, 12, 5, 0, 0, time.UTC)

	if err := store.UpsertWorkerSnapshot(domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1"},
		Status:   domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     domain.TaskPhasePreparing,
			PhaseName: "warming",
			UpdatedAt: now,
		}},
	}); err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}

	if err := store.ReplaceActiveTasks("worker-1", []domain.ActiveTask{{
		TaskID:    "task-2",
		TaskType:  "simulation",
		Phase:     domain.TaskPhaseRunning,
		PhaseName: "running",
		UpdatedAt: now,
	}}); err != nil {
		t.Fatalf("replace tasks: %v", err)
	}

	if _, ok, err := store.GetTask("task-1"); err != nil || ok {
		t.Fatalf("expected old task index to be removed, ok=%v err=%v", ok, err)
	}
	current, ok, err := store.GetTask("task-2")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if !ok {
		t.Fatalf("expected current task to exist")
	}
	if current.WorkerID != "worker-1" {
		t.Fatalf("worker_id = %q, want worker-1", current.WorkerID)
	}
}

func TestStoreRetentionPrunesHistoryAndAlertsOnly(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 19, 12, 10, 0, 0, time.UTC)

	if err := store.UpsertWorkerSnapshot(domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1"},
		Status:   domain.WorkerStatusOnline,
	}); err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}
	if err := store.AppendTaskHistory(platformstore.TaskHistoryRecord{
		TaskID:        "task-old",
		WorkerID:      "worker-1",
		Status:        "finished",
		LastUpdatedAt: now.Add(-(platformstore.DefaultRetention + time.Hour)),
	}); err != nil {
		t.Fatalf("append old task history: %v", err)
	}
	if err := store.AppendTaskHistory(platformstore.TaskHistoryRecord{
		TaskID:        "task-new",
		WorkerID:      "worker-1",
		Status:        "finished",
		LastUpdatedAt: now.Add(-6 * 24 * time.Hour),
	}); err != nil {
		t.Fatalf("append new task history: %v", err)
	}
	if err := store.AppendCaseStepHistory(platformstore.CaseStepHistoryRecord{
		TaskID:     "task-old",
		WorkerID:   "worker-1",
		ExecPlanID: "plan-1",
		Step:       "simulate",
		ObservedAt: now.Add(-(platformstore.DefaultRetention + time.Hour)),
	}); err != nil {
		t.Fatalf("append old case step history: %v", err)
	}
	if err := store.AppendWorkerEvent(domain.DomainEvent{
		Type:       domain.EventWorkerHeartbeat,
		WorkerID:   "worker-1",
		OccurredAt: now.Add(-(platformstore.DefaultRetention + time.Hour)),
	}); err != nil {
		t.Fatalf("append old event: %v", err)
	}
	if err := store.AppendAlert(platformstore.AlertRecord{
		AlertID:     "alert-old",
		WorkerID:    "worker-1",
		AlertType:   domain.AlertWorkerOffline,
		Status:      "resolved",
		TriggeredAt: now.Add(-(platformstore.DefaultRetention + time.Hour)),
	}); err != nil {
		t.Fatalf("append old alert: %v", err)
	}

	result := store.ApplyRetention(now, platformstore.DefaultRetention)
	if result.TaskHistoryPruned != 1 {
		t.Fatalf("task history pruned = %d, want 1", result.TaskHistoryPruned)
	}
	if result.CaseStepHistoryPruned != 1 {
		t.Fatalf("case step history pruned = %d, want 1", result.CaseStepHistoryPruned)
	}
	if result.WorkerEventsPruned != 1 {
		t.Fatalf("worker events pruned = %d, want 1", result.WorkerEventsPruned)
	}
	if result.AlertsPruned != 1 {
		t.Fatalf("alerts pruned = %d, want 1", result.AlertsPruned)
	}

	if _, ok, err := store.GetWorkerSnapshot("worker-1"); err != nil || !ok {
		t.Fatalf("expected current snapshot to be preserved, ok=%v err=%v", ok, err)
	}
	records, err := store.TaskHistory("task-new")
	if err != nil {
		t.Fatalf("task history: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len(task-new history) = %d, want 1", len(records))
	}
	records, err = store.TaskHistory("task-old")
	if err != nil {
		t.Fatalf("task history old: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("len(task-old history) = %d, want 0", len(records))
	}
}

func TestStoreLatestTaskFallsBackToHistory(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 19, 12, 15, 0, 0, time.UTC)

	if err := store.AppendTaskHistory(platformstore.TaskHistoryRecord{
		TaskID:        "task-1",
		WorkerID:      "worker-1",
		TaskType:      "simulation",
		Phase:         domain.TaskPhaseSucceeded,
		PhaseName:     "finished",
		Status:        "finished",
		LastUpdatedAt: now,
	}); err != nil {
		t.Fatalf("append task history: %v", err)
	}

	record, ok, err := store.LatestTask("task-1")
	if err != nil {
		t.Fatalf("latest task: %v", err)
	}
	if !ok {
		t.Fatalf("expected latest task history to exist")
	}
	if record.Status != "finished" {
		t.Fatalf("status = %q, want finished", record.Status)
	}
}

func TestStoreCaseStepHistoryFiltersAndCopiesRecords(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 19, 12, 20, 0, 0, time.UTC)

	if err := store.AppendCaseStepHistory(platformstore.CaseStepHistoryRecord{
		TaskID:     "case-1",
		WorkerID:   "worker-1",
		ExecPlanID: "plan-1",
		NodeName:   "node-a",
		PodName:    "pod-a",
		Step:       "simulate",
		Status:     domain.CaseStepStatusRunning,
		ObservedAt: now,
	}); err != nil {
		t.Fatalf("append case step: %v", err)
	}
	if err := store.AppendCaseStepHistory(platformstore.CaseStepHistoryRecord{
		TaskID:     "case-2",
		WorkerID:   "worker-2",
		ExecPlanID: "plan-1",
		NodeName:   "node-b",
		PodName:    "pod-b",
		Step:       "cleanup_env",
		Status:     domain.CaseStepStatusRunning,
		ObservedAt: now.Add(time.Second),
	}); err != nil {
		t.Fatalf("append second case step: %v", err)
	}

	records, err := store.ListCaseStepHistory(platformstore.CaseStepHistoryFilter{
		ExecPlanID: "plan-1",
		NodeName:   "node-a",
		Step:       "simulate",
	})
	if err != nil {
		t.Fatalf("list case step history: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	if records[0].TaskID != "case-1" {
		t.Fatalf("task_id = %q, want case-1", records[0].TaskID)
	}
	records[0].Step = "mutated"

	timeline, err := store.CaseStepHistory("case-1")
	if err != nil {
		t.Fatalf("case step history: %v", err)
	}
	if timeline[0].Step != "simulate" {
		t.Fatalf("stored step mutated to %q", timeline[0].Step)
	}
}

func TestStoreWorkerStepEventsMaterializeCaseStepHistory(t *testing.T) {
	store := NewStore()
	now := time.Date(2026, 4, 19, 12, 25, 0, 0, time.UTC)

	if err := store.UpsertWorkerSnapshot(domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{
			WorkerID:  "worker-1",
			Namespace: "sim",
			PodName:   "pod-a",
			NodeName:  "node-a",
		},
		ActiveTasks: []domain.ActiveTask{{
			TaskID:     "case-1",
			ExecPlanID: "plan-1",
		}},
	}); err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}

	if err := store.AppendWorkerEvent(domain.DomainEvent{
		Type:       domain.EventTaskStepFinished,
		WorkerID:   "worker-1",
		TaskID:     "case-1",
		OccurredAt: now,
		Attributes: map[string]string{
			"exec_plan_id":  "plan-1",
			"step":          "download_bundle",
			"step_name":     "download bundle",
			"step_status":   "failed",
			"result":        "failed",
			"error_class":   "object_store_timeout",
			"step_attempt":  "2",
			"step_started":  now.Add(-time.Minute).Format(time.RFC3339Nano),
			"step_finished": now.Format(time.RFC3339Nano),
		},
	}); err != nil {
		t.Fatalf("append worker event: %v", err)
	}

	records, err := store.CaseStepHistory("case-1")
	if err != nil {
		t.Fatalf("case step history: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	if records[0].NodeName != "node-a" || records[0].PodName != "pod-a" {
		t.Fatalf("unexpected topology fields %#v", records[0])
	}
	if records[0].ErrorClass != "object_store_timeout" || records[0].Attempt != 2 {
		t.Fatalf("unexpected step fields %#v", records[0])
	}
}
