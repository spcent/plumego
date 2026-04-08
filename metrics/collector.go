package metrics

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MetricType represents the type of metric being recorded
type MetricType string

const (
	// HTTP metrics
	MetricHTTPRequest MetricType = "http_request"
)

// MetricLabels represents key-value labels for metrics
type MetricLabels map[string]string

const (
	labelMethod    = "method"
	labelPath      = "path"
	labelStatus    = "status"
	labelOperation = "operation"
	labelTopic     = "topic"
	labelKVKey     = "key"
	labelPanicked  = "panicked"
	labelHit       = "hit"
	labelAddr      = "addr"
	labelTransport = "transport"
	labelBytes     = "bytes"
	labelDriver    = "driver"
	labelQuery     = "query"
	labelTable     = "table"
	labelRows      = "rows"
)

// MetricRecord represents a single metric record
type MetricRecord struct {
	// Type is optional. Stable metrics owns only the shared HTTP type; owner-
	// specific or extension-specific taxonomy should be set by the owning package.
	Type MetricType
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

// AggregateCollector is the full collector surface.
// Prefer narrower interfaces at module boundaries and reserve this contract for
// collector implementations or fan-out adapters that truly need the whole set.
type AggregateCollector interface {
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

	// ObserveIPC is a convenience method for IPC metrics
	ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error)

	// ObserveDB is a convenience method for database metrics
	ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error)

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

// PubSubObserver captures only pub/sub activity metrics.
type PubSubObserver interface {
	ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error)
}

// MQObserver captures only message queue activity metrics.
type MQObserver interface {
	ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool)
}

// KVObserver captures only key-value store activity metrics.
type KVObserver interface {
	ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool)
}

// IPCObserver captures only IPC activity metrics.
type IPCObserver interface {
	ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error)
}

// DBObserver captures only database activity metrics.
type DBObserver interface {
	ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error)
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
//   - TypeBreakdown: per-metric-type counters when the collector can provide them.
type CollectorStats struct {
	TotalRecords  int64                `json:"total_records"`
	ErrorRecords  int64                `json:"error_records"`
	ActiveSeries  int                  `json:"active_series"`
	StartTime     time.Time            `json:"start_time"`
	TypeBreakdown map[MetricType]int64 `json:"type_breakdown"`
}

// BaseMetricsCollector provides a base implementation for metrics collectors
type BaseMetricsCollector struct {
	mu         sync.RWMutex
	records    []MetricRecord
	stats      CollectorStats
	maxRecords int
}

const defaultMaxRecords = 10000

// NewBaseMetricsCollector creates a new base metrics collector.
// It retains up to defaultMaxRecords records unless overridden by WithMaxRecords.
func NewBaseMetricsCollector() *BaseMetricsCollector {
	return &BaseMetricsCollector{
		records: make([]MetricRecord, 0, 1000),
		stats: CollectorStats{
			StartTime:     time.Now(),
			TypeBreakdown: make(map[MetricType]int64),
		},
		maxRecords: defaultMaxRecords,
	}
}

// WithMaxRecords limits how many records are retained in memory.
// A non-positive value disables the limit.
func (b *BaseMetricsCollector) WithMaxRecords(max int) *BaseMetricsCollector {
	b.mu.Lock()
	defer b.mu.Unlock()

	if max <= 0 {
		b.maxRecords = 0
		return b
	}

	b.maxRecords = max
	if len(b.records) > max {
		b.records = b.records[len(b.records)-max:]
	}
	return b
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

	if b.maxRecords > 0 && len(b.records) >= b.maxRecords {
		b.records = b.records[1:]
	}

	b.records = append(b.records, record)
	b.stats.TotalRecords++
	if record.Type != "" {
		b.stats.TypeBreakdown[record.Type]++
	}

	if record.Error != nil {
		b.stats.ErrorRecords++
	}
}

// ObserveHTTP implements HTTP metrics recording
func (b *BaseMetricsCollector) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	record := MetricRecord{
		Type:  MetricHTTPRequest,
		Name:  "http_request",
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

// ObservePubSub implements PubSub metrics recording
func (b *BaseMetricsCollector) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	normalizedOperation := normalizeMetricOperation(operation)

	record := MetricRecord{
		Name:  "pubsub_" + normalizedOperation,
		Value: durationValueSeconds(duration),
		Labels: MetricLabels{
			labelOperation: normalizedOperation,
			labelTopic:     topic,
		},
		Duration: duration,
		Error:    err,
	}
	b.Record(ctx, record)
}

