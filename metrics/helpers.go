package metrics

import (
	"context"
	"time"
)

// Timer provides a convenient way to measure operation duration.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	timer := metrics.NewTimer()
//
//	// Perform operation
//	doWork()
//
//	// Record duration
//	duration := timer.Elapsed()
//	collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 100, duration)
type Timer struct {
	start time.Time
}

// NewTimer creates and starts a new timer.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	timer := metrics.NewTimer()
//	// ... do work ...
//	duration := timer.Elapsed()
func NewTimer() *Timer {
	return &Timer{start: time.Now()}
}

// Elapsed returns the duration since the timer was created.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	timer := metrics.NewTimer()
//	doSomething()
//	duration := timer.Elapsed()
//	fmt.Printf("Operation took: %v\n", duration)
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.start)
}

// Reset resets the timer to the current time.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	timer := metrics.NewTimer()
//	doFirstOperation()
//	duration1 := timer.Elapsed()
//
//	timer.Reset()
//	doSecondOperation()
//	duration2 := timer.Elapsed()
func (t *Timer) Reset() {
	t.start = time.Now()
}

// MeasureFunc measures the duration of a function and records it.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//
//	err := metrics.MeasureFunc(ctx, collector, "get", "user:123", func() error {
//		// Perform KV operation
//		return kvStore.Get("user:123")
//	}, metrics.MeasureWithKV(true))
func MeasureFunc(ctx context.Context, collector MetricsCollector, operation, subject string, fn func() error, opts ...MeasureOption) error {
	cfg := &measureConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	timer := NewTimer()
	err := fn()
	duration := timer.Elapsed()

	switch cfg.metricType {
	case "kv":
		hit := err == nil && cfg.kvHit
		collector.ObserveKV(ctx, operation, subject, duration, err, hit)
	case "pubsub":
		collector.ObservePubSub(ctx, operation, subject, duration, err)
	case "mq":
		collector.ObserveMQ(ctx, operation, subject, duration, err, cfg.mqPanicked)
	case "ipc":
		collector.ObserveIPC(ctx, operation, subject, cfg.ipcTransport, cfg.ipcBytes, duration, err)
	default:
		// Generic record
		record := MetricRecord{
			Type:     MetricType(operation),
			Name:     operation,
			Value:    float64(duration.Milliseconds()),
			Duration: duration,
			Error:    err,
		}
		collector.Record(ctx, record)
	}

	return err
}

type measureConfig struct {
	metricType   string
	kvHit        bool
	mqPanicked   bool
	ipcTransport string
	ipcBytes     int
}

// MeasureOption configures the MeasureFunc behavior.
type MeasureOption func(*measureConfig)

// MeasureWithKV configures measurement for KV operations.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	err := metrics.MeasureFunc(ctx, collector, "get", "key", fn, metrics.MeasureWithKV(true))
func MeasureWithKV(hit bool) MeasureOption {
	return func(cfg *measureConfig) {
		cfg.metricType = "kv"
		cfg.kvHit = hit
	}
}

// MeasureWithPubSub configures measurement for PubSub operations.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	err := metrics.MeasureFunc(ctx, collector, "publish", "topic", fn, metrics.MeasureWithPubSub())
func MeasureWithPubSub() MeasureOption {
	return func(cfg *measureConfig) {
		cfg.metricType = "pubsub"
	}
}

// MeasureWithMQ configures measurement for Message Queue operations.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	err := metrics.MeasureFunc(ctx, collector, "subscribe", "queue", fn, metrics.MeasureWithMQ(false))
func MeasureWithMQ(panicked bool) MeasureOption {
	return func(cfg *measureConfig) {
		cfg.metricType = "mq"
		cfg.mqPanicked = panicked
	}
}

// MeasureWithIPC configures measurement for IPC operations.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	err := metrics.MeasureFunc(ctx, collector, "read", "/tmp/socket", fn,
//		metrics.MeasureWithIPC("unix", 256))
func MeasureWithIPC(transport string, bytes int) MeasureOption {
	return func(cfg *measureConfig) {
		cfg.metricType = "ipc"
		cfg.ipcTransport = transport
		cfg.ipcBytes = bytes
	}
}

// RecordSuccess is a convenience function to record a successful operation.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	metrics.RecordSuccess(ctx, collector, "database_query", 50*time.Millisecond)
func RecordSuccess(ctx context.Context, collector MetricsCollector, operation string, duration time.Duration) {
	record := MetricRecord{
		Type:     MetricType(operation),
		Name:     operation,
		Value:    float64(duration.Milliseconds()),
		Duration: duration,
		Error:    nil,
	}
	collector.Record(ctx, record)
}

// RecordError is a convenience function to record a failed operation.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	if err := doOperation(); err != nil {
//		metrics.RecordError(ctx, collector, "database_query", 50*time.Millisecond, err)
//	}
func RecordError(ctx context.Context, collector MetricsCollector, operation string, duration time.Duration, err error) {
	record := MetricRecord{
		Type:     MetricType(operation),
		Name:     operation,
		Value:    float64(duration.Milliseconds()),
		Duration: duration,
		Error:    err,
	}
	collector.Record(ctx, record)
}

