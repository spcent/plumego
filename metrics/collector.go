package metrics

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	// HTTP metrics
	MetricHTTPRequest = "http_request"
)

// MetricLabels represents key-value labels for metrics
type MetricLabels map[string]string

const (
	labelMethod = "method"
	labelPath   = "path"
	labelStatus = "status"
)

// MetricRecord represents a single metric record
type MetricRecord struct {
	// Name is the canonical metric identity across stable and extension-owned records.
	Name string
	// Value uses seconds as the canonical unit for duration/latency metrics.
	// Non-duration metrics (for example counters or queue depth) may use domain-specific units.
	Value     float64
	Labels    MetricLabels
	Timestamp time.Time
	Duration  time.Duration
	Error     error
}

func durationValueSeconds(duration time.Duration) float64 {
	return duration.Seconds()
}

// AggregateCollector is the stable collector surface.
// Prefer narrower interfaces at module boundaries and reserve this contract for
// collector implementations or fan-out adapters that need generic recording,
// shared HTTP observation, stats, and reset semantics.
type AggregateCollector interface {
	// Record records a single metric
	Record(ctx context.Context, record MetricRecord)

	// ObserveHTTP is a convenience method for HTTP request metrics
	ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration)

	// GetStats returns current statistics
	GetStats() CollectorStats

	// Clear resets all metrics
	Clear()
}

// Recorder captures generic metric records without any domain-specific helpers.
type Recorder interface {
	Record(ctx context.Context, record MetricRecord)
}

// HTTPObserver captures only HTTP request metrics.
// Use this narrower contract in transport middleware that should not depend on
// the full cross-module metrics surface.
type HTTPObserver interface {
	ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration)
}

// StatsReader exposes aggregated collector statistics.
type StatsReader interface {
	GetStats() CollectorStats
}

// Resetter clears accumulated metrics state.
type Resetter interface {
	Clear()
}

// CollectorStats provides a contract for collector statistics payloads.
//
// Mandatory fields for all collectors:
//   - TotalRecords: total number of records processed by the collector.
//   - ErrorRecords: number of records classified as errors.
//   - ActiveSeries: number of currently active metric series maintained.
//   - StartTime: collector start/reset time (zero only when intentionally unknown).
//
// Optional fields:
//   - NameBreakdown: per-metric-name counters when the collector can provide them.
type CollectorStats struct {
	TotalRecords  int64            `json:"total_records"`
	ErrorRecords  int64            `json:"error_records"`
	ActiveSeries  int              `json:"active_series"`
	StartTime     time.Time        `json:"start_time"`
	NameBreakdown map[string]int64 `json:"name_breakdown"`
}

// BaseMetricsCollector provides a base implementation for metrics collectors
type BaseMetricsCollector struct {
	mu    sync.RWMutex
	stats CollectorStats
}

// NewBaseMetricsCollector creates a new base metrics collector.
func NewBaseMetricsCollector() *BaseMetricsCollector {
	return &BaseMetricsCollector{
		stats: CollectorStats{
			StartTime:     time.Now(),
			NameBreakdown: make(map[string]int64),
		},
	}
}

// Record implements the aggregate collector contract.
func (b *BaseMetricsCollector) Record(ctx context.Context, record MetricRecord) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ensureInitializedLocked()

	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	if len(record.Labels) > 0 {
		record.Labels = cloneLabels(record.Labels)
	}

	b.stats.TotalRecords++
	if record.Name != "" {
		b.stats.NameBreakdown[record.Name]++
	}

	if metricRecordIsError(record) {
		b.stats.ErrorRecords++
	}
}

// ObserveHTTP implements HTTP metrics recording
func (b *BaseMetricsCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	record := MetricRecord{
		Name:  MetricHTTPRequest,
		Value: durationValueSeconds(duration),
		Labels: MetricLabels{
			labelMethod: method,
			labelPath:   path,
			labelStatus: strconv.Itoa(status),
		},
		Duration: duration,
	}
	b.Record(ctx, record)
}

// GetStats returns current statistics
func (b *BaseMetricsCollector) GetStats() CollectorStats {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stats := b.stats
	if stats.ActiveSeries == 0 && len(stats.NameBreakdown) > 0 {
		stats.ActiveSeries = len(stats.NameBreakdown)
	}
	if stats.NameBreakdown != nil {
		stats.NameBreakdown = cloneBreakdown(stats.NameBreakdown)
	}
	return stats
}

// Clear resets all metrics
func (b *BaseMetricsCollector) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ensureInitializedLocked()

	b.stats = CollectorStats{
		StartTime:     time.Now(),
		NameBreakdown: make(map[string]int64),
	}
}

func (b *BaseMetricsCollector) ensureInitializedLocked() {
	if b.stats.NameBreakdown == nil {
		b.stats.NameBreakdown = make(map[string]int64)
	}
	if b.stats.StartTime.IsZero() {
		b.stats.StartTime = time.Now()
	}
}

func cloneLabels(labels MetricLabels) MetricLabels {
	if len(labels) == 0 {
		return nil
	}
	result := make(MetricLabels, len(labels))
	for key, value := range labels {
		result[key] = value
	}
	return result
}

func cloneBreakdown(breakdown map[string]int64) map[string]int64 {
	if breakdown == nil {
		return nil
	}
	result := make(map[string]int64, len(breakdown))
	for key, value := range breakdown {
		result[key] = value
	}
	return result
}

func metricRecordIsError(record MetricRecord) bool {
	if record.Error != nil {
		return true
	}
	if record.Name != MetricHTTPRequest {
		return false
	}
	status, err := strconv.Atoi(record.Labels[labelStatus])
	return err == nil && status >= http.StatusBadRequest
}
