package domain

import (
	"context"
	"time"
)

type SnapshotLister interface {
	ListCurrentWorkerSnapshots(ctx context.Context) ([]WorkerSnapshot, error)
}

type AlertRecordStore interface {
	ListAlertRecords(ctx context.Context) ([]AlertRecord, error)
	AppendAlert(ctx context.Context, record AlertRecord) error
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

func WithAlertPolicy(policy AlertPolicy) AlertEngineOption {
	return func(e *AlertEngine) {
		e.alertPolicy = policy
	}
}

type AlertEngine struct {
	snapshots   SnapshotLister
	alerts      AlertRecordStore
	clock       Clock
	policy      StatusPolicy
	alertPolicy AlertPolicy
	metrics     AlertMetricsObserver
}

func NewAlertEngine(snapshots SnapshotLister, alerts AlertRecordStore, policy StatusPolicy, clock Clock, opts ...AlertEngineOption) *AlertEngine {
	if clock == nil {
		clock = systemClock{}
	}
	if policy == (StatusPolicy{}) {
		policy = DefaultStatusPolicy()
	}
	if err := policy.Validate(); err != nil {
		policy = DefaultStatusPolicy()
	}
	engine := &AlertEngine{
		snapshots:   snapshots,
		alerts:      alerts,
		clock:       clock,
		policy:      policy,
		alertPolicy: policy.AlertPolicy(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(engine)
		}
	}
	if err := engine.alertPolicy.Validate(); err != nil {
		engine.alertPolicy = policy.AlertPolicy()
	}
	return engine
}

func (e *AlertEngine) Evaluate(ctx context.Context) ([]AlertRecord, error) {
	now := e.clock.Now()
	snapshots, err := e.snapshots.ListCurrentWorkerSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := e.alerts.ListAlertRecords(ctx)
	if err != nil {
		return nil, err
	}

	candidates := make([]AlertRecord, 0, len(snapshots)*2)
	for _, snapshot := range snapshots {
		candidates = append(candidates, EvaluateSnapshotAlerts(snapshot, now, e.alertPolicy)...)
	}
	candidates = append(candidates, EvaluateTaskConflictAlerts(snapshots, now)...)

	emitted := reconcileAlerts(existing, candidates, now)
	for _, alert := range emitted {
		if err := e.alerts.AppendAlert(ctx, alert); err != nil {
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
