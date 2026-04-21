package domain

import "time"

type SnapshotLister interface {
	ListCurrentWorkerSnapshots() ([]WorkerSnapshot, error)
}

type AlertRecordStore interface {
	ListAlertRecords() ([]AlertRecord, error)
	AppendAlert(record AlertRecord) error
}

type AlertMetricsObserver interface {
	ObserveAlerts(records []AlertRecord)
}

type AlertEngineOption func(*AlertEngine)

func WithAlertMetrics(observer AlertMetricsObserver) AlertEngineOption {
	return func(e *AlertEngine) {
		e.metrics = observer
	}
}

type AlertEngine struct {
	snapshots SnapshotLister
	alerts    AlertRecordStore
	clock     Clock
	policy    StatusPolicy
	metrics   AlertMetricsObserver
}

func NewAlertEngine(snapshots SnapshotLister, alerts AlertRecordStore, policy StatusPolicy, clock Clock, opts ...AlertEngineOption) *AlertEngine {
	if clock == nil {
		clock = systemClock{}
	}
	if policy == (StatusPolicy{}) {
		policy = DefaultStatusPolicy()
	}
	engine := &AlertEngine{
		snapshots: snapshots,
		alerts:    alerts,
		clock:     clock,
		policy:    policy,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(engine)
		}
	}
	return engine
}

func (e *AlertEngine) Evaluate() ([]AlertRecord, error) {
	now := e.clock.Now()
	snapshots, err := e.snapshots.ListCurrentWorkerSnapshots()
	if err != nil {
		return nil, err
	}
	existing, err := e.alerts.ListAlertRecords()
	if err != nil {
		return nil, err
	}

	candidates := make([]AlertRecord, 0, len(snapshots)*2)
	for _, snapshot := range snapshots {
		candidates = append(candidates, EvaluateSnapshotAlerts(snapshot, now, e.policy)...)
	}
	candidates = append(candidates, EvaluateTaskConflictAlerts(snapshots, now)...)

	emitted := reconcileAlerts(existing, candidates, now)
	for _, alert := range emitted {
		if err := e.alerts.AppendAlert(alert); err != nil {
			return nil, err
		}
	}
	if e.metrics != nil {
		e.metrics.ObserveAlerts(emitted)
	}
	return emitted, nil
}

func reconcileAlerts(existing []AlertRecord, candidates []AlertRecord, now time.Time) []AlertRecord {
	open := make(map[string]AlertRecord)
	for _, alert := range existing {
		if !isOpenAlert(alert) {
			continue
		}
		open[alert.DedupeKey] = alert
	}

	current := make(map[string]AlertRecord)
	for _, candidate := range candidates {
		current[candidate.DedupeKey] = candidate
	}

	out := make([]AlertRecord, 0, len(candidates)+len(open))
	for key, candidate := range current {
		if _, exists := open[key]; exists {
			delete(open, key)
			continue
		}
		out = append(out, candidate)
	}
	for _, alert := range open {
		out = append(out, resolveAlert(alert, now))
	}
	return out
}
