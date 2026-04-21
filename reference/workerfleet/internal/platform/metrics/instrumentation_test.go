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
	for _, want := range []string{
		`workerfleet_active_cases{namespace="sim",node="node-a",phase="finalizing",task_type="simulation"} 1.000000000`,
		`workerfleet_case_phase_transitions_total{from_phase="running",namespace="sim",node="node-a",task_type="simulation",to_phase="finalizing"} 1.000000000`,
		`workerfleet_case_phase_duration_seconds_count{namespace="sim",node="node-a",phase="running",task_type="simulation"} 1`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("metrics output missing %q\n%s", want, text)
		}
	}
	for _, forbidden := range []string{"worker_id", "task_id", "pod_name"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("metrics output contains forbidden label %q\n%s", forbidden, text)
		}
	}
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
	for _, want := range []string{
		`workerfleet_case_finished_total{namespace="sim",node="node-a",status="succeeded",task_type="simulation"} 1.000000000`,
		`workerfleet_case_total_duration_seconds_count{namespace="sim",node="node-a",status="succeeded",task_type="simulation"} 1`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("metrics output missing %q\n%s", want, text)
		}
	}
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
	for _, want := range []string{
		`workerfleet_workers{namespace="sim",node="node-a",status="online"} 1.000000000`,
		`workerfleet_worker_accepting_tasks{namespace="sim",node="node-a"} 1.000000000`,
		`workerfleet_active_cases{namespace="sim",node="node-a",phase="running",task_type="simulation"} 1.000000000`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("metrics output missing %q\n%s", want, text)
		}
	}
}

func TestObserverRecordsAlertMetrics(t *testing.T) {
	collector := NewCollector()
	observer := NewObserver(collector)

	observer.ObserveAlerts([]domain.AlertRecord{
		{AlertType: domain.AlertWorkerOffline, Status: domain.AlertStatusFiring, Severity: "critical"},
		{AlertType: domain.AlertWorkerOffline, Status: domain.AlertStatusResolved, Severity: "critical"},
	})

	text := collector.PrometheusText()
	for _, want := range []string{
		`workerfleet_alerts_total{alert_type="worker_offline",severity="critical",status="firing"} 1.000000000`,
		`workerfleet_alerts_total{alert_type="worker_offline",severity="critical",status="resolved"} 1.000000000`,
		`workerfleet_alerts_firing{alert_type="worker_offline",severity="critical"} 0.000000000`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("metrics output missing %q\n%s", want, text)
		}
	}
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
	for _, want := range []string{
		`workerfleet_pods{namespace="sim",node="node-a",phase="running"} 1.000000000`,
		`workerfleet_pods{namespace="sim",node="node-a",phase="pending"} 1.000000000`,
		`workerfleet_kube_inventory_sync_duration_seconds_count{operation="sync_once",result="success"} 1`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("metrics output missing %q\n%s", want, text)
		}
	}
}

func TestObserverNilCollectorIsSafe(t *testing.T) {
	observer := NewObserver(nil)
	observer.ObserveWorkerSnapshot(domain.WorkerSnapshot{}, domain.WorkerSnapshot{})
	observer.ObserveWorkerReportApplied("heartbeat", time.Millisecond)
	observer.ObserveAlerts(nil)
	observer.ObserveInventorySync(nil, "sync_once", "error", time.Millisecond)
}
