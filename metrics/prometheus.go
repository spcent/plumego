package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/spcent/plumego/middleware"
)

type labelKey struct {
	method string
	path   string
	status string
}

type latencyStats struct {
	count uint64
	sum   float64
	min   float64
	max   float64
}

// PrometheusCollector implements MetricsCollector without third-party dependencies.
// It exposes a text-based metrics handler compatible with Prometheus exposition format.
//
// This collector is designed for production use with:
//   - Memory-efficient metric storage
//   - Automatic eviction of old metrics
//   - Prometheus-compatible output format
//   - No external dependencies
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	defer collector.Clear()
//
//	// Use in HTTP middleware
//	mux := http.NewServeMux()
//	mux.Handle("/metrics", collector.Handler())
//
//	// Record metrics
//	collector.ObserveHTTP(context.Background(), "GET", "/api/users", 200, 100, 50*time.Millisecond)
type PrometheusCollector struct {
	namespace string
	maxMemory int // Maximum number of metric series to store

	mu        sync.RWMutex
	requests  map[labelKey]uint64
	durations map[labelKey]latencyStats
	startTime time.Time

	// Base collector for unified interface
	base     *BaseMetricsCollector
	baseOnce sync.Once
}

// NewPrometheusCollector constructs an in-memory collector with the provided namespace.
// An empty namespace defaults to "plumego".
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
func NewPrometheusCollector(namespace string) *PrometheusCollector {
	if namespace == "" {
		namespace = "plumego"
	}
	return &PrometheusCollector{
		namespace: namespace,
		requests:  make(map[labelKey]uint64),
		durations: make(map[labelKey]latencyStats),
		startTime: time.Now(),
		maxMemory: 10000, // Default: allow up to 10k unique label combinations
	}
}

// WithMaxMemory sets the maximum number of unique metric series to store.
// When exceeded, oldest entries are evicted.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp").
//		WithMaxMemory(50000)
func (p *PrometheusCollector) WithMaxMemory(max int) *PrometheusCollector {
	if max > 0 {
		p.maxMemory = max
	}
	return p
}

// Observe records a single HTTP request metric set.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//	import "github.com/spcent/plumego/middleware"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	collector.Observe(context.Background(), middleware.RequestMetrics{
//		Method:   "GET",
//		Path:     "/api/users",
//		Status:   200,
//		Bytes:    100,
//		Duration: 50 * time.Millisecond,
//	})
func (p *PrometheusCollector) Observe(_ context.Context, m middleware.RequestMetrics) {
	key := labelKey{m.Method, m.Path, strconv.Itoa(m.Status)}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check memory limit and evict if needed
	if p.maxMemory > 0 && len(p.requests) >= p.maxMemory {
		p.evictOldest()
	}

	p.requests[key]++

	duration := m.Duration.Seconds()
	stats := p.durations[key]
	stats.count++
	stats.sum += duration

	// Update min/max
	if stats.count == 1 || duration < stats.min {
		stats.min = duration
	}
	if stats.count == 1 || duration > stats.max {
		stats.max = duration
	}

	p.durations[key] = stats
}

// Handler returns an HTTP handler that emits the current metrics snapshot.
//
// This handler exposes metrics in Prometheus exposition format (text/plain).
// The endpoint can be scraped by Prometheus or viewed directly in a browser.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	http.Handle("/metrics", collector.Handler())
//	http.ListenAndServe(":8080", nil)
func (p *PrometheusCollector) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests, durations, uptime := p.snapshot()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		// Write metrics
		fmt.Fprintf(w, "# HELP %s_http_requests_total Total number of HTTP requests processed.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_requests_total counter\n", p.namespace)

		reqKeys := sortedKeys(requests)
		for _, k := range reqKeys {
			fmt.Fprintf(w, "%s_http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				p.namespace, k.method, k.path, k.status, requests[k])
		}

		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_http_request_duration_seconds_sum Sum of HTTP request latencies in seconds.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_request_duration_seconds_summary summary\n", p.namespace)

		durKeys := sortedKeys(durations)
		for _, k := range durKeys {
			stats := durations[k]
			fmt.Fprintf(w, "%s_http_request_duration_seconds_sum{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				p.namespace, k.method, k.path, k.status, stats.sum)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_count{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				p.namespace, k.method, k.path, k.status, stats.count)
			// Add min and max as additional metrics
			fmt.Fprintf(w, "%s_http_request_duration_seconds_min{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				p.namespace, k.method, k.path, k.status, stats.min)
			fmt.Fprintf(w, "%s_http_request_duration_seconds_max{method=\"%s\",path=\"%s\",status=\"%s\"} %.9f\n",
				p.namespace, k.method, k.path, k.status, stats.max)
		}

		// Add uptime metric
		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_uptime_seconds Total uptime in seconds.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_uptime_seconds gauge\n", p.namespace)
		fmt.Fprintf(w, "%s_uptime_seconds %.3f\n", p.namespace, uptime.Seconds())

		// Add total request count
		fmt.Fprintln(w)
		fmt.Fprintf(w, "# HELP %s_http_requests_total_all Total requests across all labels.\n", p.namespace)
		fmt.Fprintf(w, "# TYPE %s_http_requests_total_all counter\n", p.namespace)
		totalRequests := uint64(0)
		for _, count := range requests {
			totalRequests += count
		}
		fmt.Fprintf(w, "%s_http_requests_total_all %d\n", p.namespace, totalRequests)
	})
}