// ObserveMQ implements Message Queue metrics recording
func (b *BaseMetricsCollector) ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool) {
	normalizedOperation := normalizeMetricOperation(operation)

	record := MetricRecord{
		Name:  "mq_" + normalizedOperation,
		Value: durationValueSeconds(duration),
		Labels: MetricLabels{
			labelOperation: normalizedOperation,
			labelTopic:     topic,
			labelPanicked:  boolLabel(panicked),
		},
		Duration: duration,
		Error:    err,
	}
	b.Record(ctx, record)
}

// ObserveKV implements Key-Value Store metrics recording
func (b *BaseMetricsCollector) ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool) {
	normalizedOperation := normalizeMetricOperation(operation)

	labels := MetricLabels{
		labelOperation: normalizedOperation,
		labelHit:       boolLabel(hit),
	}
	if key != "" {
		labels[labelKVKey] = key
	}

	record := MetricRecord{
		Name:     "kv_" + normalizedOperation,
		Value:    durationValueSeconds(duration),
		Labels:   labels,
		Duration: duration,
		Error:    err,
	}
	b.Record(ctx, record)
}

// ObserveIPC implements IPC metrics recording
func (b *BaseMetricsCollector) ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error) {
	normalizedOperation := normalizeMetricOperation(operation)

	labels := MetricLabels{
		labelOperation: normalizedOperation,
		labelTransport: transport,
	}
	if addr != "" {
		labels[labelAddr] = addr
	}
	if bytes > 0 {
		labels[labelBytes] = fmt.Sprintf("%d", bytes)
	}

	record := MetricRecord{
		Name:     "ipc_" + normalizedOperation,
		Value:    durationValueSeconds(duration),
		Labels:   labels,
		Duration: duration,
		Error:    err,
	}
	b.Record(ctx, record)
}

// ObserveDB implements database metrics recording
func (b *BaseMetricsCollector) ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
	normalizedOperation := normalizeMetricOperation(operation)

	labels := MetricLabels{
		labelOperation: normalizedOperation,
	}
	if driver != "" {
		labels[labelDriver] = driver
	}
	if query != "" {
		// Truncate long queries to avoid excessive label cardinality
		maxQueryLen := 100
		if len(query) > maxQueryLen {
			labels[labelQuery] = query[:maxQueryLen] + "..."
		} else {
			labels[labelQuery] = query
		}
	}
	if rows > 0 {
		labels[labelRows] = fmt.Sprintf("%d", rows)
	}
	if table := extractTable(query); table != "" {
		labels[labelTable] = table
	}

	record := MetricRecord{
		Name:     "db_" + normalizedOperation,
		Value:    durationValueSeconds(duration),
		Labels:   labels,
		Duration: duration,
		Error:    err,
	}
	b.Record(ctx, record)
}

// GetStats returns current statistics
func (b *BaseMetricsCollector) GetStats() CollectorStats {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stats := b.stats
	if stats.ActiveSeries == 0 && len(stats.TypeBreakdown) > 0 {
		stats.ActiveSeries = len(stats.TypeBreakdown)
	}
	if stats.TypeBreakdown != nil {
		stats.TypeBreakdown = cloneBreakdown(stats.TypeBreakdown)
	}
	return stats
}

// Clear resets all metrics
func (b *BaseMetricsCollector) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ensureInitializedLocked()

	b.records = b.records[:0]
	b.stats = CollectorStats{
		StartTime:     time.Now(),
		TypeBreakdown: make(map[MetricType]int64),
	}
}

// GetRecords returns a copy of all records (for testing or advanced collectors)
func (b *BaseMetricsCollector) GetRecords() []MetricRecord {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]MetricRecord, len(b.records))
	for i, record := range b.records {
		if len(record.Labels) > 0 {
			record.Labels = cloneLabels(record.Labels)
		}
		result[i] = record
	}
	return result
}

func (b *BaseMetricsCollector) ensureInitializedLocked() {
	if b.stats.TypeBreakdown == nil {
		b.stats.TypeBreakdown = make(map[MetricType]int64)
	}
	if b.stats.StartTime.IsZero() {
		b.stats.StartTime = time.Now()
	}
	if b.records == nil {
		b.records = make([]MetricRecord, 0, 1000)
	}
}

