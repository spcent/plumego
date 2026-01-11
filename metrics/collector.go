package metrics

import (
	"context"
	"time"
)

// MetricType represents the type of metric being recorded
type MetricType string

const (
	// HTTP metrics
	MetricHTTPRequest MetricType = "http_request"

	// PubSub metrics
	MetricPubSubPublish   MetricType = "pubsub_publish"
	MetricPubSubSubscribe MetricType = "pubsub_subscribe"
	MetricPubSubDeliver   MetricType = "pubsub_deliver"
	MetricPubSubDrop      MetricType = "pubsub_drop"

	// Message Queue metrics
	MetricMQPublish   MetricType = "mq_publish"
	MetricMQSubscribe MetricType = "mq_subscribe"
	MetricMQClose     MetricType = "mq_close"
	MetricMQMetrics   MetricType = "mq_metrics"

	// Key-Value Store metrics
	MetricKVSet    MetricType = "kv_set"
	MetricKVGet    MetricType = "kv_get"
	MetricKVDelete MetricType = "kv_delete"
	MetricKVExists MetricType = "kv_exists"
	MetricKVKeys   MetricType = "kv_keys"
	MetricKVHit    MetricType = "kv_hit"
	MetricKVMiss   MetricType = "kv_miss"
	MetricKVEvict  MetricType = "kv_evict"
)

// MetricLabels represents key-value labels for metrics
type MetricLabels map[string]string

// MetricRecord represents a single metric record
type MetricRecord struct {
	Type      MetricType
	Name      string
	Value     float64
	Labels    MetricLabels
	Timestamp time.Time
	Duration  time.Duration
	Error     error
}

// MetricsCollector is the unified interface for collecting metrics across all modules
type MetricsCollector interface {
	// Record records a single metric
	Record(ctx context.Context, record MetricRecord)

	// ObserveHTTP is a convenience method for HTTP request metrics
	ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration)

	// ObservePubSub is a convenience method for PubSub metrics
	ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error)

	// ObserveMQ is a convenience method for Message Queue metrics
	ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool)

	// ObserveKV is a convenience method for Key-Value Store metrics
	ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool)

	// GetStats returns current statistics
	GetStats() CollectorStats

	// Clear resets all metrics
	Clear()
}

