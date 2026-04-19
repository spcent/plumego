package domain

import "time"

type Clock interface {
	Now() time.Time
}

type SnapshotStore interface {
	UpsertWorkerSnapshot(snapshot WorkerSnapshot) error
	GetWorkerSnapshot(workerID WorkerID) (WorkerSnapshot, bool, error)
}

type TaskHistoryStore interface {
	AppendTaskHistory(record TaskHistoryRecord) error
}

type WorkerEventStore interface {
	AppendWorkerEvent(event DomainEvent) error
}

type RegisterCommand struct {
	Identity   WorkerIdentity
	ObservedAt time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

type IngestService struct {
	snapshots SnapshotStore
	history   TaskHistoryStore
	events    WorkerEventStore
	clock     Clock
	policy    StatusPolicy
}

func NewIngestService(
	snapshots SnapshotStore,
	history TaskHistoryStore,
	events WorkerEventStore,
	policy StatusPolicy,
	clock Clock,
) *IngestService {
	if clock == nil {
		clock = systemClock{}
	}
	if policy == (StatusPolicy{}) {
		policy = DefaultStatusPolicy()
	}
	return &IngestService{
		snapshots: snapshots,
		history:   history,
		events:    events,
		clock:     clock,
		policy:    policy,
	}
}

func (s *IngestService) Register(command RegisterCommand) (WorkerSnapshot, error) {
	now := s.clock.Now()
	snapshot, found, err := s.snapshots.GetWorkerSnapshot(command.Identity.WorkerID)
	if err != nil {
		return WorkerSnapshot{}, err
	}
	if !found {
		snapshot = WorkerSnapshot{}
	}

	snapshot.Identity = mergeIdentity(snapshot.Identity, command.Identity)
	snapshot, _ = applyStatus(WorkerSnapshot{Status: snapshot.Status}, snapshot, now, s.policy)
	if err := s.snapshots.UpsertWorkerSnapshot(snapshot); err != nil {
		return WorkerSnapshot{}, err
	}

	if !found && s.events != nil {
		if err := s.events.AppendWorkerEvent(DomainEvent{
			Type:       EventWorkerRegistered,
			OccurredAt: nonZeroTime(command.ObservedAt, now),
			WorkerID:   snapshot.Identity.WorkerID,
			Reason:     "worker_registered",
		}); err != nil {
			return WorkerSnapshot{}, err
		}
	}

	return snapshot, nil
}

func (s *IngestService) Heartbeat(report WorkerReport) (WorkerSnapshot, error) {
	now := s.clock.Now()
	previous, found, err := s.snapshots.GetWorkerSnapshot(report.Identity.WorkerID)
	if err != nil {
		return WorkerSnapshot{}, err
	}
	if !found {
		previous = WorkerSnapshot{}
	}

	merged, events := MergeWorkerReport(previous, report, now, s.policy)
	if err := s.snapshots.UpsertWorkerSnapshot(merged); err != nil {
		return WorkerSnapshot{}, err
	}
	if err := s.persistEvents(events); err != nil {
		return WorkerSnapshot{}, err
	}
	if err := s.persistTaskHistory(previous, merged, events, now); err != nil {
		return WorkerSnapshot{}, err
	}

	return merged, nil
}

func (s *IngestService) persistEvents(events []DomainEvent) error {
	if s.events == nil {
		return nil
	}
	for _, event := range events {
		if err := s.events.AppendWorkerEvent(event); err != nil {
			return err
		}
	}
	return nil
}

func (s *IngestService) persistTaskHistory(previous WorkerSnapshot, current WorkerSnapshot, events []DomainEvent, now time.Time) error {
	if s.history == nil {
		return nil
	}

	records := BuildTaskHistoryRecords(previous, current, events, now)
	for _, record := range records {
		if err := s.history.AppendTaskHistory(record); err != nil {
			return err
		}
	}
	return nil
}

func nonZeroTime(primary time.Time, fallback time.Time) time.Time {
	if !primary.IsZero() {
		return primary
	}
	return fallback
}
