package domain

import (
	"context"
	"testing"
	"time"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type ingestSnapshotStore struct {
	snapshots map[WorkerID]WorkerSnapshot
}

func newIngestSnapshotStore() *ingestSnapshotStore {
	return &ingestSnapshotStore{snapshots: make(map[WorkerID]WorkerSnapshot)}
}

func (s *ingestSnapshotStore) UpsertWorkerSnapshot(ctx context.Context, snapshot WorkerSnapshot) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.snapshots[snapshot.Identity.WorkerID] = snapshot
	return nil
}

func (s *ingestSnapshotStore) GetWorkerSnapshot(ctx context.Context, workerID WorkerID) (WorkerSnapshot, bool, error) {
	if err := ctx.Err(); err != nil {
		return WorkerSnapshot{}, false, err
	}
	snapshot, ok := s.snapshots[workerID]
	return snapshot, ok, nil
}

type ingestTaskHistoryStore struct {
	records []TaskHistoryRecord
}

func (s *ingestTaskHistoryStore) AppendTaskHistory(ctx context.Context, record TaskHistoryRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.records = append(s.records, record)
	return nil
}

type ingestEventStore struct {
	events []DomainEvent
}

func (s *ingestEventStore) AppendWorkerEvent(ctx context.Context, event DomainEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.events = append(s.events, event)
	return nil
}

type ingestMetricsObserver struct {
	snapshots int
	events    int
	applied   int
}

func (o *ingestMetricsObserver) ObserveWorkerSnapshot(WorkerSnapshot, WorkerSnapshot) {
	o.snapshots++
}

func (o *ingestMetricsObserver) ObserveWorkerEvents(WorkerSnapshot, WorkerSnapshot, []DomainEvent) {
	o.events++
}

func (o *ingestMetricsObserver) ObserveWorkerReportApplied(string, time.Duration) {
	o.applied++
}