// CollectorStats provides statistics about the collector
type CollectorStats struct {
	TotalRecords  int64                `json:"total_records"`
	ErrorRecords  int64                `json:"error_records"`
	ActiveSeries  int                  `json:"active_series"`
	StartTime     time.Time            `json:"start_time"`
	TypeBreakdown map[MetricType]int64 `json:"type_breakdown"`
	// Legacy fields for backward compatibility
	Series         int     `json:"series"`
	TotalRequests  uint64  `json:"total_requests"`
	AverageLatency float64 `json:"average_latency"`
	// Tracing-specific fields
	TotalSpans      int           `json:"total_spans"`
	ErrorSpans      int           `json:"error_spans"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
}

// BaseMetricsCollector provides a base implementation for metrics collectors
type BaseMetricsCollector struct {
	records []MetricRecord
	stats   CollectorStats
}

// NewBaseMetricsCollector creates a new base metrics collector
func NewBaseMetricsCollector() *BaseMetricsCollector {
	return &BaseMetricsCollector{
		records: make([]MetricRecord, 0, 1000),
		stats: CollectorStats{
			StartTime:     time.Now(),
			TypeBreakdown: make(map[MetricType]int64),
		},
	}
}

// Record implements the MetricsCollector interface
func (b *BaseMetricsCollector) Record(ctx context.Context, record MetricRecord) {
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	b.records = append(b.records, record)
	b.stats.TotalRecords++
	b.stats.TypeBreakdown[record.Type]++

	if record.Error != nil {
		b.stats.ErrorRecords++
	}
}

// ObserveHTTP implements HTTP metrics recording
func (b *BaseMetricsCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	record := MetricRecord{
		Type:  MetricHTTPRequest,
		Name:  "http_request",
		Value: float64(duration.Milliseconds()),
		Labels: MetricLabels{
			"method": method,
			"path":   path,
			"status": string(rune(status)),
		},
		Duration: duration,
	}
	b.Record(ctx, record)
}

// ObservePubSub implements PubSub metrics recording
func (b *BaseMetricsCollector) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	var metricType MetricType
	switch operation {
	case "publish":
		metricType = MetricPubSubPublish
	case "subscribe":
		metricType = MetricPubSubSubscribe
	case "deliver":
		metricType = MetricPubSubDeliver
	case "drop":
		metricType = MetricPubSubDrop
	default:
		metricType = MetricPubSubPublish
	}

	record := MetricRecord{
		Type:  metricType,
		Name:  "pubsub_" + operation,
		Value: float64(duration.Milliseconds()),
		Labels: MetricLabels{
			"operation": operation,
			"topic":     topic,
		},
		Duration: duration,
		Error:    err,
	}
	b.Record(ctx, record)
}

// ObserveMQ implements Message Queue metrics recording
func (b *BaseMetricsCollector) ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool) {
	var metricType MetricType
	switch operation {
	case "publish":
		metricType = MetricMQPublish
	case "subscribe":
		metricType = MetricMQSubscribe
	case "close":
		metricType = MetricMQClose
	case "metrics":
		metricType = MetricMQMetrics
	default:
		metricType = MetricMQPublish
	}

	record := MetricRecord{
		Type:  metricType,
		Name:  "mq_" + operation,
		Value: float64(duration.Milliseconds()),
		Labels: MetricLabels{
			"operation": operation,
			"topic":     topic,
			"panicked":  map[bool]string{true: "true", false: "false"}[panicked],
		},
		Duration: duration,
		Error:    err,
	}
	b.Record(ctx, record)
}

// ObserveKV implements Key-Value Store metrics recording
func (b *BaseMetricsCollector) ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool) {
	var metricType MetricType
	switch operation {
	case "set":
		metricType = MetricKVSet
	case "get":
		metricType = MetricKVGet
	case "delete":
		metricType = MetricKVDelete
	case "exists":
		metricType = MetricKVExists
	case "keys":
		metricType = MetricKVKeys
	default:
		metricType = MetricKVGet
	}

	labels := MetricLabels{
		"operation": operation,
		"hit":       map[bool]string{true: "true", false: "false"}[hit],
	}
	if key != "" {
		labels["key"] = key
	}

	record := MetricRecord{
		Type:     metricType,
		Name:     "kv_" + operation,
		Value:    float64(duration.Milliseconds()),
		Labels:   labels,
		Duration: duration,
		Error:    err,
	}
	b.Record(ctx, record)

	// Also record hit/miss specific metrics
	if operation == "get" {
		var hitType MetricType
		if hit {
			hitType = MetricKVHit
		} else {
			hitType = MetricKVMiss
		}
		hitRecord := MetricRecord{
			Type:     hitType,
			Name:     "kv_" + operation + "_" + map[bool]string{true: "hit", false: "miss"}[hit],
			Value:    float64(duration.Milliseconds()),
			Labels:   labels,
			Duration: duration,
			Error:    err,
		}
		b.Record(ctx, hitRecord)
	}
}

// GetStats returns current statistics
func (b *BaseMetricsCollector) GetStats() CollectorStats {
	return b.stats
}

// Clear resets all metrics
func (b *BaseMetricsCollector) Clear() {
	b.records = b.records[:0]
	b.stats = CollectorStats{
		StartTime:     time.Now(),
		TypeBreakdown: make(map[MetricType]int64),
	}
}

// GetRecords returns a copy of all records (for testing or advanced collectors)
func (b *BaseMetricsCollector) GetRecords() []MetricRecord {
	result := make([]MetricRecord, len(b.records))
	copy(result, b.records)
	return result
}