// GetStats returns statistics about the collected metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	// ... use collector ...
//	stats := collector.GetStats()
//	fmt.Printf("Total requests: %d, Average latency: %.3f\n", stats.TotalRequests, stats.AverageLatency)
func (p *PrometheusCollector) GetStats() CollectorStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var stats CollectorStats
	stats.Series = len(p.requests)
	stats.StartTime = p.startTime

	// Calculate request rate
	totalRequests := uint64(0)
	for _, count := range p.requests {
		totalRequests += count
	}
	stats.TotalRequests = totalRequests

	// Calculate average latency
	var totalDuration float64
	var totalSamples uint64
	for _, d := range p.durations {
		totalDuration += d.sum
		totalSamples += d.count
	}
	if totalSamples > 0 {
		stats.AverageLatency = totalDuration / float64(totalSamples)
	}

	return stats
}

// Clear resets all collected metrics.
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	collector := metrics.NewPrometheusCollector("myapp")
//	// ... use collector ...
//	collector.Clear()
func (p *PrometheusCollector) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.requests = make(map[labelKey]uint64)
	p.durations = make(map[labelKey]latencyStats)
	p.startTime = time.Now()

	if p.base != nil {
		p.base.Clear()
	}
}

// Record implements the unified MetricsCollector interface
func (p *PrometheusCollector) Record(ctx context.Context, record MetricRecord) {
	// For HTTP requests, use the existing Prometheus format
	if record.Type == MetricHTTPRequest {
		if metrics, ok := httpMetricsFromRecord(record); ok {
			p.Observe(ctx, metrics)
			return
		}
	}

	// For other metric types, use the base collector
	p.baseCollector().Record(ctx, record)
}

// ObserveHTTP implements the unified MetricsCollector interface
func (p *PrometheusCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	metrics := middleware.RequestMetrics{
		Method:    method,
		Path:      path,
		Status:    status,
		Bytes:     bytes,
		Duration:  duration,
		TraceID:   "",
		UserAgent: "",
	}
	p.Observe(ctx, metrics)
}

// ObservePubSub implements the unified MetricsCollector interface
func (p *PrometheusCollector) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	p.baseCollector().ObservePubSub(ctx, operation, topic, duration, err)
}

// ObserveMQ implements the unified MetricsCollector interface
func (p *PrometheusCollector) ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool) {
	p.baseCollector().ObserveMQ(ctx, operation, topic, duration, err, panicked)
}

// ObserveKV implements the unified MetricsCollector interface
func (p *PrometheusCollector) ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool) {
	p.baseCollector().ObserveKV(ctx, operation, key, duration, err, hit)
}

// ObserveIPC implements the unified MetricsCollector interface
func (p *PrometheusCollector) ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error) {
	p.baseCollector().ObserveIPC(ctx, operation, addr, transport, bytes, duration, err)
}

func (p *PrometheusCollector) ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
	p.baseCollector().ObserveDB(ctx, operation, driver, query, rows, duration, err)
}

func (p *PrometheusCollector) snapshot() (map[labelKey]uint64, map[labelKey]latencyStats, time.Duration) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	reqCopy := make(map[labelKey]uint64, len(p.requests))
	for k, v := range p.requests {
		reqCopy[k] = v
	}

	durCopy := make(map[labelKey]latencyStats, len(p.durations))
	for k, v := range p.durations {
		durCopy[k] = v
	}

	uptime := time.Since(p.startTime)
	return reqCopy, durCopy, uptime
}

func (p *PrometheusCollector) evictOldest() {
	// Simple eviction: remove 10% of least-used entries
	evictCount := len(p.requests) / 10
	if evictCount == 0 {
		evictCount = 1
	}

	// Collect all keys
	keys := make([]labelKey, 0, len(p.requests))
	for k := range p.requests {
		keys = append(keys, k)
	}

	// Sort by request count (evict least used first)
	sort.Slice(keys, func(i, j int) bool {
		return p.requests[keys[i]] < p.requests[keys[j]]
	})

	// Remove oldest
	for i := 0; i < evictCount && i < len(keys); i++ {
		delete(p.requests, keys[i])
		delete(p.durations, keys[i])
	}
}

func (p *PrometheusCollector) baseCollector() *BaseMetricsCollector {
	p.baseOnce.Do(func() {
		p.base = NewBaseMetricsCollector()
	})
	return p.base
}

func httpMetricsFromRecord(record MetricRecord) (middleware.RequestMetrics, bool) {
	if record.Labels == nil {
		return middleware.RequestMetrics{}, false
	}

	statusStr, ok := record.Labels[labelStatus]
	if !ok {
		return middleware.RequestMetrics{}, false
	}
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		return middleware.RequestMetrics{}, false
	}
	method, ok := record.Labels[labelMethod]
	if !ok {
		return middleware.RequestMetrics{}, false
	}
	path, ok := record.Labels[labelPath]
	if !ok {
		return middleware.RequestMetrics{}, false
	}

	return middleware.RequestMetrics{
		Method:   method,
		Path:     path,
		Status:   status,
		Duration: record.Duration,
	}, true
}

func sortedKeys[T any](m map[labelKey]T) []labelKey {
	keys := make([]labelKey, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		if keys[i].path != keys[j].path {
			return keys[i].path < keys[j].path
		}
		return keys[i].status < keys[j].status
	})
	return keys
}
