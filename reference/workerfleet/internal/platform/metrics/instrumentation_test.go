package metrics

import (
	"strings"
	"testing"
	"time"

	"workerfleet/internal/domain"
)

func TestObserverRecordsWorkerTaskMetrics(t *testing.T) {
	now := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	collector := NewCollector()
	observer := NewObserver(collector)
	previous := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a"},
		Runtime:  domain.WorkerRuntime{AcceptingTasks: true, LastHeartbeatAt: now.Add(-30 * time.Second)},
		Status:   domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     domain.TaskPhaseRunning,
			PhaseName: "running",
			StartedAt: now.Add(-2 * time.Minute),
			UpdatedAt: now.Add(-30 * time.Second),
		}},
	}
	current := previous
	current.Runtime.LastHeartbeatAt = now
	current.ActiveTasks = []domain.ActiveTask{{
		TaskID:    "task-1",
		TaskType:  "simulation",
		Phase:     domain.TaskPhaseFinalizing,
		PhaseName: "finalizing",
		StartedAt: now.Add(-2 * time.Minute),
		UpdatedAt: now,
	}}

	observer.ObserveWorkerSnapshot(previous, current)
	text := collector.PrometheusText()
	assertMetricsTextContains(t, text,
		`workerfleet_active_cases{namespace="sim",node="node-a",phase="finalizing",task_type="simulation"} 1.000000000`,
		`workerfleet_case_phase_transitions_total{from_phase="running",namespace="sim",node="node-a",task_type="simulation",to_phase="finalizing"} 1.000000000`,
		`workerfleet_case_phase_duration_seconds_count{namespace="sim",node="node-a",phase="running",task_type="simulation"} 1`,
	)
	assertMetricsTextOmits(t, text, "worker_id", "task_id", "pod_name")
}

func TestObserverRecordsFinishedTaskMetrics(t *testing.T) {
	now := time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC)
	collector := NewCollector()
	observer := NewObserver(collector)
	previous := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a"},
		Runtime:  domain.WorkerRuntime{LastHeartbeatAt: now.Add(-1 * time.Minute)},
		Status:   domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     domain.TaskPhaseSucceeded,
			PhaseName: "succeeded",
			StartedAt: now.Add(-5 * time.Minute),
			UpdatedAt: now.Add(-1 * time.Minute),
		}},
	}
	current := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a"},
		Runtime:  domain.WorkerRuntime{LastHeartbeatAt: now},
		Status:   domain.WorkerStatusOnline,
	}

	observer.ObserveWorkerSnapshot(previous, current)
	text := collector.PrometheusText()
	assertMetricsTextContains(t, text,
		`workerfleet_case_finished_total{namespace="sim",node="node-a",status="succeeded",task_type="simulation"} 1.000000000`,
		`workerfleet_case_total_duration_seconds_count{namespace="sim",node="node-a",status="succeeded",task_type="simulation"} 1`,
	)
}

func TestObserverRecordsPodCaseThroughputAndDurationMetrics(t *testing.T) {
	now := time.Date(2026, 4, 20, 11, 10, 0, 0, time.UTC)
	collector := NewCollector()
	observer := NewObserver(collector)
	previous := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a", PodName: "pod-a"},
		Runtime:  domain.WorkerRuntime{LastHeartbeatAt: now.Add(-1 * time.Minute)},
		Status:   domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:     "case-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      domain.TaskPhaseFailed,
			PhaseName:  "failed",
			CurrentStep: domain.CaseStepRuntime{
				Step:       "simulate",
				Status:     domain.CaseStepStatusFailed,
				StartedAt:  now.Add(-3 * time.Minute),
				FinishedAt: now.Add(-1 * time.Minute),
				ErrorClass: "simulation_timeout",
			},
			StartedAt: now.Add(-10 * time.Minute),
			UpdatedAt: now.Add(-1 * time.Minute),
		}},
	}
	current := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a", PodName: "pod-a"},
		Runtime:  domain.WorkerRuntime{LastHeartbeatAt: now},
		Status:   domain.WorkerStatusOnline,
	}

	observer.ObserveWorkerSnapshot(previous, current)
	text := collector.PrometheusText()
	assertMetricsTextContains(t, text,
		`workerfleet_case_completed_total{exec_plan_id="plan-1",namespace="sim",node="node-a",pod="pod-a",result="failed",task_type="simulation"} 1.000000000`,
		`workerfleet_case_failed_total{error_class="simulation_timeout",exec_plan_id="plan-1",namespace="sim",node="node-a",pod="pod-a",task_type="simulation"} 1.000000000`,
		`workerfleet_case_duration_seconds_count{exec_plan_id="plan-1",namespace="sim",node="node-a",pod="pod-a",result="failed",task_type="simulation"} 1`,
	)
}

