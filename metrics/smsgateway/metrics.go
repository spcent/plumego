package smsgateway

import (
	"context"
	"strconv"
	"time"

	"github.com/spcent/plumego/metrics"
)

const (
	LabelQueue    = "queue"
	LabelState    = "state"
	LabelProvider = "provider"
	LabelTenant   = "tenant"
	LabelResult   = "result"
	LabelStatus   = "status"
	LabelAttempt  = "attempt"
)

const (
	MetricQueueDepth     = "queue_depth"
	MetricSendLatency    = "send_latency"
	MetricProviderResult = "provider_result"
	MetricReceiptDelay   = "receipt_delay"
	MetricRetry          = "retry"
	MetricStatus         = "status"
)

// QueueStats summarizes queue depth by state.
type QueueStats struct {
	Queued  int64
	Leased  int64
	Dead    int64
	Expired int64
}

// Reporter provides a unified SMS gateway metrics interface.
type Reporter struct {
	Collector metrics.MetricsCollector
	Now       func() time.Time
}

// NewReporter creates a new Reporter.
func NewReporter(collector metrics.MetricsCollector) *Reporter {
	return &Reporter{Collector: collector, Now: time.Now}
}

// RecordQueueDepth emits queue depth metrics.
func (r *Reporter) RecordQueueDepth(ctx context.Context, queue string, stats QueueStats) {
	if r == nil || r.Collector == nil {
		return
	}
	r.record(ctx, MetricQueueDepth, float64(stats.Queued), labels(queue, "queued", "", "", "", "", 0), 0, nil)
	r.record(ctx, MetricQueueDepth, float64(stats.Leased), labels(queue, "leased", "", "", "", "", 0), 0, nil)
	r.record(ctx, MetricQueueDepth, float64(stats.Dead), labels(queue, "dead", "", "", "", "", 0), 0, nil)
	r.record(ctx, MetricQueueDepth, float64(stats.Expired), labels(queue, "expired", "", "", "", "", 0), 0, nil)
}

// RecordSendLatency emits send latency metrics per provider.
func (r *Reporter) RecordSendLatency(ctx context.Context, tenant, provider string, duration time.Duration, err error) {
	if r == nil || r.Collector == nil {
		return
	}
	r.record(ctx, MetricSendLatency, float64(duration.Milliseconds()), labels("", "", provider, tenant, "", "", 0), duration, err)
}

// RecordProviderResult records a provider success/failure sample.
func (r *Reporter) RecordProviderResult(ctx context.Context, tenant, provider string, success bool) {
	if r == nil || r.Collector == nil {
		return
	}
	result := "success"
	if !success {
		result = "failure"
	}
	r.record(ctx, MetricProviderResult, 1, labels("", "", provider, tenant, result, "", 0), 0, nil)
}

// RecordReceiptDelay records receipt latency from send to delivery.
func (r *Reporter) RecordReceiptDelay(ctx context.Context, tenant, provider string, delay time.Duration) {
	if r == nil || r.Collector == nil {
		return
	}
	r.record(ctx, MetricReceiptDelay, float64(delay.Milliseconds()), labels("", "", provider, tenant, "", "", 0), delay, nil)
}

// RecordRetry records retry attempts (attempt >= 2 recommended).
func (r *Reporter) RecordRetry(ctx context.Context, tenant, provider string, attempt int) {
	if r == nil || r.Collector == nil {
		return
	}
	r.record(ctx, MetricRetry, 1, labels("", "", provider, tenant, "", "", attempt), 0, nil)
}

// RecordStatus records message status transitions.
func (r *Reporter) RecordStatus(ctx context.Context, tenant string, status string) {
	if r == nil || r.Collector == nil {
		return
	}
	r.record(ctx, MetricStatus, 1, labels("", "", "", tenant, "", status, 0), 0, nil)
}

func (r *Reporter) record(ctx context.Context, name string, value float64, labels metrics.MetricLabels, duration time.Duration, err error) {
	record := metrics.MetricRecord{
		Type:     metrics.MetricSMSGateway,
		Name:     name,
		Value:    value,
		Labels:   labels,
		Duration: duration,
		Error:    err,
	}
	r.Collector.Record(ctx, record)
}

func labels(queue, state, provider, tenant, result, status string, attempt int) metrics.MetricLabels {
	lbls := metrics.MetricLabels{}
	if queue != "" {
		lbls[LabelQueue] = queue
	}
	if state != "" {
		lbls[LabelState] = state
	}
	if provider != "" {
		lbls[LabelProvider] = provider
	}
	if tenant != "" {
		lbls[LabelTenant] = tenant
	}
	if result != "" {
		lbls[LabelResult] = result
	}
	if status != "" {
		lbls[LabelStatus] = status
	}
	if attempt > 0 {
		lbls[LabelAttempt] = strconv.Itoa(attempt)
	}
	return lbls
}