func boolLabel(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func normalizeMetricOperation(operation string) string {
	normalized := strings.ToLower(strings.TrimSpace(operation))
	if normalized == "" {
		return "unknown"
	}
	return normalized
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

func cloneBreakdown(breakdown map[MetricType]int64) map[MetricType]int64 {
	if breakdown == nil {
		return nil
	}
	result := make(map[MetricType]int64, len(breakdown))
	for key, value := range breakdown {
		result[key] = value
	}
	return result
}

// extractTable attempts to extract the primary table name from a SQL query.
// It handles common patterns: SELECT ... FROM, INSERT INTO, UPDATE, DELETE FROM,
// CREATE/ALTER/DROP/TRUNCATE TABLE, REPLACE INTO, and MERGE INTO.
// Returns an empty string if no table name can be determined.
func extractTable(query string) string {
	fields := strings.Fields(query)
	if len(fields) == 0 {
		return ""
	}

	for i := 0; i < len(fields); i++ {
		upper := strings.ToUpper(fields[i])
		switch upper {
		case "FROM", "INTO":
			// SELECT ... FROM table, INSERT INTO table, DELETE FROM table,
			// MERGE INTO table, REPLACE INTO table
			if i+1 < len(fields) {
				next := fields[i+1]
				// Skip subqueries
				if strings.HasPrefix(next, "(") {
					continue
				}
				return cleanTableName(next)
			}
		case "UPDATE":
			// UPDATE table SET ...
			if i+1 < len(fields) {
				return cleanTableName(fields[i+1])
			}
		case "TABLE":
			// CREATE TABLE, ALTER TABLE, DROP TABLE, TRUNCATE TABLE
			next := i + 1
			// Skip IF [NOT] EXISTS
			if next < len(fields) && strings.EqualFold(fields[next], "IF") {
				next++
				if next < len(fields) && strings.EqualFold(fields[next], "NOT") {
					next++
				}
				if next < len(fields) && strings.EqualFold(fields[next], "EXISTS") {
					next++
				}
			}
			if next < len(fields) {
				return cleanTableName(fields[next])
			}
		}
	}

	return ""
}

// cleanTableName removes surrounding quotes, trailing punctuation, and
// extracts the table part from schema-qualified identifiers (schema.table).
func cleanTableName(s string) string {
	// Remove trailing punctuation (comma, semicolon, parenthesis)
	s = strings.TrimRight(s, ",;()")
	if s == "" {
		return ""
	}

	s = unquoteIdent(s)

	// Handle schema.table -> take only the table part
	if idx := strings.LastIndexByte(s, '.'); idx >= 0 && idx+1 < len(s) {
		s = unquoteIdent(s[idx+1:])
	}

	return s
}

// unquoteIdent removes surrounding quote characters from an identifier.
func unquoteIdent(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') ||
			(first == '`' && last == '`') ||
			(first == '[' && last == ']') {
			s = s[1 : len(s)-1]
		}
	}
	return s
}

// baseForwarder provides lazy-initialized forwarding of aggregate collector methods
// to an underlying BaseMetricsCollector. Embed this in collectors that delegate
// common observation methods (PubSub, MQ, KV, IPC, DB) to the base implementation.
type baseForwarder struct {
	base     *BaseMetricsCollector
	baseOnce sync.Once
}

func (f *baseForwarder) getBase() *BaseMetricsCollector {
	f.baseOnce.Do(func() {
		f.base = NewBaseMetricsCollector()
	})
	return f.base
}

func (f *baseForwarder) clearBase() {
	if f.base != nil {
		f.base.Clear()
	}
}

// Record forwards to the base collector.
func (f *baseForwarder) Record(ctx context.Context, record MetricRecord) {
	f.getBase().Record(ctx, record)
}

// ObserveHTTP forwards to the base collector.
func (f *baseForwarder) ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration) {
	f.getBase().ObserveHTTP(ctx, method, path, status, bytes, duration)
}

// ObservePubSub forwards to the base collector.
func (f *baseForwarder) ObservePubSub(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	f.getBase().ObservePubSub(ctx, operation, topic, duration, err)
}

// ObserveMQ forwards to the base collector.
func (f *baseForwarder) ObserveMQ(ctx context.Context, operation, topic string, duration time.Duration, err error, panicked bool) {
	f.getBase().ObserveMQ(ctx, operation, topic, duration, err, panicked)
}

// ObserveKV forwards to the base collector.
func (f *baseForwarder) ObserveKV(ctx context.Context, operation, key string, duration time.Duration, err error, hit bool) {
	f.getBase().ObserveKV(ctx, operation, key, duration, err, hit)
}

// ObserveIPC forwards to the base collector.
func (f *baseForwarder) ObserveIPC(ctx context.Context, operation, addr, transport string, bytes int, duration time.Duration, err error) {
	f.getBase().ObserveIPC(ctx, operation, addr, transport, bytes, duration, err)
}

// ObserveDB forwards to the base collector.
func (f *baseForwarder) ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
	f.getBase().ObserveDB(ctx, operation, driver, query, rows, duration, err)
}