func TestObserverRecordsStepCompletionAndDurationMetrics(t *testing.T) {
	now := time.Date(2026, 4, 20, 11, 20, 0, 0, time.UTC)
	collector := NewCollector()
	observer := NewObserver(collector)
	previous := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a", PodName: "pod-a"},
		Runtime:  domain.WorkerRuntime{LastHeartbeatAt: now.Add(-2 * time.Minute)},
		Status:   domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:     "case-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      domain.TaskPhaseRunning,
			PhaseName:  "running",
			CurrentStep: domain.CaseStepRuntime{
				Step:      "download_bundle",
				Status:    domain.CaseStepStatusRunning,
				StartedAt: now.Add(-5 * time.Minute),
				UpdatedAt: now.Add(-2 * time.Minute),
			},
			StartedAt: now.Add(-6 * time.Minute),
			UpdatedAt: now.Add(-2 * time.Minute),
		}},
	}
	current := previous
	current.Runtime.LastHeartbeatAt = now
	current.ActiveTasks = []domain.ActiveTask{{
		TaskID:     "case-1",
		ExecPlanID: "plan-1",
		TaskType:   "simulation",
		Phase:      domain.TaskPhaseRunning,
		PhaseName:  "running",
		CurrentStep: domain.CaseStepRuntime{
			Step:      "simulate",
			Status:    domain.CaseStepStatusRunning,
			StartedAt: now,
			UpdatedAt: now,
		},
		StartedAt: now.Add(-6 * time.Minute),
		UpdatedAt: now,
	}}

	observer.ObserveWorkerSnapshot(previous, current)
	text := collector.PrometheusText()
	assertMetricsTextContains(t, text,
		`workerfleet_case_step_completed_total{exec_plan_id="plan-1",namespace="sim",node="node-a",pod="pod-a",result="transitioned",step="download_bundle",task_type="simulation"} 1.000000000`,
		`workerfleet_case_step_duration_seconds_count{exec_plan_id="plan-1",namespace="sim",node="node-a",pod="pod-a",result="transitioned",step="download_bundle",task_type="simulation"} 1`,
	)
}

func TestObserverRecordsActiveStepStuckAndOldestAgeMetrics(t *testing.T) {
	now := time.Date(2026, 4, 20, 11, 30, 0, 0, time.UTC)
	collector := NewCollector()
	observer := NewObserver(collector)
	snapshot := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a", PodName: "pod-a"},
		Runtime:  domain.WorkerRuntime{LastHeartbeatAt: now},
		Status:   domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:     "case-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      domain.TaskPhaseRunning,
			PhaseName:  "running",
			CurrentStep: domain.CaseStepRuntime{
				Step:      "simulate",
				Status:    domain.CaseStepStatusRunning,
				StartedAt: now.Add(-45 * time.Minute),
				UpdatedAt: now,
			},
			StartedAt: now.Add(-50 * time.Minute),
			UpdatedAt: now,
		}},
	}

	observer.ObserveWorkerSnapshot(domain.WorkerSnapshot{}, snapshot)
	text := collector.PrometheusText()
	assertMetricsTextContains(t, text,
		`workerfleet_worker_active_cases{namespace="sim",node="node-a",pod="pod-a"} 1.000000000`,
		`workerfleet_case_step_stuck_cases{exec_plan_id="plan-1",namespace="sim",node="node-a",pod="pod-a",severity="stuck",step="simulate",task_type="simulation"} 1.000000000`,
		`workerfleet_case_step_oldest_active_age_seconds{exec_plan_id="plan-1",namespace="sim",node="node-a",pod="pod-a",step="simulate",task_type="simulation"} 2700.000000000`,
	)
}

