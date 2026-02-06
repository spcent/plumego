package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/spcent/plumego/metrics"
)

// mockDB is a configurable stub connection (driver.Conn + driver.Pinger) for metrics tests.
// It allows injecting errors for each operation type and configuring rows affected.
type mockDB struct {
	pingErr      error
	execErr      error
	queryErr     error
	beginErr     error
	closeErr     error
	rowsAffected int64
}

func (c *mockDB) Prepare(query string) (driver.Stmt, error) {
	return &mockStmt{
		execErr:      c.execErr,
		queryErr:     c.queryErr,
		rowsAffected: c.rowsAffected,
	}, nil
}

func (c *mockDB) Close() error { return c.closeErr }

func (c *mockDB) Begin() (driver.Tx, error) {
	if c.beginErr != nil {
		return nil, c.beginErr
	}
	return stubTx{}, nil
}

func (c *mockDB) Ping(ctx context.Context) error { return c.pingErr }

type mockStmt struct {
	execErr      error
	queryErr     error
	rowsAffected int64
}

func (s *mockStmt) Close() error  { return nil }
func (s *mockStmt) NumInput() int { return -1 }

func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	if s.execErr != nil {
		return nil, s.execErr
	}
	return mockResult{rowsAffected: s.rowsAffected}, nil
}

func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	if s.queryErr != nil {
		return nil, s.queryErr
	}
	return &mockRows{}, nil
}

type mockResult struct {
	rowsAffected int64
}

func (r mockResult) LastInsertId() (int64, error) { return 0, nil }
func (r mockResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type mockRows struct{}

func (r *mockRows) Columns() []string              { return []string{} }
func (r *mockRows) Close() error                   { return nil }
func (r *mockRows) Next(dest []driver.Value) error { return io.EOF }

type mockConnector struct {
	conn driver.Conn
}

func (c *mockConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.conn, nil
}

func (c *mockConnector) Driver() driver.Driver { return stubDriver{} }

// newMockDB creates a real *sql.DB backed by a configurable stub connection.
// Options can be passed to configure error behavior and rows affected.
func newMockDB(opts ...func(*mockDB)) *sql.DB {
	conn := &mockDB{rowsAffected: 1}
	for _, opt := range opts {
		opt(conn)
	}
	return sql.OpenDB(&mockConnector{conn: conn})
}

func TestNewInstrumentedDB(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()

	idb := NewInstrumentedDB(db, collector, "postgres")

	if idb == nil {
		t.Fatal("expected non-nil InstrumentedDB")
	}
	if idb.db != db {
		t.Error("expected db to be set")
	}
	if idb.collector != collector {
		t.Error("expected collector to be set")
	}
	if idb.driver != "postgres" {
		t.Errorf("expected driver 'postgres', got %q", idb.driver)
	}
}

func TestInstrumentedDB_ExecContext(t *testing.T) {
	db := newMockDB(func(m *mockDB) {
		m.rowsAffected = 3
	})
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "mysql")

	ctx := context.Background()
	result, err := idb.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows, _ := result.RowsAffected()
	if rows != 3 {
		t.Errorf("expected 3 rows affected, got %d", rows)
	}

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 DB call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Operation != "exec" {
		t.Errorf("expected operation 'exec', got %q", call.Operation)
	}
	if call.Driver != "mysql" {
		t.Errorf("expected driver 'mysql', got %q", call.Driver)
	}
	if call.Query != "INSERT INTO users (name) VALUES (?)" {
		t.Errorf("unexpected query: %q", call.Query)
	}
	if call.Rows != 3 {
		t.Errorf("expected 3 rows in metrics, got %d", call.Rows)
	}
	if call.Duration <= 0 {
		t.Error("expected positive duration")
	}
	if call.Err != nil {
		t.Errorf("expected nil error, got %v", call.Err)
	}
}

func TestInstrumentedDB_ExecContext_Error(t *testing.T) {
	execErr := errors.New("exec failed")
	db := newMockDB(func(m *mockDB) {
		m.execErr = execErr
	})
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	_, err := idb.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "alice")
	if err == nil {
		t.Fatal("expected error")
	}

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 DB call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Operation != "exec" {
		t.Errorf("expected operation 'exec', got %q", call.Operation)
	}
	if call.Err == nil {
		t.Error("expected error to be recorded in metrics")
	}
	if call.Rows != 0 {
		t.Errorf("expected 0 rows on error, got %d", call.Rows)
	}
}

