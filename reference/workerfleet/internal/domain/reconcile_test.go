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
