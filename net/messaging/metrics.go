package messaging

import (
	"context"
	"time"

	"github.com/spcent/plumego/metrics"
)

// metricsWrapper wraps the MetricsCollector to provide messaging-specific
// helper methods for recording send latency, errors, and queue depth.
type metricsWrapper struct {
	collector metrics.MetricsCollector
}

func newMetrics(collector metrics.MetricsCollector) *metricsWrapper {
	return &metricsWrapper{collector: collector}
}

func (m *metricsWrapper) enabled() bool {
	return m != nil && m.collector != nil
}

// ObserveSend records a send operation (success or failure).
func (m *metricsWrapper) ObserveSend(ctx context.Context, channel Channel, provider string, duration time.Duration, err error) {
	if !m.enabled() {
		return
	}
	m.collector.ObserveMQ(ctx, "messaging_send", string(channel), duration, err, false)
	m.collector.Record(ctx, metrics.MetricRecord{
		Type:      "messaging_send",
		Name:      "messaging.send." + string(channel),
		Value:     duration.Seconds(),
		Labels:    metrics.MetricLabels{"channel": string(channel), "provider": provider},
		Duration:  duration,
		Error:     err,
		Timestamp: time.Now(),
	})
}

// ObserveEnqueue records an enqueue operation.
func (m *metricsWrapper) ObserveEnqueue(ctx context.Context, channel Channel, duration time.Duration, err error) {
	if !m.enabled() {
		return
	}
	m.collector.ObserveMQ(ctx, "messaging_enqueue", string(channel), duration, err, false)
}

// ObserveValidation records a validation failure.
func (m *metricsWrapper) ObserveValidation(ctx context.Context, channel Channel) {
	if !m.enabled() {
		return
	}
	m.collector.Record(ctx, metrics.MetricRecord{
		Type:      "messaging_validation_error",
		Name:      "messaging.validation." + string(channel),
		Value:     1,
		Labels:    metrics.MetricLabels{"channel": string(channel)},
		Timestamp: time.Now(),
	})
}