func TestInstrumentedDB_QueryContext(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	rows, err := idb.QueryContext(ctx, "SELECT * FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rows.Close()

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 DB call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Operation != "query" {
		t.Errorf("expected operation 'query', got %q", call.Operation)
	}
	if call.Driver != "postgres" {
		t.Errorf("expected driver 'postgres', got %q", call.Driver)
	}
	if call.Query != "SELECT * FROM users" {
		t.Errorf("unexpected query: %q", call.Query)
	}
	if call.Err != nil {
		t.Errorf("expected nil error, got %v", call.Err)
	}
}

func TestInstrumentedDB_QueryContext_Error(t *testing.T) {
	queryErr := errors.New("query failed")
	db := newMockDB(func(m *mockDB) {
		m.queryErr = queryErr
	})
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	_, err := idb.QueryContext(ctx, "SELECT * FROM invalid")
	if err == nil {
		t.Fatal("expected error")
	}

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 DB call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Operation != "query" {
		t.Errorf("expected operation 'query', got %q", call.Operation)
	}
	if call.Err == nil {
		t.Error("expected error to be recorded in metrics")
	}
}

func TestInstrumentedDB_QueryRowContext(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	row := idb.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", 1)
	if row == nil {
		t.Fatal("expected non-nil row")
	}

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 DB call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Operation != "query" {
		t.Errorf("expected operation 'query', got %q", call.Operation)
	}
	if call.Rows != 1 {
		t.Errorf("expected 1 row in metrics, got %d", call.Rows)
	}
	if call.Query != "SELECT * FROM users WHERE id = ?" {
		t.Errorf("unexpected query: %q", call.Query)
	}
}

func TestInstrumentedDB_BeginTx(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	tx, err := idb.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = tx.Rollback()

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 DB call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Operation != "transaction" {
		t.Errorf("expected operation 'transaction', got %q", call.Operation)
	}
	if call.Query != "BEGIN" {
		t.Errorf("expected query 'BEGIN', got %q", call.Query)
	}
	if call.Err != nil {
		t.Errorf("expected nil error, got %v", call.Err)
	}
	if call.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestInstrumentedDB_BeginTx_Error(t *testing.T) {
	beginErr := errors.New("begin failed")
	db := newMockDB(func(m *mockDB) {
		m.beginErr = beginErr
	})
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	_, err := idb.BeginTx(ctx, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 DB call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Operation != "transaction" {
		t.Errorf("expected operation 'transaction', got %q", call.Operation)
	}
	if call.Err == nil {
		t.Error("expected error to be recorded in metrics")
	}
}

func TestInstrumentedDB_PingContext(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	err := idb.PingContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 DB call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Operation != "ping" {
		t.Errorf("expected operation 'ping', got %q", call.Operation)
	}
	if call.Driver != "postgres" {
		t.Errorf("expected driver 'postgres', got %q", call.Driver)
	}
	if call.Err != nil {
		t.Errorf("expected nil error, got %v", call.Err)
	}
	if call.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestInstrumentedDB_PingContext_Error(t *testing.T) {
	pingErr := errors.New("ping failed")
	db := newMockDB(func(m *mockDB) {
		m.pingErr = pingErr
	})
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	err := idb.PingContext(ctx)
	if err == nil {
		t.Fatal("expected error")
	}

	call := collector.GetLastDBCall()
	if call.Operation != "ping" {
		t.Errorf("expected operation 'ping', got %q", call.Operation)
	}
	if call.Err == nil {
		t.Error("expected error to be recorded in metrics")
	}
}

func TestInstrumentedDB_Close(t *testing.T) {
	db := newMockDB()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	err := idb.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 DB call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Operation != "close" {
		t.Errorf("expected operation 'close', got %q", call.Operation)
	}
	if call.Driver != "postgres" {
		t.Errorf("expected driver 'postgres', got %q", call.Driver)
	}
}

func TestInstrumentedDB_NilCollector(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	idb := NewInstrumentedDB(db, nil, "postgres")

	ctx := context.Background()

	// None of these should panic with nil collector
	_, _ = idb.ExecContext(ctx, "INSERT INTO test VALUES (?)", 1)

	rows, _ := idb.QueryContext(ctx, "SELECT 1")
	if rows != nil {
		rows.Close()
	}

	_ = idb.QueryRowContext(ctx, "SELECT 1")

	tx, err := idb.BeginTx(ctx, nil)
	if err == nil {
		_ = tx.Rollback()
	}

	_ = idb.PingContext(ctx)
	idb.RecordPoolStats(ctx)
}

func TestInstrumentedDB_RecordPoolStats(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	idb.RecordPoolStats(ctx)

	if collector.DBCallCount() != 7 {
		t.Fatalf("expected 7 pool metrics, got %d", collector.DBCallCount())
	}

	operations := make(map[string]bool)
	for _, call := range collector.GetDBCalls() {
		operations[call.Operation] = true
		if call.Driver != "postgres" {
			t.Errorf("expected driver 'postgres' for %q, got %q", call.Operation, call.Driver)
		}
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
	db := newMockDB()
	defer db.Close()
	idb := NewInstrumentedDB(db, nil, "postgres")

	ctx := context.Background()
	// Should not panic
	idb.RecordPoolStats(ctx)
}

func TestInstrumentedDB_Stats(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	db.SetMaxOpenConns(20)
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	stats := idb.Stats()
	if stats.MaxOpenConnections != 20 {
		t.Errorf("expected MaxOpenConnections 20, got %d", stats.MaxOpenConnections)
	}
}

func TestInstrumentedDB_SetPoolConfig(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	idb.SetMaxOpenConns(15)
	idb.SetMaxIdleConns(5)
	idb.SetConnMaxLifetime(30 * time.Minute)
	idb.SetConnMaxIdleTime(5 * time.Minute)

	stats := idb.Stats()
	if stats.MaxOpenConnections != 15 {
		t.Errorf("expected MaxOpenConnections 15, got %d", stats.MaxOpenConnections)
	}
}

func TestInstrumentedDB_Unwrap(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	unwrapped := idb.Unwrap()
	if unwrapped != db {
		t.Error("expected Unwrap to return the original db")
	}
}

func TestInstrumentedDB_MultipleOperations(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()

	// Execute multiple operations through InstrumentedDB
	_, _ = idb.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "alice")

	rows, _ := idb.QueryContext(ctx, "SELECT * FROM users")
	if rows != nil {
		rows.Close()
	}

	_ = idb.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", 1)
	_ = idb.PingContext(ctx)

	if collector.DBCallCount() != 4 {
		t.Errorf("expected 4 DB calls, got %d", collector.DBCallCount())
	}

	calls := collector.GetDBCalls()
	expectedOps := []string{"exec", "query", "query", "ping"}
	for i, expected := range expectedOps {
		if calls[i].Operation != expected {
			t.Errorf("call %d: expected operation %q, got %q", i, expected, calls[i].Operation)
		}
		if calls[i].Driver != "postgres" {
			t.Errorf("call %d: expected driver 'postgres', got %q", i, calls[i].Driver)
		}
	}
}

func TestInstrumentedDB_ErrorTracking(t *testing.T) {
	queryErr := errors.New("table not found")
	db := newMockDB(func(m *mockDB) {
		m.queryErr = queryErr
	})
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	ctx := context.Background()
	_, err := idb.QueryContext(ctx, "SELECT * FROM nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}

	if collector.DBCallCount() != 1 {
		t.Fatalf("expected 1 metric call, got %d", collector.DBCallCount())
	}

	call := collector.GetLastDBCall()
	if call.Err == nil {
		t.Error("expected error to be recorded")
	}
	if call.Operation != "query" {
		t.Errorf("expected operation 'query', got %q", call.Operation)
	}
}

func TestInstrumentedDB_InterfaceCompliance(t *testing.T) {
	db := newMockDB()
	defer db.Close()
	collector := metrics.NewMockCollector()
	idb := NewInstrumentedDB(db, collector, "postgres")

	// Verify it implements DB interface
	var _ DB = idb
}
