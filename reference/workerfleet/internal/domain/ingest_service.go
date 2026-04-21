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

type IngestMetricsObserver interface {
	ObserveWorkerSnapshot(previous WorkerSnapshot, current WorkerSnapshot)
	ObserveWorkerReportApplied(operation string, duration time.Duration)
}

type IngestOption func(*IngestService)

func WithIngestMetrics(observer IngestMetricsObserver) IngestOption {
	return func(s *IngestService) {
		s.metrics = observer
	}
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
	metrics   IngestMetricsObserver
}

func NewIngestService(
	snapshots SnapshotStore,
	history TaskHistoryStore,
	events WorkerEventStore,
	policy StatusPolicy,
	clock Clock,
	opts ...IngestOption,
) *IngestService {
	if clock == nil {
		clock = systemClock{}
	}
	if policy == (StatusPolicy{}) {
		policy = DefaultStatusPolicy()
	}
	service := &IngestService{
		snapshots: snapshots,
		history:   history,
		events:    events,
		clock:     clock,
		policy:    policy,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(service)
		}
	}
	return service
}

func (s *IngestService) Register(command RegisterCommand) (WorkerSnapshot, error) {
	started := time.Now()
	now := s.clock.Now()
	previous, found, err := s.snapshots.GetWorkerSnapshot(command.Identity.WorkerID)
	if err != nil {
		return WorkerSnapshot{}, err
	}
	if !found {
		previous = WorkerSnapshot{}
	}

	snapshot := previous
	snapshot.Identity = mergeIdentity(previous.Identity, command.Identity)
	snapshot, _ = applyStatus(WorkerSnapshot{Status: previous.Status}, snapshot, now, s.policy)
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
	if s.metrics != nil {
		s.metrics.ObserveWorkerSnapshot(previous, snapshot)
		s.metrics.ObserveWorkerReportApplied("register", time.Since(started))
	}

	return snapshot, nil
}

func (s *IngestService) Heartbeat(report WorkerReport) (WorkerSnapshot, error) {
	started := time.Now()
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
	if s.metrics != nil {
		s.metrics.ObserveWorkerSnapshot(previous, merged)
		s.metrics.ObserveWorkerReportApplied("heartbeat", time.Since(started))
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
