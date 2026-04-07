package metrics

import (
	"context"
	"time"
)

// MultiCollector wraps multiple aggregate collectors and forwards all operations to each.
type MultiCollector struct {
	collectors []AggregateCollector
}

// NewMultiCollector creates a new multi-collector that forwards to all provided collectors.
func NewMultiCollector(collectors ...AggregateCollector) *MultiCollector {
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

// ObserveDB forwards the database observation to all collectors.
func (m *MultiCollector) ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
	for _, c := range m.collectors {
		c.ObserveDB(ctx, operation, driver, query, rows, duration, err)
	}
}

// GetStats returns combined statistics from all collectors.
func (m *MultiCollector) GetStats() CollectorStats {
	if len(m.collectors) == 0 {
		return CollectorStats{TypeBreakdown: make(map[MetricType]int64)}
	}

	combined := CollectorStats{
		TypeBreakdown: make(map[MetricType]int64),
	}

	for _, c := range m.collectors {
		stats := c.GetStats()
		combined.TotalRecords += stats.TotalRecords
		combined.ErrorRecords += stats.ErrorRecords
		combined.ActiveSeries += stats.ActiveSeries

		for k, v := range stats.TypeBreakdown {
			combined.TypeBreakdown[k] += v
		}

		if combined.StartTime.IsZero() || (!stats.StartTime.IsZero() && stats.StartTime.Before(combined.StartTime)) {
			combined.StartTime = stats.StartTime
		}
	}

	if combined.ActiveSeries == 0 && len(combined.TypeBreakdown) > 0 {
		combined.ActiveSeries = len(combined.TypeBreakdown)
	}

	return combined
}

// Clear clears all collectors.
func (m *MultiCollector) Clear() {
	for _, c := range m.collectors {
		c.Clear()
	}
}

var _ AggregateCollector = (*MultiCollector)(nil)