// RecordWithLabels is a convenience function to record a metric with custom labels.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	metrics.RecordWithLabels(ctx, collector, "api_call", 100*time.Millisecond,
//		metrics.MetricLabels{
//			"endpoint": "/api/users",
//			"method":   "GET",
//			"status":   "success",
//		})
func RecordWithLabels(ctx context.Context, collector MetricsCollector, operation string, duration time.Duration, labels MetricLabels) {
	record := MetricRecord{
		Type:     MetricType(operation),
		Name:     operation,
		Value:    float64(duration.Milliseconds()),
		Duration: duration,
		Labels:   labels,
	}
	collector.Record(ctx, record)
}

// MultiCollector wraps multiple collectors and forwards all operations to each.
//
// This is useful when you want to collect metrics to multiple destinations
// simultaneously (e.g., Prometheus and OpenTelemetry).
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	prom := metrics.NewPrometheusCollector("myapp")
//	otel := metrics.NewOpenTelemetryTracer("myapp")
//	multi := metrics.NewMultiCollector(prom, otel)
//
//	// Metrics are recorded to both collectors
//	multi.ObserveHTTP(ctx, "GET", "/api/users", 200, 100, 50*time.Millisecond)
type MultiCollector struct {
	collectors []MetricsCollector
}

// NewMultiCollector creates a new multi-collector that forwards to all provided collectors.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	prom := metrics.NewPrometheusCollector("myapp")
//	otel := metrics.NewOpenTelemetryTracer("myapp")
//	base := metrics.NewBaseMetricsCollector()
//	multi := metrics.NewMultiCollector(prom, otel, base)
func NewMultiCollector(collectors ...MetricsCollector) *MultiCollector {
	return &MultiCollector{collectors: collectors}
}

// Record forwards the record to all collectors.
func (m *MultiCollector) Record(ctx context.Context, record MetricRecord) {
	for _, c := range m.collectors {
		c.Record(ctx, record)
	}
}

// ObserveHTTP forwards the HTTP observation to all collectors.
func (m *MultiCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	for _, c := range m.collectors {
		c.ObserveHTTP(ctx, method, path, status, bytes, duration)
	}
}

// ObservePubSub forwards the PubSub observation to all collectors.
func (m *MultiCollector) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	for _, c := range m.collectors {
		c.ObservePubSub(ctx, operation, topic, duration, err)
	}
}

// ObserveMQ forwards the MQ observation to all collectors.
func (m *MultiCollector) ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool) {
	for _, c := range m.collectors {
		c.ObserveMQ(ctx, operation, topic, duration, err, panicked)
	}
}

// ObserveKV forwards the KV observation to all collectors.
func (m *MultiCollector) ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool) {
	for _, c := range m.collectors {
		c.ObserveKV(ctx, operation, key, duration, err, hit)
	}
}

// ObserveIPC forwards the IPC observation to all collectors.
func (m *MultiCollector) ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error) {
	for _, c := range m.collectors {
		c.ObserveIPC(ctx, operation, addr, transport, bytes, duration, err)
	}
}

// GetStats returns combined statistics from all collectors.
// The statistics are aggregated across all collectors.
func (m *MultiCollector) GetStats() CollectorStats {
	if len(m.collectors) == 0 {
		return CollectorStats{}
	}

	combined := CollectorStats{
		TypeBreakdown: make(map[MetricType]int64),
	}

	for _, c := range m.collectors {
		stats := c.GetStats()
		combined.TotalRecords += stats.TotalRecords
		combined.ErrorRecords += stats.ErrorRecords
		combined.ActiveSeries += stats.ActiveSeries
		combined.TotalRequests += stats.TotalRequests
		combined.TotalSpans += stats.TotalSpans
		combined.ErrorSpans += stats.ErrorSpans
		combined.TotalDuration += stats.TotalDuration

		// Merge type breakdown
		for k, v := range stats.TypeBreakdown {
			combined.TypeBreakdown[k] += v
		}

		// Use earliest start time
		if combined.StartTime.IsZero() || (!stats.StartTime.IsZero() && stats.StartTime.Before(combined.StartTime)) {
			combined.StartTime = stats.StartTime
		}
	}

	// Calculate average latency
	if combined.TotalRequests > 0 {
		// This is a simple average across all collectors
		totalLatency := 0.0
		for _, c := range m.collectors {
			totalLatency += c.GetStats().AverageLatency
		}
		combined.AverageLatency = totalLatency / float64(len(m.collectors))
	}

	// Calculate average duration
	if combined.TotalSpans > 0 {
		combined.AverageDuration = combined.TotalDuration / time.Duration(combined.TotalSpans)
	}

	return combined
}

// Clear clears all collectors.
func (m *MultiCollector) Clear() {
	for _, c := range m.collectors {
		c.Clear()
	}
}

// Verify that MultiCollector implements MetricsCollector interface at compile time
var _ MetricsCollector = (*MultiCollector)(nil)
