package sharding

import (
	"context"
	"sync"
	"time"
)

// MetricsCollector collects metrics for database sharding operations
type MetricsCollector struct {
	mu sync.RWMutex

	// Query metrics
	totalQueries       uint64
	singleShardQueries uint64
	crossShardQueries  uint64
	failedQueries      uint64

	// Per-shard metrics
	shardQueryCounts []uint64
	shardErrorCounts []uint64

	// Query type metrics
	selectQueries uint64
	insertQueries uint64
	updateQueries uint64
	deleteQueries uint64

	// Latency metrics (in microseconds)
	totalLatency   uint64
	minLatency     uint64
	maxLatency     uint64
	latencyBuckets map[string]uint64 // latency histogram buckets
	latencyP50     uint64
	latencyP95     uint64
	latencyP99     uint64

	// Rewrite metrics
	rewriteCount  uint64
	cacheHits     uint64
	cacheMisses   uint64
	rewriteErrors uint64

	// Enabled flag
	enabled bool
}

// MetricsSnapshot is a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	// Query metrics
	TotalQueries       uint64 `json:"total_queries"`
	SingleShardQueries uint64 `json:"single_shard_queries"`
	CrossShardQueries  uint64 `json:"cross_shard_queries"`
	FailedQueries      uint64 `json:"failed_queries"`

	// Per-shard metrics
	ShardQueryCounts []uint64 `json:"shard_query_counts"`
	ShardErrorCounts []uint64 `json:"shard_error_counts"`

	// Query type metrics
	SelectQueries uint64 `json:"select_queries"`
	InsertQueries uint64 `json:"insert_queries"`
	UpdateQueries uint64 `json:"update_queries"`
	DeleteQueries uint64 `json:"delete_queries"`

	// Latency metrics (in microseconds)
	TotalLatency   uint64            `json:"total_latency_us"`
	MinLatency     uint64            `json:"min_latency_us"`
	MaxLatency     uint64            `json:"max_latency_us"`
	AvgLatency     uint64            `json:"avg_latency_us"`
	LatencyBuckets map[string]uint64 `json:"latency_buckets"`
	LatencyP50     uint64            `json:"latency_p50_us"`
	LatencyP95     uint64            `json:"latency_p95_us"`
	LatencyP99     uint64            `json:"latency_p99_us"`

	// Rewrite metrics
	RewriteCount  uint64  `json:"rewrite_count"`
	CacheHits     uint64  `json:"cache_hits"`
	CacheMisses   uint64  `json:"cache_misses"`
	CacheHitRate  float64 `json:"cache_hit_rate"`
	RewriteErrors uint64  `json:"rewrite_errors"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(shardCount int) *MetricsCollector {
	return &MetricsCollector{
		shardQueryCounts: make([]uint64, shardCount),
		shardErrorCounts: make([]uint64, shardCount),
		latencyBuckets:   make(map[string]uint64),
		enabled:          true,
	}
}

// RecordQuery records a query execution
func (m *MetricsCollector) RecordQuery(ctx context.Context, queryType string, shardIndex int, latency time.Duration, err error) {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalQueries++

	// Record query type
	switch queryType {
	case "SELECT":
		m.selectQueries++
	case "INSERT":
		m.insertQueries++
	case "UPDATE":
		m.updateQueries++
	case "DELETE":
		m.deleteQueries++
	}

	// Record shard metrics
	if shardIndex >= 0 && shardIndex < len(m.shardQueryCounts) {
		m.shardQueryCounts[shardIndex]++
		if err != nil {
			m.shardErrorCounts[shardIndex]++
		}
	}

	// Record errors
	if err != nil {
		m.failedQueries++
	}

	// Record latency
	latencyUs := uint64(latency.Microseconds())
	m.totalLatency += latencyUs

	if m.minLatency == 0 || latencyUs < m.minLatency {
		m.minLatency = latencyUs
	}
	if latencyUs > m.maxLatency {
		m.maxLatency = latencyUs
	}

	// Record in buckets (histogram)
	m.recordLatencyBucket(latencyUs)
}

// RecordSingleShardQuery records a single-shard query
func (m *MetricsCollector) RecordSingleShardQuery() {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.singleShardQueries++
}

// RecordCrossShardQuery records a cross-shard query
func (m *MetricsCollector) RecordCrossShardQuery() {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.crossShardQueries++
}

// RecordRewrite records a SQL rewrite operation
func (m *MetricsCollector) RecordRewrite(cached bool, err error) {
	if !m.enabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rewriteCount++

	if cached {
		m.cacheHits++
	} else {
		m.cacheMisses++
	}

	if err != nil {
		m.rewriteErrors++
	}
}

// recordLatencyBucket records latency in histogram buckets
func (m *MetricsCollector) recordLatencyBucket(latencyUs uint64) {
	// Define buckets (in microseconds)
	buckets := []uint64{
		100,     // 0.1ms
		500,     // 0.5ms
		1000,    // 1ms
		5000,    // 5ms
		10000,   // 10ms
		50000,   // 50ms
		100000,  // 100ms
		500000,  // 500ms
		1000000, // 1s
	}

	for _, bucket := range buckets {
		if latencyUs <= bucket {
			key := formatLatencyBucket(bucket)
			m.latencyBuckets[key]++
			return
		}
	}

	// > 1s
	m.latencyBuckets[">1s"]++
}

// formatLatencyBucket formats a bucket value for display
func formatLatencyBucket(bucketUs uint64) string {
	if bucketUs < 1000 {
		return "<0.1ms"
	}
	ms := bucketUs / 1000
	if ms < 1 {
		return "<1ms"
	}
	if ms < 10 {
		return "<10ms"
	}
	if ms < 100 {
		return "<100ms"
	}
	if ms < 1000 {
		return "<1s"
	}
	return ">1s"
}

// Snapshot returns a point-in-time snapshot of metrics
func (m *MetricsCollector) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate average latency
	var avgLatency uint64
	if m.totalQueries > 0 {
		avgLatency = m.totalLatency / m.totalQueries
	}

	// Calculate cache hit rate
	var cacheHitRate float64
	totalCacheOps := m.cacheHits + m.cacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(m.cacheHits) / float64(totalCacheOps) * 100
	}

	// Copy shard counts
	shardQueryCounts := make([]uint64, len(m.shardQueryCounts))
	copy(shardQueryCounts, m.shardQueryCounts)

	shardErrorCounts := make([]uint64, len(m.shardErrorCounts))
	copy(shardErrorCounts, m.shardErrorCounts)

	// Copy latency buckets
	latencyBuckets := make(map[string]uint64, len(m.latencyBuckets))
	for k, v := range m.latencyBuckets {
		latencyBuckets[k] = v
	}

	return MetricsSnapshot{
		TotalQueries:       m.totalQueries,
		SingleShardQueries: m.singleShardQueries,
		CrossShardQueries:  m.crossShardQueries,
		FailedQueries:      m.failedQueries,
		ShardQueryCounts:   shardQueryCounts,
		ShardErrorCounts:   shardErrorCounts,
		SelectQueries:      m.selectQueries,
		InsertQueries:      m.insertQueries,
		UpdateQueries:      m.updateQueries,
		DeleteQueries:      m.deleteQueries,
		TotalLatency:       m.totalLatency,
		MinLatency:         m.minLatency,
		MaxLatency:         m.maxLatency,
		AvgLatency:         avgLatency,
		LatencyBuckets:     latencyBuckets,
		LatencyP50:         m.latencyP50,
		LatencyP95:         m.latencyP95,
		LatencyP99:         m.latencyP99,
		RewriteCount:       m.rewriteCount,
		CacheHits:          m.cacheHits,
		CacheMisses:        m.cacheMisses,
		CacheHitRate:       cacheHitRate,
		RewriteErrors:      m.rewriteErrors,
		Timestamp:          time.Now(),
	}
}

// Reset resets all metrics to zero
func (m *MetricsCollector) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalQueries = 0
	m.singleShardQueries = 0
	m.crossShardQueries = 0
	m.failedQueries = 0

	for i := range m.shardQueryCounts {
		m.shardQueryCounts[i] = 0
		m.shardErrorCounts[i] = 0
	}

	m.selectQueries = 0
	m.insertQueries = 0
	m.updateQueries = 0
	m.deleteQueries = 0

	m.totalLatency = 0
	m.minLatency = 0
	m.maxLatency = 0

	m.latencyBuckets = make(map[string]uint64)

	m.rewriteCount = 0
	m.cacheHits = 0
	m.cacheMisses = 0
	m.rewriteErrors = 0
}

// Enable enables metrics collection
func (m *MetricsCollector) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables metrics collection
func (m *MetricsCollector) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// IsEnabled returns whether metrics collection is enabled
func (m *MetricsCollector) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// PrometheusMetrics returns metrics in Prometheus text format
func (m *MetricsCollector) PrometheusMetrics() string {
	snapshot := m.Snapshot()

	metrics := ""

	// Query counters
	metrics += "# HELP db_sharding_queries_total Total number of queries\n"
	metrics += "# TYPE db_sharding_queries_total counter\n"
	metrics += formatPrometheusMetric("db_sharding_queries_total", snapshot.TotalQueries)

	metrics += "# HELP db_sharding_single_shard_queries_total Single-shard queries\n"
	metrics += "# TYPE db_sharding_single_shard_queries_total counter\n"
	metrics += formatPrometheusMetric("db_sharding_single_shard_queries_total", snapshot.SingleShardQueries)

	metrics += "# HELP db_sharding_cross_shard_queries_total Cross-shard queries\n"
	metrics += "# TYPE db_sharding_cross_shard_queries_total counter\n"
	metrics += formatPrometheusMetric("db_sharding_cross_shard_queries_total", snapshot.CrossShardQueries)

	metrics += "# HELP db_sharding_failed_queries_total Failed queries\n"
	metrics += "# TYPE db_sharding_failed_queries_total counter\n"
	metrics += formatPrometheusMetric("db_sharding_failed_queries_total", snapshot.FailedQueries)

	// Per-shard metrics
	for i, count := range snapshot.ShardQueryCounts {
		metrics += formatPrometheusMetricWithLabels("db_sharding_shard_queries_total", count, map[string]string{"shard": formatInt(i)})
	}

	for i, count := range snapshot.ShardErrorCounts {
		metrics += formatPrometheusMetricWithLabels("db_sharding_shard_errors_total", count, map[string]string{"shard": formatInt(i)})
	}

	// Query type metrics
	metrics += formatPrometheusMetricWithLabels("db_sharding_query_type_total", snapshot.SelectQueries, map[string]string{"type": "SELECT"})
	metrics += formatPrometheusMetricWithLabels("db_sharding_query_type_total", snapshot.InsertQueries, map[string]string{"type": "INSERT"})
	metrics += formatPrometheusMetricWithLabels("db_sharding_query_type_total", snapshot.UpdateQueries, map[string]string{"type": "UPDATE"})
	metrics += formatPrometheusMetricWithLabels("db_sharding_query_type_total", snapshot.DeleteQueries, map[string]string{"type": "DELETE"})

	// Latency metrics
	metrics += "# HELP db_sharding_latency_microseconds Query latency in microseconds\n"
	metrics += "# TYPE db_sharding_latency_microseconds summary\n"
	metrics += formatPrometheusMetricWithLabels("db_sharding_latency_microseconds", snapshot.MinLatency, map[string]string{"quantile": "min"})
	metrics += formatPrometheusMetricWithLabels("db_sharding_latency_microseconds", snapshot.AvgLatency, map[string]string{"quantile": "avg"})
	metrics += formatPrometheusMetricWithLabels("db_sharding_latency_microseconds", snapshot.MaxLatency, map[string]string{"quantile": "max"})

	// Cache metrics
	metrics += "# HELP db_sharding_cache_hits_total Cache hits\n"
	metrics += "# TYPE db_sharding_cache_hits_total counter\n"
	metrics += formatPrometheusMetric("db_sharding_cache_hits_total", snapshot.CacheHits)

	metrics += "# HELP db_sharding_cache_misses_total Cache misses\n"
	metrics += "# TYPE db_sharding_cache_misses_total counter\n"
	metrics += formatPrometheusMetric("db_sharding_cache_misses_total", snapshot.CacheMisses)

	return metrics
}

func formatPrometheusMetric(name string, value uint64) string {
	return name + " " + formatUint64(value) + "\n"
}

func formatPrometheusMetricWithLabels(name string, value uint64, labels map[string]string) string {
	labelsStr := ""
	for k, v := range labels {
		if labelsStr != "" {
			labelsStr += ","
		}
		labelsStr += k + `="` + v + `"`
	}

	if labelsStr != "" {
		return name + "{" + labelsStr + "} " + formatUint64(value) + "\n"
	}
	return name + " " + formatUint64(value) + "\n"
}

func formatInt(n int) string {
	return formatUint64(uint64(n))
}

func formatUint64(n uint64) string {
	if n == 0 {
		return "0"
	}

	var buf [20]byte
	i := len(buf) - 1

	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}

	return string(buf[i+1:])
}
