package db

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// QueryStats tracks statistics for database queries.
type QueryStats struct {
	// Query metrics
	TotalQueries   int64         `json:"total_queries"`
	SuccessQueries int64         `json:"success_queries"`
	FailedQueries  int64         `json:"failed_queries"`
	TotalDuration  time.Duration `json:"total_duration"`
	MinDuration    time.Duration `json:"min_duration"`
	MaxDuration    time.Duration `json:"max_duration"`
	AvgDuration    time.Duration `json:"avg_duration"`

	// Query type breakdown
	SelectQueries int64 `json:"select_queries"`
	InsertQueries int64 `json:"insert_queries"`
	UpdateQueries int64 `json:"update_queries"`
	DeleteQueries int64 `json:"delete_queries"`

	// Slow query tracking
	SlowQueries     int64         `json:"slow_queries"`
	SlowQueryThresh time.Duration `json:"slow_query_threshold"`

	// Last update
	LastUpdate time.Time `json:"last_update"`
}

// DBStatsAggregator aggregates database statistics from metrics.
// It provides a higher-level view of database performance by aggregating
// metrics across operations, tables, and time windows.
type DBStatsAggregator struct {
	mu sync.RWMutex

	// Global stats
	global QueryStats

	// Per-operation stats
	byOperation map[string]QueryStats

	// Per-table stats (extracted from queries)
	byTable map[string]QueryStats

	// Slow query threshold
	slowQueryThreshold time.Duration

	// Tracking
	startTime time.Time
}

// NewDBStatsAggregator creates a new database statistics aggregator.
func NewDBStatsAggregator(slowQueryThreshold time.Duration) *DBStatsAggregator {
	if slowQueryThreshold == 0 {
		slowQueryThreshold = 1 * time.Second // default
	}

	return &DBStatsAggregator{
		byOperation:        make(map[string]QueryStats),
		byTable:            make(map[string]QueryStats),
		slowQueryThreshold: slowQueryThreshold,
		startTime:          time.Now(),
	}
}

// RecordQuery records a query execution.
func (a *DBStatsAggregator) RecordQuery(operation, query string, duration time.Duration, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	queryType := detectQueryType(query)

	// Update global stats
	a.updateStats(&a.global, operation, duration, err, now)
	a.updateQueryType(&a.global, queryType)

	// Update per-operation stats
	opStats := a.byOperation[operation]
	a.updateStats(&opStats, operation, duration, err, now)
	a.updateQueryType(&opStats, queryType)
	a.byOperation[operation] = opStats

	// Extract table name and update per-table stats
	if table := extractTableName(query); table != "" {
		tableStats := a.byTable[table]
		a.updateStats(&tableStats, operation, duration, err, now)
		a.updateQueryType(&tableStats, queryType)
		a.byTable[table] = tableStats
	}
}

func (a *DBStatsAggregator) updateStats(stats *QueryStats, operation string, duration time.Duration, err error, now time.Time) {
	stats.TotalQueries++
	stats.LastUpdate = now

	if err == nil {
		stats.SuccessQueries++
	} else {
		stats.FailedQueries++
	}

	stats.TotalDuration += duration

	if stats.MinDuration == 0 || duration < stats.MinDuration {
		stats.MinDuration = duration
	}
	if duration > stats.MaxDuration {
		stats.MaxDuration = duration
	}

	if stats.TotalQueries > 0 {
		stats.AvgDuration = stats.TotalDuration / time.Duration(stats.TotalQueries)
	}

	// Track slow queries
	if duration >= a.slowQueryThreshold {
		stats.SlowQueries++
	}
	stats.SlowQueryThresh = a.slowQueryThreshold
}

func (a *DBStatsAggregator) updateQueryType(stats *QueryStats, queryType string) {
	switch queryType {
	case "SELECT":
		stats.SelectQueries++
	case "INSERT":
		stats.InsertQueries++
	case "UPDATE":
		stats.UpdateQueries++
	case "DELETE":
		stats.DeleteQueries++
	}
}

func detectQueryType(query string) string {
	if query == "" {
		return ""
	}

	query = strings.TrimSpace(strings.ToUpper(query))

	if strings.HasPrefix(query, "SELECT") {
		return "SELECT"
	}
	if strings.HasPrefix(query, "INSERT") {
		return "INSERT"
	}
	if strings.HasPrefix(query, "UPDATE") {
		return "UPDATE"
	}
	if strings.HasPrefix(query, "DELETE") {
		return "DELETE"
	}

	return ""
}

