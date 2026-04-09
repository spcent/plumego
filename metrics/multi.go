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

// GetStats returns combined statistics from all collectors.
func (m *MultiCollector) GetStats() CollectorStats {
	if len(m.collectors) == 0 {
		return CollectorStats{NameBreakdown: make(map[string]int64)}
	}

	combined := CollectorStats{
		NameBreakdown: make(map[string]int64),
	}

	for _, c := range m.collectors {
		stats := c.GetStats()
		combined.TotalRecords += stats.TotalRecords
		combined.ErrorRecords += stats.ErrorRecords
		combined.ActiveSeries += stats.ActiveSeries

		for k, v := range stats.NameBreakdown {
			combined.NameBreakdown[k] += v
		}

		if combined.StartTime.IsZero() || (!stats.StartTime.IsZero() && stats.StartTime.Before(combined.StartTime)) {
			combined.StartTime = stats.StartTime
		}
	}

	if combined.ActiveSeries == 0 && len(combined.NameBreakdown) > 0 {
		combined.ActiveSeries = len(combined.NameBreakdown)
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