func TestObserverBaselinesUnchangedExistingSnapshot(t *testing.T) {
	now := time.Date(2026, 4, 20, 11, 30, 0, 0, time.UTC)
	collector := NewCollector()
	observer := NewObserver(collector)
	snapshot := domain.WorkerSnapshot{
		Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a"},
		Runtime:  domain.WorkerRuntime{AcceptingTasks: true, LastHeartbeatAt: now},
		Status:   domain.WorkerStatusOnline,
		ActiveTasks: []domain.ActiveTask{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     domain.TaskPhaseRunning,
			PhaseName: "running",
			StartedAt: now.Add(-1 * time.Minute),
			UpdatedAt: now,
		}},
	}

	observer.ObserveWorkerSnapshot(snapshot, snapshot)

	text := collector.PrometheusText()
	assertMetricsTextContains(t, text,
		`workerfleet_workers{namespace="sim",node="node-a",status="online"} 1.000000000`,
		`workerfleet_worker_accepting_tasks{namespace="sim",node="node-a"} 1.000000000`,
		`workerfleet_active_cases{namespace="sim",node="node-a",phase="running",task_type="simulation"} 1.000000000`,
	)
}

func TestObserverRecordsAlertMetrics(t *testing.T) {
	collector := NewCollector()
	observer := NewObserver(collector)

	observer.ObserveAlerts([]domain.AlertRecord{
		{AlertType: domain.AlertWorkerOffline, Status: domain.AlertStatusFiring, Severity: "critical"},
		{AlertType: domain.AlertWorkerOffline, Status: domain.AlertStatusResolved, Severity: "critical"},
	})

	text := collector.PrometheusText()
	assertMetricsTextContains(t, text,
		`workerfleet_alerts_total{alert_type="worker_offline",severity="critical",status="firing"} 1.000000000`,
		`workerfleet_alerts_total{alert_type="worker_offline",severity="critical",status="resolved"} 1.000000000`,
		`workerfleet_alerts_firing{alert_type="worker_offline",severity="critical"} 0.000000000`,
	)
}

func TestObserverRecordsInventorySyncMetrics(t *testing.T) {
	collector := NewCollector()
	observer := NewObserver(collector)

	observer.ObserveInventorySync([]domain.WorkerSnapshot{
		{
			Identity: domain.WorkerIdentity{WorkerID: "worker-1", Namespace: "sim", NodeName: "node-a"},
			Pod:      domain.PodSnapshot{Phase: domain.PodPhaseRunning},
		},
		{
			Identity: domain.WorkerIdentity{WorkerID: "worker-2", Namespace: "sim", NodeName: "node-a"},
			Pod:      domain.PodSnapshot{Phase: domain.PodPhasePending},
		},
	}, "sync_once", "success", 25*time.Millisecond)

	text := collector.PrometheusText()
	assertMetricsTextContains(t, text,
		`workerfleet_pods{namespace="sim",node="node-a",phase="running"} 1.000000000`,
		`workerfleet_pods{namespace="sim",node="node-a",phase="pending"} 1.000000000`,
		`workerfleet_kube_inventory_sync_duration_seconds_count{operation="sync_once",result="success"} 1`,
	)
}

func TestObserverNilCollectorIsSafe(t *testing.T) {
	observer := NewObserver(nil)
	observer.ObserveWorkerSnapshot(domain.WorkerSnapshot{}, domain.WorkerSnapshot{})
	observer.ObserveWorkerReportApplied("heartbeat", time.Millisecond)
	observer.ObserveAlerts(nil)
	observer.ObserveInventorySync(nil, "sync_once", "error", time.Millisecond)
}

func assertMetricsTextContains(t *testing.T, text string, wants ...string) {
	t.Helper()

	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Fatalf("metrics output missing %q\n%s", want, text)
		}
	}
}

func assertMetricsTextOmits(t *testing.T, text string, forbidden ...string) {
	t.Helper()

	for _, fragment := range forbidden {
		if strings.Contains(text, fragment) {
			t.Fatalf("metrics output contains forbidden fragment %q\n%s", fragment, text)
		}
	}
}