// GetGlobalStats returns global statistics.
func (a *DBStatsAggregator) GetGlobalStats() QueryStats {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.global
}

// GetOperationStats returns statistics for a specific operation.
func (a *DBStatsAggregator) GetOperationStats(operation string) (QueryStats, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	stats, ok := a.byOperation[operation]
	return stats, ok
}

// GetAllOperationStats returns statistics for all operations.
func (a *DBStatsAggregator) GetAllOperationStats() map[string]QueryStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]QueryStats, len(a.byOperation))
	for k, v := range a.byOperation {
		result[k] = v
	}
	return result
}

// GetTableStats returns statistics for a specific table.
func (a *DBStatsAggregator) GetTableStats(table string) (QueryStats, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	stats, ok := a.byTable[table]
	return stats, ok
}

// GetAllTableStats returns statistics for all tables.
func (a *DBStatsAggregator) GetAllTableStats() map[string]QueryStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]QueryStats, len(a.byTable))
	for k, v := range a.byTable {
		result[k] = v
	}
	return result
}

// GetTopSlowTables returns the top N tables by slow query count.
func (a *DBStatsAggregator) GetTopSlowTables(n int) []TableStat {
	a.mu.RLock()
	defer a.mu.RUnlock()

	tables := make([]TableStat, 0, len(a.byTable))
	for name, stats := range a.byTable {
		tables = append(tables, TableStat{
			Table:        name,
			SlowQueries:  stats.SlowQueries,
			AvgDuration:  stats.AvgDuration,
			TotalQueries: stats.TotalQueries,
		})
	}

	// Sort by slow query count descending
	for i := 0; i < len(tables); i++ {
		for j := i + 1; j < len(tables); j++ {
			if tables[j].SlowQueries > tables[i].SlowQueries {
				tables[i], tables[j] = tables[j], tables[i]
			}
		}
	}

	if n > len(tables) {
		n = len(tables)
	}
	return tables[:n]
}

// TableStat represents statistics for a single table.
type TableStat struct {
	Table        string        `json:"table"`
	SlowQueries  int64         `json:"slow_queries"`
	AvgDuration  time.Duration `json:"avg_duration"`
	TotalQueries int64         `json:"total_queries"`
}

// Reset resets all statistics.
func (a *DBStatsAggregator) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.global = QueryStats{}
	a.byOperation = make(map[string]QueryStats)
	a.byTable = make(map[string]QueryStats)
	a.startTime = time.Now()
}

// Summary returns a formatted summary of all statistics.
func (a *DBStatsAggregator) Summary() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var sb strings.Builder

	sb.WriteString("=== Database Statistics Summary ===\n\n")
	sb.WriteString(fmt.Sprintf("Uptime: %s\n", time.Since(a.startTime).Round(time.Second)))
	sb.WriteString(fmt.Sprintf("Slow Query Threshold: %s\n\n", a.slowQueryThreshold))

	sb.WriteString("Global Stats:\n")
	sb.WriteString(formatQueryStats(a.global))
	sb.WriteString("\n")

	if len(a.byOperation) > 0 {
		sb.WriteString("Stats by Operation:\n")
		for op, stats := range a.byOperation {
			sb.WriteString(fmt.Sprintf("  %s:\n", op))
			sb.WriteString(indentString(formatQueryStats(stats), "    "))
		}
		sb.WriteString("\n")
	}

	if len(a.byTable) > 0 {
		sb.WriteString("Top Slow Tables:\n")
		topTables := a.getTopSlowTablesLocked(5)
		for i, ts := range topTables {
			sb.WriteString(fmt.Sprintf("  %d. %s - %d slow queries (avg: %s)\n",
				i+1, ts.Table, ts.SlowQueries, ts.AvgDuration.Round(time.Millisecond)))
		}
	}

	return sb.String()
}

