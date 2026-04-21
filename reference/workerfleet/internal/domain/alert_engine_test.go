package domain

import (
	"testing"
	"time"
)

type snapshotListStub struct {
	snapshots []WorkerSnapshot
}

func (s snapshotListStub) ListCurrentWorkerSnapshots() ([]WorkerSnapshot, error) {
	return s.snapshots, nil
}

type alertStoreStub struct {
	records []AlertRecord
}

func (s *alertStoreStub) ListAlertRecords() ([]AlertRecord, error) {
	out := make([]AlertRecord, len(s.records))
	copy(out, s.records)
	return out, nil
}

func (s *alertStoreStub) AppendAlert(record AlertRecord) error {
	s.records = append(s.records, record)
	return nil
}

type alertMetricsObserver struct {
	records []AlertRecord
}

func (o *alertMetricsObserver) ObserveAlerts(records []AlertRecord) {
	o.records = append(o.records, records...)
}

func TestAlertEngineFiresOfflineAlert(t *testing.T) {
	now := time.Date(2026, 4, 19, 15, 0, 0, 0, time.UTC)
	engine := NewAlertEngine(
		snapshotListStub{snapshots: []WorkerSnapshot{{
			Identity: WorkerIdentity{WorkerID: "worker-1"},
			Status:   WorkerStatusOffline,
		}}},
		&alertStoreStub{},
		DefaultStatusPolicy(),
		fixedClock{now: now},
	)

	emitted, err := engine.Evaluate()
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(emitted) != 1 {
		t.Fatalf("len(emitted) = %d, want 1", len(emitted))
	}
	if emitted[0].AlertType != AlertWorkerOffline {
		t.Fatalf("alert_type = %q, want %q", emitted[0].AlertType, AlertWorkerOffline)
	}
}

func TestAlertEngineDedupesExistingOpenAlert(t *testing.T) {
	now := time.Date(2026, 4, 19, 15, 5, 0, 0, time.UTC)
	store := &alertStoreStub{records: []AlertRecord{{
		AlertID:     "existing",
		WorkerID:    "worker-1",
		AlertType:   AlertWorkerOffline,
		Status:      AlertStatusFiring,
		DedupeKey:   dedupeKey(AlertWorkerOffline, "worker-1", ""),
		Message:     "worker is offline",
		TriggeredAt: now.Add(-1 * time.Minute),
	}}}
	engine := NewAlertEngine(
		snapshotListStub{snapshots: []WorkerSnapshot{{
			Identity: WorkerIdentity{WorkerID: "worker-1"},
			Status:   WorkerStatusOffline,
		}}},
		store,
		DefaultStatusPolicy(),
		fixedClock{now: now},
	)

	emitted, err := engine.Evaluate()
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(emitted) != 0 {
		t.Fatalf("expected no newly emitted alerts, got %#v", emitted)
	}
}

func TestAlertEngineResolvesRecoveredAlert(t *testing.T) {
	now := time.Date(2026, 4, 19, 15, 10, 0, 0, time.UTC)
	store := &alertStoreStub{records: []AlertRecord{{
		AlertID:     "existing",
		WorkerID:    "worker-1",
		AlertType:   AlertWorkerOffline,
		Status:      AlertStatusFiring,
		DedupeKey:   dedupeKey(AlertWorkerOffline, "worker-1", ""),
		Message:     "worker is offline",
		TriggeredAt: now.Add(-2 * time.Minute),
	}}}
	engine := NewAlertEngine(
		snapshotListStub{snapshots: []WorkerSnapshot{{
			Identity: WorkerIdentity{WorkerID: "worker-1"},
			Status:   WorkerStatusOnline,
		}}},
		store,
		DefaultStatusPolicy(),
		fixedClock{now: now},
	)

	emitted, err := engine.Evaluate()
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(emitted) != 1 {
		t.Fatalf("len(emitted) = %d, want 1", len(emitted))
	}
	if emitted[0].Status != AlertStatusResolved {
		t.Fatalf("status = %q, want %q", emitted[0].Status, AlertStatusResolved)
	}
}

func TestAlertEngineDetectsTaskConflict(t *testing.T) {
	now := time.Date(2026, 4, 19, 15, 15, 0, 0, time.UTC)
	engine := NewAlertEngine(
		snapshotListStub{snapshots: []WorkerSnapshot{
			{
				Identity: WorkerIdentity{WorkerID: "worker-1"},
				Status:   WorkerStatusOnline,
				ActiveTasks: []ActiveTask{{
					TaskID:    "task-1",
					TaskType:  "simulation",
					Phase:     TaskPhaseRunning,
					PhaseName: "running",
				}},
			},
			{
				Identity: WorkerIdentity{WorkerID: "worker-2"},
				Status:   WorkerStatusOnline,
				ActiveTasks: []ActiveTask{{
					TaskID:    "task-1",
					TaskType:  "simulation",
					Phase:     TaskPhaseRunning,
					PhaseName: "running",
				}},
			},
		}},
		&alertStoreStub{},
		DefaultStatusPolicy(),
		fixedClock{now: now},
	)

	emitted, err := engine.Evaluate()
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(emitted) != 1 {
		t.Fatalf("len(emitted) = %d, want 1", len(emitted))
	}
	if emitted[0].AlertType != AlertTaskConflict {
		t.Fatalf("alert_type = %q, want %q", emitted[0].AlertType, AlertTaskConflict)
	}
}

func TestAlertEngineCallsMetricsObserver(t *testing.T) {
	now := time.Date(2026, 4, 19, 15, 20, 0, 0, time.UTC)
	observer := &alertMetricsObserver{}
	engine := NewAlertEngine(
		snapshotListStub{snapshots: []WorkerSnapshot{{
			Identity: WorkerIdentity{WorkerID: "worker-1"},
			Status:   WorkerStatusOffline,
		}}},
		&alertStoreStub{},
		DefaultStatusPolicy(),
		fixedClock{now: now},
		WithAlertMetrics(observer),
	)

	if _, err := engine.Evaluate(); err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(observer.records) != 1 {
		t.Fatalf("observer records = %d, want 1", len(observer.records))
	}
}