func TestIngestServiceRegisterCreatesUnknownSnapshot(t *testing.T) {
	now := time.Date(2026, 4, 19, 13, 0, 0, 0, time.UTC)
	snapshots := newIngestSnapshotStore()
	history := &ingestTaskHistoryStore{}
	events := &ingestEventStore{}
	service := NewIngestService(snapshots, history, events, DefaultStatusPolicy(), fixedClock{now: now})

	snapshot, err := service.Register(context.Background(), RegisterCommand{
		Identity: WorkerIdentity{
			WorkerID:      "worker-1",
			Namespace:     "sim",
			PodName:       "worker-1",
			ContainerName: "worker",
		},
		ObservedAt: now,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if snapshot.Status != WorkerStatusUnknown {
		t.Fatalf("status = %q, want %q", snapshot.Status, WorkerStatusUnknown)
	}
	if len(events.events) != 1 || events.events[0].Type != EventWorkerRegistered {
		t.Fatalf("unexpected events %#v", events.events)
	}
}

func TestIngestServiceRegisterHonorsCanceledContext(t *testing.T) {
	snapshots := newIngestSnapshotStore()
	service := NewIngestService(snapshots, nil, nil, DefaultStatusPolicy(), fixedClock{now: time.Now().UTC()})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := service.Register(ctx, RegisterCommand{Identity: WorkerIdentity{WorkerID: "worker-1"}}); err == nil {
		t.Fatalf("Register succeeded with canceled context")
	}
}

func TestIngestServiceHeartbeatUpdatesSnapshotAndHistory(t *testing.T) {
	now := time.Date(2026, 4, 19, 13, 5, 0, 0, time.UTC)
	snapshots := newIngestSnapshotStore()
	history := &ingestTaskHistoryStore{}
	events := &ingestEventStore{}
	service := NewIngestService(snapshots, history, events, DefaultStatusPolicy(), fixedClock{now: now})

	if _, err := service.Register(context.Background(), RegisterCommand{
		Identity: WorkerIdentity{
			WorkerID:      "worker-1",
			Namespace:     "sim",
			PodName:       "worker-1",
			ContainerName: "worker",
		},
		ObservedAt: now.Add(-1 * time.Minute),
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	snapshot, err := service.Heartbeat(context.Background(), WorkerReport{
		Identity:       WorkerIdentity{WorkerID: "worker-1"},
		ProcessAlive:   true,
		AcceptingTasks: false,
		ObservedAt:     now,
		ActiveTasks: []TaskReport{
			{
				TaskID:    "task-1",
				TaskType:  "simulation",
				Phase:     TaskPhaseRunning,
				PhaseName: "running",
				StartedAt: now.Add(-30 * time.Second),
				UpdatedAt: now,
			},
		},
	})
	if err != nil {
		t.Fatalf("heartbeat: %v", err)
	}

	if snapshot.Status != WorkerStatusOnline {
		t.Fatalf("status = %q, want %q", snapshot.Status, WorkerStatusOnline)
	}
	if snapshot.ActiveTaskCount != 1 {
		t.Fatalf("active task count = %d, want 1", snapshot.ActiveTaskCount)
	}
	if len(history.records) != 1 || history.records[0].TaskID != "task-1" {
		t.Fatalf("unexpected history records %#v", history.records)
	}
	if !hasEventType(events.events, EventWorkerHeartbeat) || !hasEventType(events.events, EventTaskStarted) {
		t.Fatalf("unexpected events %#v", events.events)
	}
}

func TestIngestServiceMetricsObserverIsOptional(t *testing.T) {
	now := time.Date(2026, 4, 19, 13, 7, 0, 0, time.UTC)
	snapshots := newIngestSnapshotStore()
	service := NewIngestService(snapshots, nil, nil, DefaultStatusPolicy(), fixedClock{now: now}, WithIngestMetrics(nil))

	if _, err := service.Register(context.Background(), RegisterCommand{
		Identity:   WorkerIdentity{WorkerID: "worker-1", Namespace: "sim"},
		ObservedAt: now,
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
}

func TestIngestServiceCallsMetricsObserverAfterHeartbeat(t *testing.T) {
	now := time.Date(2026, 4, 19, 13, 8, 0, 0, time.UTC)
	snapshots := newIngestSnapshotStore()
	observer := &ingestMetricsObserver{}
	service := NewIngestService(snapshots, nil, nil, DefaultStatusPolicy(), fixedClock{now: now}, WithIngestMetrics(observer))

	if _, err := service.Heartbeat(context.Background(), WorkerReport{
		Identity:       WorkerIdentity{WorkerID: "worker-1", Namespace: "sim"},
		ProcessAlive:   true,
		AcceptingTasks: true,
		ObservedAt:     now,
	}); err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	if observer.snapshots != 1 || observer.events != 1 || observer.applied != 1 {
		t.Fatalf("observer calls = snapshots:%d events:%d applied:%d, want 1/1/1", observer.snapshots, observer.events, observer.applied)
	}
}

func TestBuildTaskHistoryRecordsTracksFinishedTasks(t *testing.T) {
	now := time.Date(2026, 4, 19, 13, 10, 0, 0, time.UTC)
	previous := WorkerSnapshot{
		Identity: WorkerIdentity{WorkerID: "worker-1"},
		ActiveTasks: []ActiveTask{{
			TaskID:    "task-1",
			TaskType:  "simulation",
			Phase:     TaskPhaseRunning,
			PhaseName: "running",
			StartedAt: now.Add(-2 * time.Minute),
			UpdatedAt: now.Add(-1 * time.Minute),
		}},
	}
	current := WorkerSnapshot{
		Identity: WorkerIdentity{WorkerID: "worker-1"},
	}
	records := BuildTaskHistoryRecords(previous, current, []DomainEvent{{
		Type:       EventTaskFinished,
		TaskID:     "task-1",
		OccurredAt: now,
	}}, now)

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	if records[0].Status != "finished" {
		t.Fatalf("status = %q, want finished", records[0].Status)
	}
	if records[0].EndedAt.IsZero() {
		t.Fatalf("expected ended_at to be set")
	}
}

func hasEventType(events []DomainEvent, want EventType) bool {
	for _, event := range events {
		if event.Type == want {
			return true
		}
	}
	return false
}
