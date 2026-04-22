package mongo

import (
	"testing"
	"time"

	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func TestWorkerSnapshotDocFromDomain(t *testing.T) {
	now := time.Date(2026, 4, 19, 20, 0, 0, 0, time.UTC)
	doc := WorkerSnapshotDocFromDomain(domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{
			WorkerID:      "worker-1",
			Namespace:     "sim",
			PodName:       "worker-1",
			ContainerName: "worker",
		},
		Runtime: domain.WorkerRuntime{
			ProcessAlive:   true,
			AcceptingTasks: true,
			LastSeenAt:     now,
		},
		Pod: domain.PodSnapshot{
			Phase:        domain.PodPhaseRunning,
			RestartCount: 2,
		},
		Status: domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:     "task-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      domain.TaskPhaseRunning,
			PhaseName:  "running",
			CurrentStep: domain.CaseStepRuntime{
				Step:       "simulate",
				StepName:   "simulation",
				Status:     domain.CaseStepStatusRunning,
				StartedAt:  now.Add(-time.Minute),
				UpdatedAt:  now,
				Attempt:    2,
				ErrorClass: "none",
			},
			Metadata: map[string]string{"scene": "A"},
		}},
		ActiveTaskCount: 1,
	}, now)

	if doc.ID != "worker-1" {
		t.Fatalf("id = %q, want worker-1", doc.ID)
	}
	if doc.SchemaVersion != SchemaVersion {
		t.Fatalf("schema_version = %d, want %d", doc.SchemaVersion, SchemaVersion)
	}
	if len(doc.ActiveTasks) != 1 || doc.ActiveTasks[0].Metadata["scene"] != "A" {
		t.Fatalf("unexpected active tasks %#v", doc.ActiveTasks)
	}
	if doc.ActiveTasks[0].ExecPlanID != "plan-1" || doc.ActiveTasks[0].CurrentStep.Step != "simulate" {
		t.Fatalf("active task step projection lost fields %#v", doc.ActiveTasks[0])
	}

	roundTrip := doc.Domain()
	if roundTrip.Identity.WorkerID != "worker-1" {
		t.Fatalf("round-trip worker_id = %q, want worker-1", roundTrip.Identity.WorkerID)
	}
	if roundTrip.ActiveTaskCount != 1 || roundTrip.ActiveTasks[0].TaskID != "task-1" {
		t.Fatalf("unexpected round-trip active tasks %#v", roundTrip.ActiveTasks)
	}
	if roundTrip.ActiveTasks[0].ExecPlanID != "plan-1" || roundTrip.ActiveTasks[0].CurrentStep.Attempt != 2 {
		t.Fatalf("round-trip active task lost step fields %#v", roundTrip.ActiveTasks[0])
	}
}

func TestHistoryAndAlertDocsCarryExpireAt(t *testing.T) {
	now := time.Date(2026, 4, 19, 20, 5, 0, 0, time.UTC)
	expireAt := now.Add(7 * 24 * time.Hour)

	taskDoc := TaskHistoryDocFromRecord(platformstore.TaskHistoryRecord{
		TaskID:     "task-1",
		WorkerID:   "worker-1",
		ExecPlanID: "plan-1",
		CurrentStep: domain.CaseStepRuntime{
			Step:   "cleanup_env",
			Status: domain.CaseStepStatusSucceeded,
		},
		Status:        "finished",
		LastUpdatedAt: now,
	}, expireAt)
	if !taskDoc.ExpireAt.Equal(expireAt) {
		t.Fatalf("task expire_at = %v, want %v", taskDoc.ExpireAt, expireAt)
	}
	if taskDoc.ExecPlanID != "plan-1" || taskDoc.CurrentStep.Step != "cleanup_env" {
		t.Fatalf("task history lost step fields %#v", taskDoc)
	}
	if taskDoc.Record().CurrentStep.Status != domain.CaseStepStatusSucceeded {
		t.Fatalf("task history round-trip step status = %q", taskDoc.Record().CurrentStep.Status)
	}

	alertDoc := AlertEventDocFromRecord(platformstore.AlertRecord{
		DedupeKey:   "worker_offline:worker-1",
		AlertType:   domain.AlertWorkerOffline,
		WorkerID:    "worker-1",
		Status:      domain.AlertStatusFiring,
		TriggeredAt: now,
	}, expireAt)
	if !alertDoc.ExpireAt.Equal(expireAt) {
		t.Fatalf("alert expire_at = %v, want %v", alertDoc.ExpireAt, expireAt)
	}
	if alertDoc.Record().DedupeKey != "worker_offline:worker-1" {
		t.Fatalf("alert round-trip dedupe key mismatch")
	}
}