func (a *DBStatsAggregator) getTopSlowTablesLocked(n int) []TableStat {
	tables := make([]TableStat, 0, len(a.byTable))
	for name, stats := range a.byTable {
		tables = append(tables, TableStat{
			Table:        name,
			SlowQueries:  stats.SlowQueries,
			AvgDuration:  stats.AvgDuration,
			TotalQueries: stats.TotalQueries,
		})
	}

	for i := 0; i < len(tables); i++ {
		for j := i + 1; j < len(tables); j++ {
			if tables[j].SlowQueries > tables[i].SlowQueries {
				tables[i], tables[j] = tables[j], tables[i]
			}
		}
	}

	if n > len(tables) {
		n = len(tables)
	}
	return tables[:n]
}

func formatQueryStats(stats QueryStats) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total Queries: %d\n", stats.TotalQueries))
	sb.WriteString(fmt.Sprintf("Success: %d, Failed: %d\n", stats.SuccessQueries, stats.FailedQueries))
	if stats.TotalQueries > 0 {
		successRate := float64(stats.SuccessQueries) / float64(stats.TotalQueries) * 100
		sb.WriteString(fmt.Sprintf("Success Rate: %.2f%%\n", successRate))
	}
	sb.WriteString(fmt.Sprintf("Duration: min=%s, avg=%s, max=%s\n",
		stats.MinDuration.Round(time.Millisecond),
		stats.AvgDuration.Round(time.Millisecond),
		stats.MaxDuration.Round(time.Millisecond)))
	sb.WriteString(fmt.Sprintf("Query Types: SELECT=%d, INSERT=%d, UPDATE=%d, DELETE=%d\n",
		stats.SelectQueries, stats.InsertQueries, stats.UpdateQueries, stats.DeleteQueries))
	sb.WriteString(fmt.Sprintf("Slow Queries: %d (>%s)\n", stats.SlowQueries, stats.SlowQueryThresh))
	return sb.String()
}

func indentString(s, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// extractTableName attempts to extract the table name from a SQL query.
// This is a simple heuristic and may not work for all query types.
func extractTableName(query string) string {
	if query == "" {
		return ""
	}

	query = strings.TrimSpace(strings.ToUpper(query))

	// Handle SELECT
	if strings.HasPrefix(query, "SELECT") {
		if idx := strings.Index(query, "FROM"); idx != -1 {
			rest := strings.TrimSpace(query[idx+4:])
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				return strings.ToLower(cleanTableName(parts[0]))
			}
		}
	}

	// Handle INSERT
	if strings.HasPrefix(query, "INSERT INTO") {
		rest := strings.TrimSpace(query[11:])
		parts := strings.Fields(rest)
		if len(parts) > 0 {
			return strings.ToLower(cleanTableName(parts[0]))
		}
	}

	// Handle UPDATE
	if strings.HasPrefix(query, "UPDATE") {
		rest := strings.TrimSpace(query[6:])
		parts := strings.Fields(rest)
		if len(parts) > 0 {
			return strings.ToLower(cleanTableName(parts[0]))
		}
	}

	// Handle DELETE
	if strings.HasPrefix(query, "DELETE FROM") {
		rest := strings.TrimSpace(query[11:])
		parts := strings.Fields(rest)
		if len(parts) > 0 {
			return strings.ToLower(cleanTableName(parts[0]))
		}
	}

	return ""
}

func cleanTableName(name string) string {
	// Remove common SQL characters
	name = strings.TrimFunc(name, func(r rune) bool {
		return r == '`' || r == '"' || r == '\'' || r == '(' || r == ','
	})
	return name
}

// AggregatingCollector wraps a MetricsCollector and also aggregates statistics.
type AggregatingCollector struct {
	base       MetricsCollector
	aggregator *DBStatsAggregator
}

// NewAggregatingCollector creates a collector that both records metrics and aggregates statistics.
func NewAggregatingCollector(base MetricsCollector, slowQueryThreshold time.Duration) *AggregatingCollector {
	return &AggregatingCollector{
		base:       base,
		aggregator: NewDBStatsAggregator(slowQueryThreshold),
	}
}

// ObserveDB implements MetricsCollector and also aggregates statistics.
func (c *AggregatingCollector) ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
	// Forward to base collector
	if c.base != nil {
		c.base.ObserveDB(ctx, operation, driver, query, rows, duration, err)
	}

	// Aggregate statistics
	c.aggregator.RecordQuery(operation, query, duration, err)
}

// GetAggregator returns the underlying statistics aggregator.
func (c *AggregatingCollector) GetAggregator() *DBStatsAggregator {
	return c.aggregator
}
