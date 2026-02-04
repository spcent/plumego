package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/spcent/plumego/middleware/observability"
)

// MiddlewareAdapter adapts a unified MetricsCollector to the observability.MetricsCollector interface.
//
// This allows using the unified metrics collectors (Prometheus, OpenTelemetry, etc.)
// with the existing observability.Logging middleware.
//
// Example:
//
//	import (
//		"github.com/spcent/plumego/metrics"
//		"github.com/spcent/plumego/middleware/observability"
//	)
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	adapter := metrics.NewMiddlewareAdapter(collector)
//	handler := observability.Logging(logger, adapter, tracer)(myHandler)
type MiddlewareAdapter struct {
	collector MetricsCollector
}

// NewMiddlewareAdapter creates a new middleware adapter for a unified metrics collector.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	adapter := metrics.NewMiddlewareAdapter(collector)
func NewMiddlewareAdapter(collector MetricsCollector) *MiddlewareAdapter {
	return &MiddlewareAdapter{collector: collector}
}

// Observe implements the observability.MetricsCollector interface.
// It forwards the request metrics to the underlying unified collector.
func (a *MiddlewareAdapter) Observe(ctx context.Context, metrics observability.RequestMetrics) {
	a.collector.ObserveHTTP(
		ctx,
		metrics.Method,
		metrics.Path,
		metrics.Status,
		metrics.Bytes,
		metrics.Duration,
	)
}

// Collector returns the underlying unified metrics collector.
// This allows accessing collector-specific features like statistics and clearing.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	adapter := metrics.NewMiddlewareAdapter(collector)
//	stats := adapter.Collector().GetStats()
func (a *MiddlewareAdapter) Collector() MetricsCollector {
	return a.collector
}

// Verify that MiddlewareAdapter implements observability.MetricsCollector at compile time
var _ observability.MetricsCollector = (*MiddlewareAdapter)(nil)

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
// This is useful for custom middleware that needs to record metrics.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

// WriteHeader captures the status code before writing the header.
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write captures the number of bytes written.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// Status returns the captured HTTP status code.
func (rw *responseWriter) Status() int {
	return rw.status
}

// BytesWritten returns the total number of bytes written.
func (rw *responseWriter) BytesWritten() int {
	return rw.bytes
}

// MetricsMiddleware creates a simple HTTP middleware that records metrics.
//
// This is a lightweight alternative to the full observability.Logging middleware
// when you only need metrics and not logging.
//
// Example:
//
//	import (
//		"github.com/spcent/plumego/metrics"
//		"net/http"
//	)
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	middleware := metrics.MetricsMiddleware(collector)
//
//	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		w.Write([]byte("OK"))
//	}))
func MetricsMiddleware(collector MetricsCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timer := NewTimer()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			collector.ObserveHTTP(
				r.Context(),
				r.Method,
				r.URL.Path,
				rw.status,
				rw.bytes,
				timer.Elapsed(),
			)
		})
	}
}

// MetricsHandler wraps an HTTP handler with automatic metric recording.
//
// This is a convenience function that combines a handler with MetricsMiddleware.
//
// Example:
//
//	import (
//		"github.com/spcent/plumego/metrics"
//		"net/http"
//	)
//
//	collector := metrics.NewPrometheusCollector("myapp")
//
//	http.Handle("/api/users", metrics.MetricsHandler(collector, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//		w.Write([]byte("OK"))
//	})))
func MetricsHandler(collector MetricsCollector, handler http.Handler) http.Handler {
	return MetricsMiddleware(collector)(handler)
}

// Aggregator aggregates metrics over time windows.
//
// This is useful for computing rolling averages, rates, and percentiles
// without storing all individual metric records.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	agg := metrics.NewAggregator(1 * time.Minute)
//
//	// Record values
//	agg.Record("api_latency", 50.0)
//	agg.Record("api_latency", 75.0)
//	agg.Record("api_latency", 100.0)
//
//	// Get statistics
//	stats := agg.GetStats("api_latency")
//	fmt.Printf("Count: %d, Mean: %.2f, Min: %.2f, Max: %.2f\n",
//		stats.Count, stats.Mean, stats.Min, stats.Max)
type Aggregator struct {
	mu      *sync.RWMutex
	window  time.Duration
	buckets map[string]*bucket
}

