package domain

import (
	"testing"
	"time"
)

func TestReconcileActiveTasks(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 5, 0, 0, time.UTC)
	workerID := WorkerID("worker-1")

	t.Run("adds tasks and emits started event", func(t *testing.T) {
		tasks, events := ReconcileActiveTasks(nil, []TaskReport{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     TaskPhasePreparing,
			PhaseName: "warming",
		}}, now, workerID)

		if len(tasks) != 1 {
			t.Fatalf("len(tasks) = %d, want 1", len(tasks))
		}
		if tasks[0].StartedAt.IsZero() || tasks[0].UpdatedAt.IsZero() {
			t.Fatalf("expected normalized timestamps to be set")
		}
		assertEventTypes(t, events, EventTaskStarted)
	})

	t.Run("phase change emits phase changed event", func(t *testing.T) {
		previous := []ActiveTask{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     TaskPhasePreparing,
			PhaseName: "warming",
			StartedAt: now.Add(-1 * time.Minute),
			UpdatedAt: now.Add(-30 * time.Second),
		}}
		next := []TaskReport{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     TaskPhaseRunning,
			PhaseName: "running",
			StartedAt: now.Add(-1 * time.Minute),
			UpdatedAt: now,
		}}

		tasks, events := ReconcileActiveTasks(previous, next, now, workerID)
		if len(tasks) != 1 {
			t.Fatalf("len(tasks) = %d, want 1", len(tasks))
		}
		if tasks[0].Phase != TaskPhaseRunning {
			t.Fatalf("phase = %q, want %q", tasks[0].Phase, TaskPhaseRunning)
		}
		assertEventTypes(t, events, EventTaskPhaseChanged)
	})

	t.Run("step transition emits deterministic step events", func(t *testing.T) {
		previous := []ActiveTask{{
			TaskID:     "task-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      TaskPhaseRunning,
			PhaseName:  "running",
			CurrentStep: CaseStepRuntime{
				Step:      "download_bundle",
				StepName:  "download bundle",
				Status:    CaseStepStatusRunning,
				StartedAt: now.Add(-2 * time.Minute),
				UpdatedAt: now.Add(-1 * time.Minute),
				Attempt:   1,
			},
			StartedAt: now.Add(-3 * time.Minute),
			UpdatedAt: now.Add(-1 * time.Minute),
		}}
		next := []TaskReport{{
			TaskID:     "task-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      TaskPhaseRunning,
			PhaseName:  "running",
			CurrentStep: CaseStepRuntime{
				Step:      "simulate",
				StepName:  "simulation",
				Status:    CaseStepStatusRunning,
				StartedAt: now,
				UpdatedAt: now,
				Attempt:   1,
			},
			StartedAt: now.Add(-3 * time.Minute),
			UpdatedAt: now,
		}}

		tasks, events := ReconcileActiveTasks(previous, next, now, workerID)
		if len(tasks) != 1 {
			t.Fatalf("len(tasks) = %d, want 1", len(tasks))
		}
		if tasks[0].ExecPlanID != "plan-1" {
			t.Fatalf("exec_plan_id = %q, want plan-1", tasks[0].ExecPlanID)
		}
		if tasks[0].CurrentStep.Step != "simulate" {
			t.Fatalf("current step = %q, want simulate", tasks[0].CurrentStep.Step)
		}
		assertEventTypes(t, events, EventTaskStepFinished, EventTaskStepChanged)
		finished := findEvent(t, events, EventTaskStepFinished)
		if finished.Attributes["step"] != "download_bundle" {
			t.Fatalf("finished step = %q, want download_bundle", finished.Attributes["step"])
		}
		changed := findEvent(t, events, EventTaskStepChanged)
		if changed.Attributes["to_step"] != "simulate" {
			t.Fatalf("to_step = %q, want simulate", changed.Attributes["to_step"])
		}
	})

	t.Run("terminal step status emits completion event", func(t *testing.T) {
		previous := []ActiveTask{{
			TaskID:     "task-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      TaskPhaseRunning,
			PhaseName:  "running",
			CurrentStep: CaseStepRuntime{
				Step:      "upload_result",
				StepName:  "upload result",
				Status:    CaseStepStatusRunning,
				StartedAt: now.Add(-2 * time.Minute),
				UpdatedAt: now.Add(-1 * time.Minute),
				Attempt:   1,
			},
			StartedAt: now.Add(-3 * time.Minute),
			UpdatedAt: now.Add(-1 * time.Minute),
		}}
		next := []TaskReport{{
			TaskID:     "task-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      TaskPhaseRunning,
			PhaseName:  "running",
			CurrentStep: CaseStepRuntime{
				Step:       "upload_result",
				Status:     CaseStepStatusFailed,
				UpdatedAt:  now,
				ErrorClass: "object_store_timeout",
			},
			StartedAt: now.Add(-3 * time.Minute),
			UpdatedAt: now,
		}}

		tasks, events := ReconcileActiveTasks(previous, next, now, workerID)
		if len(tasks) != 1 {
			t.Fatalf("len(tasks) = %d, want 1", len(tasks))
		}
		if tasks[0].CurrentStep.FinishedAt.IsZero() {
			t.Fatalf("expected finished_at to be set for terminal step")
		}
		assertEventTypes(t, events, EventTaskStepFinished)
		finished := findEvent(t, events, EventTaskStepFinished)
		if finished.Attributes["error_class"] != "object_store_timeout" {
			t.Fatalf("error_class = %q, want object_store_timeout", finished.Attributes["error_class"])
		}
	})

	t.Run("omitted step preserves previous step without emitting step events", func(t *testing.T) {
		previous := []ActiveTask{{
			TaskID:     "task-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      TaskPhaseRunning,
			PhaseName:  "running",
			CurrentStep: CaseStepRuntime{
				Step:     "simulate",
				StepName: "simulation",
				Status:   CaseStepStatusRunning,
				Attempt:  1,
			},
			StartedAt: now.Add(-2 * time.Minute),
			UpdatedAt: now.Add(-1 * time.Minute),
		}}
		next := []TaskReport{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     TaskPhaseRunning,
			PhaseName: "running",
			StartedAt: now.Add(-2 * time.Minute),
			UpdatedAt: now,
		}}

		tasks, events := ReconcileActiveTasks(previous, next, now, workerID)
		if len(tasks) != 1 {
			t.Fatalf("len(tasks) = %d, want 1", len(tasks))
		}
		if tasks[0].CurrentStep.Step != "simulate" {
			t.Fatalf("current step = %q, want simulate", tasks[0].CurrentStep.Step)
		}
		if hasEvent(events, EventTaskStepChanged) || hasEvent(events, EventTaskStepFinished) {
			t.Fatalf("unexpected step events %#v", events)
		}
	})

	t.Run("task removal emits finished event", func(t *testing.T) {
		previous := []ActiveTask{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     TaskPhaseRunning,
			PhaseName: "running",
			StartedAt: now.Add(-2 * time.Minute),
			UpdatedAt: now.Add(-30 * time.Second),
		}}

		tasks, events := ReconcileActiveTasks(previous, nil, now, workerID)
		if len(tasks) != 0 {
			t.Fatalf("len(tasks) = %d, want 0", len(tasks))
		}
		assertEventTypes(t, events, EventTaskFinished)
	})

	t.Run("duplicate task IDs emit conflict", func(t *testing.T) {
		_, events := ReconcileActiveTasks(nil, []TaskReport{
			{TaskID: "task-1", TaskType: "simulation", Phase: TaskPhasePreparing},
			{TaskID: "task-1", TaskType: "simulation", Phase: TaskPhaseRunning},
		}, now, workerID)

		assertEventTypes(t, events, EventStateConflictDetected, EventTaskStarted)
	})
}

