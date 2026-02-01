package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// mockMetricsCollector is a test implementation of MetricsCollector
type mockMetricsCollector struct {
	calls []metricsCall
}

type metricsCall struct {
	operation string
	driver    string
	query     string
	rows      int
	duration  time.Duration
	err       error
}

func (m *mockMetricsCollector) ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error) {
	m.calls = append(m.calls, metricsCall{
		operation: operation,
		driver:    driver,
		query:     query,
		rows:      rows,
		duration:  duration,
		err:       err,
	})
}

func (m *mockMetricsCollector) getLastCall() *metricsCall {
	if len(m.calls) == 0 {
		return nil
	}
	return &m.calls[len(m.calls)-1]
}

func (m *mockMetricsCollector) callCount() int {
	return len(m.calls)
}

func (m *mockMetricsCollector) reset() {
	m.calls = nil
}

func TestNewInstrumentedDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	if idb == nil {
		t.Fatal("expected non-nil InstrumentedDB")
	}
	if idb.db != db {
		t.Error("expected db to be set")
	}
	if idb.collector != collector {
		t.Error("expected collector to be set")
	}
	if idb.driver != "sqlite3" {
		t.Errorf("expected driver to be 'sqlite3', got %q", idb.driver)
	}
}

func TestInstrumentedDB_ExecContext(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	ctx := context.Background()
	result, err := idb.ExecContext(ctx, "INSERT INTO test (name) VALUES (?)", "test")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	affected, _ := result.RowsAffected()
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}

	// Verify metrics were recorded
	if collector.callCount() != 1 {
		t.Fatalf("expected 1 metric call, got %d", collector.callCount())
	}

	call := collector.getLastCall()
	if call.operation != "exec" {
		t.Errorf("expected operation 'exec', got %q", call.operation)
	}
	if call.driver != "sqlite3" {
		t.Errorf("expected driver 'sqlite3', got %q", call.driver)
	}
	if call.rows != 1 {
		t.Errorf("expected 1 row, got %d", call.rows)
	}
	if call.err != nil {
		t.Errorf("expected no error, got %v", call.err)
	}
	if call.duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestInstrumentedDB_QueryContext(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create and populate test table
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = db.Exec("INSERT INTO test (name) VALUES ('test1'), ('test2')")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	ctx := context.Background()
	rows, err := idb.QueryContext(ctx, "SELECT * FROM test")
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}

	// Verify metrics were recorded
	if collector.callCount() != 1 {
		t.Fatalf("expected 1 metric call, got %d", collector.callCount())
	}

	call := collector.getLastCall()
	if call.operation != "query" {
		t.Errorf("expected operation 'query', got %q", call.operation)
	}
	if call.driver != "sqlite3" {
		t.Errorf("expected driver 'sqlite3', got %q", call.driver)
	}
	if call.err != nil {
		t.Errorf("expected no error, got %v", call.err)
	}
	if call.duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestInstrumentedDB_QueryRowContext(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create and populate test table
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = db.Exec("INSERT INTO test (name) VALUES ('test1')")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	ctx := context.Background()
	row := idb.QueryRowContext(ctx, "SELECT name FROM test WHERE id = ?", 1)

	var name string
	err = row.Scan(&name)
	if err != nil {
		t.Fatalf("failed to scan: %v", err)
	}

	if name != "test1" {
		t.Errorf("expected name 'test1', got %q", name)
	}

	// Verify metrics were recorded
	if collector.callCount() != 1 {
		t.Fatalf("expected 1 metric call, got %d", collector.callCount())
	}

	call := collector.getLastCall()
	if call.operation != "query" {
		t.Errorf("expected operation 'query', got %q", call.operation)
	}
	if call.driver != "sqlite3" {
		t.Errorf("expected driver 'sqlite3', got %q", call.driver)
	}
	if call.rows != 1 {
		t.Errorf("expected 1 row, got %d", call.rows)
	}
	if call.err != nil {
		t.Errorf("expected no error, got %v", call.err)
	}
}

func TestInstrumentedDB_BeginTx(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	ctx := context.Background()
	tx, err := idb.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Verify metrics were recorded
	if collector.callCount() != 1 {
		t.Fatalf("expected 1 metric call, got %d", collector.callCount())
	}

	call := collector.getLastCall()
	if call.operation != "transaction" {
		t.Errorf("expected operation 'transaction', got %q", call.operation)
	}
	if call.driver != "sqlite3" {
		t.Errorf("expected driver 'sqlite3', got %q", call.driver)
	}
	if call.query != "BEGIN" {
		t.Errorf("expected query 'BEGIN', got %q", call.query)
	}
	if call.err != nil {
		t.Errorf("expected no error, got %v", call.err)
	}
}

func TestInstrumentedDB_PingContext(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	ctx := context.Background()
	err = idb.PingContext(ctx)
	if err != nil {
		t.Fatalf("failed to ping: %v", err)
	}

	// Verify metrics were recorded
	if collector.callCount() != 1 {
		t.Fatalf("expected 1 metric call, got %d", collector.callCount())
	}

	call := collector.getLastCall()
	if call.operation != "ping" {
		t.Errorf("expected operation 'ping', got %q", call.operation)
	}
	if call.driver != "sqlite3" {
		t.Errorf("expected driver 'sqlite3', got %q", call.driver)
	}
	if call.err != nil {
		t.Errorf("expected no error, got %v", call.err)
	}
}

