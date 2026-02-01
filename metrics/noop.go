package metrics

import (
	"context"
	"time"
)

// NoopCollector is a metrics collector that does nothing.
// It implements the MetricsCollector interface but discards all metrics.
//
// This collector is useful for:
//   - Testing when you don't want to track metrics
//   - Production environments where metrics are disabled
//   - Benchmarking to measure overhead of metrics collection
//
// All operations are no-ops with minimal overhead.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	// Disable metrics collection
//	collector := metrics.NewNoopCollector()
//
//	// All methods do nothing
//	collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 100, 50*time.Millisecond)
//	collector.Clear() // No-op
type NoopCollector struct{}

// NewNoopCollector creates a new no-op metrics collector.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewNoopCollector()
func NewNoopCollector() *NoopCollector {
	return &NoopCollector{}
}

// Record does nothing and returns immediately.
func (n *NoopCollector) Record(ctx context.Context, record MetricRecord) {
	// No-op
}

// ObserveHTTP does nothing and returns immediately.
func (n *NoopCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	// No-op
}

// ObservePubSub does nothing and returns immediately.
func (n *NoopCollector) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	// No-op
}

// ObserveMQ does nothing and returns immediately.
func (n *NoopCollector) ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool) {
	// No-op
}

// ObserveKV does nothing and returns immediately.
func (n *NoopCollector) ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool) {
	// No-op
}

// ObserveIPC does nothing and returns immediately.
func (n *NoopCollector) ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error) {
	// No-op
}

// ObserveDB does nothing and returns immediately.
func (n *NoopCollector) ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
	// No-op
}

// GetStats returns an empty statistics structure.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewNoopCollector()
//	stats := collector.GetStats()
//	// stats will be zero-valued
func (n *NoopCollector) GetStats() CollectorStats {
	return CollectorStats{
		TotalRecords:  0,
		ErrorRecords:  0,
		ActiveSeries:  0,
		StartTime:     time.Time{},
		TypeBreakdown: make(map[MetricType]int64),
	}
}

// Clear does nothing and returns immediately.
func (n *NoopCollector) Clear() {
	// No-op
}

// Verify that NoopCollector implements MetricsCollector interface at compile time
var _ MetricsCollector = (*NoopCollector)(nil)