func TestMergeWorkerReport(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 10, 0, 0, time.UTC)
	policy := DefaultStatusPolicy()

	merged, events := MergeWorkerReport(WorkerSnapshot{}, WorkerReport{
		Identity: WorkerIdentity{
			WorkerID:      "worker-1",
			Namespace:     "sim",
			PodName:       "worker-1",
			ContainerName: "worker",
		},
		ProcessAlive:   true,
		AcceptingTasks: true,
		ObservedAt:     now,
		ActiveTasks: []TaskReport{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     TaskPhaseRunning,
			PhaseName: "running",
		}},
	}, now, policy)

	if merged.Status != WorkerStatusOnline {
		t.Fatalf("status = %q, want %q", merged.Status, WorkerStatusOnline)
	}
	if merged.ActiveTaskCount != 1 {
		t.Fatalf("active task count = %d, want 1", merged.ActiveTaskCount)
	}
	assertEventTypes(t, events, EventWorkerRegistered, EventWorkerHeartbeat, EventTaskStarted)
}

func TestMergePodSnapshot(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 15, 0, 0, time.UTC)
	policy := DefaultStatusPolicy()
	snapshot := WorkerSnapshot{
		Identity: WorkerIdentity{WorkerID: "worker-1"},
		Runtime: WorkerRuntime{
			ProcessAlive:    true,
			AcceptingTasks:  true,
			LastHeartbeatAt: now.Add(-5 * time.Second),
			LastSeenAt:      now.Add(-5 * time.Second),
		},
		Pod: PodSnapshot{
			Phase:        PodPhaseRunning,
			RestartCount: 1,
		},
		Status: WorkerStatusOnline,
	}

	merged, events := MergePodSnapshot(snapshot, PodSnapshot{
		Phase:        PodPhaseRunning,
		RestartCount: 2,
	}, now, policy)

	if merged.Pod.RestartCount != 2 {
		t.Fatalf("restart count = %d, want 2", merged.Pod.RestartCount)
	}
	assertEventTypes(t, events, EventPodRestarted)
}

func assertEventTypes(t *testing.T, events []DomainEvent, want ...EventType) {
	t.Helper()

	for _, expected := range want {
		found := false
		for _, event := range events {
			if event.Type == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing event %q in %#v", expected, events)
		}
	}
}

func findEvent(t *testing.T, events []DomainEvent, eventType EventType) DomainEvent {
	t.Helper()

	for _, event := range events {
		if event.Type == eventType {
			return event
		}
	}
	t.Fatalf("missing event %q in %#v", eventType, events)
	return DomainEvent{}
}

func hasEvent(events []DomainEvent, eventType EventType) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}