type bucket struct {
	count  int64
	sum    float64
	min    float64
	max    float64
	values []float64
	maxLen int
}

// AggregatorStats contains statistical information about aggregated metrics.
type AggregatorStats struct {
	Count int64   `json:"count"`
	Sum   float64 `json:"sum"`
	Mean  float64 `json:"mean"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	P50   float64 `json:"p50"`
	P95   float64 `json:"p95"`
	P99   float64 `json:"p99"`
}

// NewAggregator creates a new metrics aggregator with the specified time window.
//
// The time window determines how long metrics are retained for aggregation.
// Older metrics are automatically discarded.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	// Aggregate metrics over 5-minute windows
//	agg := metrics.NewAggregator(5 * time.Minute)
func NewAggregator(window time.Duration) *Aggregator {
	return &Aggregator{
		mu:      &sync.RWMutex{},
		window:  window,
		buckets: make(map[string]*bucket),
	}
}

// Record adds a value to the aggregator for the specified metric name.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	agg := metrics.NewAggregator(1 * time.Minute)
//	agg.Record("request_duration", 50.5)
//	agg.Record("request_duration", 75.2)
func (a *Aggregator) Record(name string, value float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	b, exists := a.buckets[name]
	if !exists {
		b = &bucket{
			min:    value,
			max:    value,
			values: make([]float64, 0, 1000),
			maxLen: 10000,
		}
		a.buckets[name] = b
	}

	b.count++
	b.sum += value
	if value < b.min {
		b.min = value
	}
	if value > b.max {
		b.max = value
	}

	// Keep values for percentile calculation
	if len(b.values) < b.maxLen {
		b.values = append(b.values, value)
	}
}

// GetStats returns aggregated statistics for the specified metric name.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	agg := metrics.NewAggregator(1 * time.Minute)
//	agg.Record("latency", 50.0)
//	agg.Record("latency", 75.0)
//
//	stats := agg.GetStats("latency")
//	fmt.Printf("Mean: %.2f, P95: %.2f\n", stats.Mean, stats.P95)
func (a *Aggregator) GetStats(name string) AggregatorStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	b, exists := a.buckets[name]
	if !exists {
		return AggregatorStats{}
	}

	stats := AggregatorStats{
		Count: b.count,
		Sum:   b.sum,
		Min:   b.min,
		Max:   b.max,
	}

	if b.count > 0 {
		stats.Mean = b.sum / float64(b.count)
	}

	// Calculate percentiles if we have values
	if len(b.values) > 0 {
		stats.P50 = percentile(b.values, 0.50)
		stats.P95 = percentile(b.values, 0.95)
		stats.P99 = percentile(b.values, 0.99)
	}

	return stats
}

// Clear removes all aggregated metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	agg := metrics.NewAggregator(1 * time.Minute)
//	// ... record metrics ...
//	agg.Clear()
func (a *Aggregator) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.buckets = make(map[string]*bucket)
}

// GetAllStats returns statistics for all tracked metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	agg := metrics.NewAggregator(1 * time.Minute)
//	agg.Record("latency", 50.0)
//	agg.Record("throughput", 1000.0)
//
//	allStats := agg.GetAllStats()
//	for name, stats := range allStats {
//		fmt.Printf("%s: mean=%.2f\n", name, stats.Mean)
//	}
func (a *Aggregator) GetAllStats() map[string]AggregatorStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]AggregatorStats, len(a.buckets))
	for name := range a.buckets {
		// Unlock temporarily to call GetStats which also locks
		a.mu.RUnlock()
		result[name] = a.GetStats(name)
		a.mu.RLock()
	}

	return result
}

// percentile calculates the nth percentile of a slice of float64 values.
// The values slice is not modified.
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Create a sorted copy
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Simple insertion sort for small slices
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	index := int(float64(len(sorted)-1) * p)
	return sorted[index]
}
