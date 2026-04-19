package domain

import (
	"testing"
	"time"
)

func TestEvaluateWorkerStatus(t *testing.T) {
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	policy := StatusPolicy{
		StaleAfter:      45 * time.Second,
		OfflineAfter:    90 * time.Second,
		StageStuckAfter: 5 * time.Minute,
	}

	base := WorkerSnapshot{
		Identity: WorkerIdentity{WorkerID: "worker-1"},
		Runtime: WorkerRuntime{
			ProcessAlive:    true,
			AcceptingTasks:  true,
			LastHeartbeatAt: now.Add(-15 * time.Second),
			LastSeenAt:      now.Add(-15 * time.Second),
		},
		Pod: PodSnapshot{Phase: PodPhaseRunning},
	}

	tests := []struct {
		name   string
		input  WorkerSnapshot
		status WorkerStatus
		reason string
	}{
		{
			name:   "first registration remains unknown without heartbeat",
			input:  WorkerSnapshot{Identity: WorkerIdentity{WorkerID: "worker-1"}},
			status: WorkerStatusUnknown,
			reason: "awaiting_first_heartbeat",
		},
		{
			name:   "fresh heartbeat and accepting tasks is online",
			input:  base,
			status: WorkerStatusOnline,
			reason: "ready",
		},
		{
			name: "busy worker stays online",
			input: WorkerSnapshot{
				Identity: base.Identity,
				Runtime: WorkerRuntime{
					ProcessAlive:    true,
					AcceptingTasks:  false,
					LastHeartbeatAt: now.Add(-20 * time.Second),
					LastSeenAt:      now.Add(-20 * time.Second),
				},
				Pod: base.Pod,
				ActiveTasks: []ActiveTask{{
					TaskID:    "task-1",
					Phase:     TaskPhaseRunning,
					PhaseName: "running",
					StartedAt: now.Add(-2 * time.Minute),
					UpdatedAt: now.Add(-10 * time.Second),
				}},
			},
			status: WorkerStatusOnline,
			reason: "busy",
		},
		{
			name: "not accepting tasks without workload is degraded",
			input: WorkerSnapshot{
				Identity: base.Identity,
				Runtime: WorkerRuntime{
					ProcessAlive:    true,
					AcceptingTasks:  false,
					LastHeartbeatAt: now.Add(-10 * time.Second),
					LastSeenAt:      now.Add(-10 * time.Second),
				},
				Pod: base.Pod,
			},
			status: WorkerStatusDegraded,
			reason: "not_accepting_tasks",
		},
		{
			name: "stale heartbeat is degraded",
			input: WorkerSnapshot{
				Identity: base.Identity,
				Runtime: WorkerRuntime{
					ProcessAlive:    true,
					AcceptingTasks:  true,
					LastHeartbeatAt: now.Add(-60 * time.Second),
					LastSeenAt:      now.Add(-60 * time.Second),
				},
				Pod: base.Pod,
			},
			status: WorkerStatusDegraded,
			reason: "heartbeat_stale",
		},
		{
			name: "expired heartbeat is offline",
			input: WorkerSnapshot{
				Identity: base.Identity,
				Runtime: WorkerRuntime{
					ProcessAlive:    true,
					AcceptingTasks:  true,
					LastHeartbeatAt: now.Add(-2 * time.Minute),
					LastSeenAt:      now.Add(-2 * time.Minute),
				},
				Pod: base.Pod,
			},
			status: WorkerStatusOffline,
			reason: "heartbeat_expired",
		},
		{
			name: "failed pod is offline",
			input: WorkerSnapshot{
				Identity: base.Identity,
				Runtime:  base.Runtime,
				Pod:      PodSnapshot{Phase: PodPhaseFailed},
			},
			status: WorkerStatusOffline,
			reason: "pod_failed",
		},
		{
			name: "worker reported error is degraded",
			input: WorkerSnapshot{
				Identity: base.Identity,
				Runtime: WorkerRuntime{
					ProcessAlive:    true,
					AcceptingTasks:  true,
					LastHeartbeatAt: now.Add(-5 * time.Second),
					LastSeenAt:      now.Add(-5 * time.Second),
					LastError:       "disk pressure",
				},
				Pod: base.Pod,
			},
			status: WorkerStatusDegraded,
			reason: "worker_reported_error",
		},
		{
			name: "stuck task is degraded",
			input: WorkerSnapshot{
				Identity: base.Identity,
				Runtime: WorkerRuntime{
					ProcessAlive:    true,
					AcceptingTasks:  false,
					LastHeartbeatAt: now.Add(-5 * time.Second),
					LastSeenAt:      now.Add(-5 * time.Second),
				},
				Pod: base.Pod,
				ActiveTasks: []ActiveTask{{
					TaskID:    "task-1",
					Phase:     TaskPhaseRunning,
					PhaseName: "running",
					StartedAt: now.Add(-10 * time.Minute),
					UpdatedAt: now.Add(-6 * time.Minute),
				}},
			},
			status: WorkerStatusDegraded,
			reason: "task_stage_stuck",
		},
		{
			name: "process not alive is offline",
			input: WorkerSnapshot{
				Identity: base.Identity,
				Runtime: WorkerRuntime{
					ProcessAlive:    false,
					AcceptingTasks:  false,
					LastHeartbeatAt: now.Add(-1 * time.Second),
					LastSeenAt:      now.Add(-1 * time.Second),
				},
				Pod: base.Pod,
			},
			status: WorkerStatusOffline,
			reason: "process_not_alive",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status, reason := EvaluateWorkerStatus(test.input, now, policy)
			if status != test.status {
				t.Fatalf("status = %q, want %q", status, test.status)
			}
			if reason != test.reason {
				t.Fatalf("reason = %q, want %q", reason, test.reason)
			}
		})
	}
}
