package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// SlowQuery represents a slow query execution.
type SlowQuery struct {
	Query     string        `json:"query"`
	Operation string        `json:"operation"`
	Duration  time.Duration `json:"duration"`
	Error     error         `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Driver    string        `json:"driver"`
}

// SlowQueryDetector detects and logs slow database queries.
type SlowQueryDetector struct {
	mu sync.RWMutex

	threshold     time.Duration
	maxRecords    int
	slowQueries   []SlowQuery
	totalDetected int64

	// Callbacks
	onSlowQuery func(SlowQuery)
}

// SlowQueryDetectorOption configures the slow query detector.
type SlowQueryDetectorOption func(*SlowQueryDetector)

// WithSlowQueryThreshold sets the threshold for slow queries.
func WithSlowQueryThreshold(threshold time.Duration) SlowQueryDetectorOption {
	return func(d *SlowQueryDetector) {
		d.threshold = threshold
	}
}

// WithSlowQueryMaxRecords sets the maximum number of slow queries to keep.
func WithSlowQueryMaxRecords(max int) SlowQueryDetectorOption {
	return func(d *SlowQueryDetector) {
		d.maxRecords = max
	}
}

// WithSlowQueryCallback sets a callback to be called when a slow query is detected.
func WithSlowQueryCallback(callback func(SlowQuery)) SlowQueryDetectorOption {
	return func(d *SlowQueryDetector) {
		d.onSlowQuery = callback
	}
}

// NewSlowQueryDetector creates a new slow query detector.
func NewSlowQueryDetector(opts ...SlowQueryDetectorOption) *SlowQueryDetector {
	d := &SlowQueryDetector{
		threshold:   1 * time.Second, // default threshold
		maxRecords:  100,             // default max records
		slowQueries: make([]SlowQuery, 0, 100),
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Check checks if a query is slow and records it if so.
// Returns true if the query was slow.
func (d *SlowQueryDetector) Check(operation, driver, query string, duration time.Duration, err error) bool {
	if duration < d.threshold {
		return false
	}

	slowQuery := SlowQuery{
		Query:     query,
		Operation: operation,
		Duration:  duration,
		Error:     err,
		Timestamp: time.Now(),
		Driver:    driver,
	}

	d.mu.Lock()
	d.totalDetected++

	// Add to records
	if len(d.slowQueries) >= d.maxRecords {
		// Remove oldest entry (FIFO)
		d.slowQueries = d.slowQueries[1:]
	}
	d.slowQueries = append(d.slowQueries, slowQuery)
	d.mu.Unlock()

	// Call callback if set
	if d.onSlowQuery != nil {
		d.onSlowQuery(slowQuery)
	}

	return true
}

// GetSlowQueries returns all recorded slow queries.
func (d *SlowQueryDetector) GetSlowQueries() []SlowQuery {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]SlowQuery, len(d.slowQueries))
	copy(result, d.slowQueries)
	return result
}

// GetRecentSlowQueries returns the most recent N slow queries.
func (d *SlowQueryDetector) GetRecentSlowQueries(n int) []SlowQuery {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if n > len(d.slowQueries) {
		n = len(d.slowQueries)
	}

	if n == 0 {
		return nil
	}

	// Return last N queries (most recent)
	result := make([]SlowQuery, n)
	start := len(d.slowQueries) - n
	copy(result, d.slowQueries[start:])
	return result
}

// GetTotalDetected returns the total number of slow queries detected.
func (d *SlowQueryDetector) GetTotalDetected() int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.totalDetected
}

// Clear clears all recorded slow queries.
func (d *SlowQueryDetector) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.slowQueries = make([]SlowQuery, 0, d.maxRecords)
	d.totalDetected = 0
}

// Summary returns a formatted summary of slow queries.
func (d *SlowQueryDetector) Summary() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.totalDetected == 0 {
		return "No slow queries detected"
	}

	summary := fmt.Sprintf("Slow Query Summary (threshold: %s)\n", d.threshold)
	summary += fmt.Sprintf("Total detected: %d\n", d.totalDetected)
	summary += fmt.Sprintf("Currently stored: %d\n\n", len(d.slowQueries))

	if len(d.slowQueries) > 0 {
		summary += "Recent slow queries:\n"
		// Show up to 10 most recent
		start := 0
		if len(d.slowQueries) > 10 {
			start = len(d.slowQueries) - 10
		}
		for i := start; i < len(d.slowQueries); i++ {
			sq := d.slowQueries[i]
			summary += fmt.Sprintf("  [%s] %s (%s) - %s\n",
				sq.Timestamp.Format("15:04:05"),
				sq.Operation,
				sq.Duration.Round(time.Millisecond),
				truncateQuery(sq.Query, 60))
			if sq.Error != nil {
				summary += fmt.Sprintf("    Error: %v\n", sq.Error)
			}
		}
	}

	return summary
}

func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}

// InstrumentedDBWithSlowQueryDetection wraps InstrumentedDB with slow query detection.
type InstrumentedDBWithSlowQueryDetection struct {
	*InstrumentedDB
	detector *SlowQueryDetector
}

// NewInstrumentedDBWithSlowQueryDetection creates an instrumented DB with slow query detection.
func NewInstrumentedDBWithSlowQueryDetection(
	db *InstrumentedDB,
	opts ...SlowQueryDetectorOption,
) *InstrumentedDBWithSlowQueryDetection {
	return &InstrumentedDBWithSlowQueryDetection{
		InstrumentedDB: db,
		detector:       NewSlowQueryDetector(opts...),
	}
}

// GetSlowQueryDetector returns the slow query detector.
func (idb *InstrumentedDBWithSlowQueryDetection) GetSlowQueryDetector() *SlowQueryDetector {
	return idb.detector
}

// ExecContext wraps ExecContext with slow query detection.
func (idb *InstrumentedDBWithSlowQueryDetection) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := idb.InstrumentedDB.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	idb.detector.Check("exec", idb.driver, query, duration, err)

	return result, err
}

// QueryContext wraps QueryContext with slow query detection.
func (idb *InstrumentedDBWithSlowQueryDetection) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := idb.InstrumentedDB.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	idb.detector.Check("query", idb.driver, query, duration, err)

	return rows, err
}

// QueryRowContext wraps QueryRowContext with slow query detection.
func (idb *InstrumentedDBWithSlowQueryDetection) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := idb.InstrumentedDB.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	idb.detector.Check("query", idb.driver, query, duration, nil)

	return row
}

// MetricsCollectorWithSlowQueryDetection wraps a MetricsCollector with slow query detection.
type MetricsCollectorWithSlowQueryDetection struct {
	base     MetricsCollector
	detector *SlowQueryDetector
}

// NewMetricsCollectorWithSlowQueryDetection creates a metrics collector with slow query detection.
func NewMetricsCollectorWithSlowQueryDetection(
	base MetricsCollector,
	opts ...SlowQueryDetectorOption,
) *MetricsCollectorWithSlowQueryDetection {
	return &MetricsCollectorWithSlowQueryDetection{
		base:     base,
		detector: NewSlowQueryDetector(opts...),
	}
}

// ObserveDB implements MetricsCollector and also detects slow queries.
func (c *MetricsCollectorWithSlowQueryDetection) ObserveDB(
	ctx context.Context,
	operation, driver, query string,
	rows int,
	duration time.Duration,
	err error,
) {
	// Forward to base collector
	if c.base != nil {
		c.base.ObserveDB(ctx, operation, driver, query, rows, duration, err)
	}

	// Check for slow query
	c.detector.Check(operation, driver, query, duration, err)
}

// GetSlowQueryDetector returns the slow query detector.
func (c *MetricsCollectorWithSlowQueryDetection) GetSlowQueryDetector() *SlowQueryDetector {
	return c.detector
}