func TestInstrumentedDB_Close(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	err = idb.Close()
	if err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	// Verify metrics were recorded
	if collector.callCount() != 1 {
		t.Fatalf("expected 1 metric call, got %d", collector.callCount())
	}

	call := collector.getLastCall()
	if call.operation != "close" {
		t.Errorf("expected operation 'close', got %q", call.operation)
	}
	if call.driver != "sqlite3" {
		t.Errorf("expected driver 'sqlite3', got %q", call.driver)
	}
	if call.err != nil {
		t.Errorf("expected no error, got %v", call.err)
	}
}

func TestInstrumentedDB_ErrorTracking(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	ctx := context.Background()

	// Execute invalid query to trigger error
	_, err = idb.ExecContext(ctx, "INVALID SQL")
	if err == nil {
		t.Fatal("expected error for invalid SQL")
	}

	// Verify error was recorded in metrics
	if collector.callCount() != 1 {
		t.Fatalf("expected 1 metric call, got %d", collector.callCount())
	}

	call := collector.getLastCall()
	if call.err == nil {
		t.Error("expected error to be recorded in metrics")
	}
}

func TestInstrumentedDB_NilCollector(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create instrumented DB with nil collector
	idb := NewInstrumentedDB(db, nil, "sqlite3")

	ctx := context.Background()

	// All operations should work without panicking
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	_, err = idb.ExecContext(ctx, "INSERT INTO test (id) VALUES (1)")
	if err != nil {
		t.Fatalf("failed to exec: %v", err)
	}

	_, err = idb.QueryContext(ctx, "SELECT * FROM test")
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	_ = idb.QueryRowContext(ctx, "SELECT * FROM test WHERE id = 1")

	_, err = idb.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	err = idb.PingContext(ctx)
	if err != nil {
		t.Fatalf("failed to ping: %v", err)
	}

	// Test should complete without panics
}

func TestInstrumentedDB_Unwrap(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	unwrapped := idb.Unwrap()
	if unwrapped != db {
		t.Error("expected Unwrap to return the original db")
	}
}

func TestInstrumentedDB_Stats(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	stats := idb.Stats()
	// Just verify it doesn't panic and returns stats
	if stats.MaxOpenConnections < 0 {
		t.Error("expected valid stats")
	}
}

func TestInstrumentedDB_PoolMethods(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	// Test pool configuration methods don't panic
	idb.SetMaxOpenConns(10)
	idb.SetMaxIdleConns(5)
	idb.SetConnMaxLifetime(30 * time.Minute)
	idb.SetConnMaxIdleTime(5 * time.Minute)

	stats := idb.Stats()
	if stats.MaxOpenConnections != 10 {
		t.Errorf("expected MaxOpenConnections to be 10, got %d", stats.MaxOpenConnections)
	}
}

func TestInstrumentedDB_InterfaceCompliance(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	// Verify it implements DB interface
	var _ DB = idb
}

// BenchmarkInstrumentedDB_Overhead measures the overhead of metrics collection
func BenchmarkInstrumentedDB_Overhead(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		b.Fatalf("failed to create table: %v", err)
	}

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	ctx := context.Background()

	b.Run("WithMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = idb.ExecContext(ctx, "INSERT INTO test (name) VALUES (?)", "test")
		}
	})

	b.Run("WithoutMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = db.ExecContext(ctx, "INSERT INTO test (name) VALUES (?)", "test")
		}
	})
}

func TestInstrumentedDB_RecordPoolStats(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Configure pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	collector := &mockMetricsCollector{}
	idb := NewInstrumentedDB(db, collector, "sqlite3")

	// Record pool stats
	ctx := context.Background()
	idb.RecordPoolStats(ctx)

	// Verify metrics were recorded
	// Should record 7 pool metrics: open_connections, in_use, idle, wait_count, wait_duration, max_idle_closed, max_lifetime_closed
	if collector.callCount() < 7 {
		t.Errorf("expected at least 7 pool metrics, got %d", collector.callCount())
	}

	// Verify some specific operations were recorded
	operations := make(map[string]bool)
	for _, call := range collector.calls {
		operations[call.operation] = true
	}

	expectedOps := []string{
		"pool_open_connections",
		"pool_in_use",
		"pool_idle",
		"pool_wait_count",
		"pool_wait_duration",
		"pool_max_idle_closed",
		"pool_max_lifetime_closed",
	}

	for _, op := range expectedOps {
		if !operations[op] {
			t.Errorf("expected operation %q to be recorded", op)
		}
	}
}

func TestInstrumentedDB_RecordPoolStats_NilCollector(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create with nil collector
	idb := NewInstrumentedDB(db, nil, "sqlite3")

	// Should not panic
	ctx := context.Background()
	idb.RecordPoolStats(ctx)
}
