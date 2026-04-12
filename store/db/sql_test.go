package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "missing driver",
			config:  Config{DSN: "dsn"},
			wantErr: true,
		},
		{
			name:    "missing dsn",
			config:  Config{Driver: "driver"},
			wantErr: true,
		},
		{
			name:    "negative max open",
			config:  Config{Driver: "driver", DSN: "dsn", MaxOpenConns: -1},
			wantErr: true,
		},
		{
			name:    "negative ping timeout",
			config:  Config{Driver: "driver", DSN: "dsn", PingTimeout: -1},
			wantErr: true,
		},
		{
			name:    "valid",
			config:  Config{Driver: "driver", DSN: "dsn", MaxOpenConns: 5},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		if err := tc.config.Validate(); (err != nil) != tc.wantErr {
			t.Fatalf("%s: Validate() error = %v", tc.name, err)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("mysql", "user:pass@tcp(localhost:3306)/db")

	if config.Driver != "mysql" {
		t.Fatalf("expected driver mysql, got %s", config.Driver)
	}
	if config.DSN != "user:pass@tcp(localhost:3306)/db" {
		t.Fatalf("expected DSN, got %s", config.DSN)
	}
	if config.MaxOpenConns != 10 {
		t.Fatalf("expected MaxOpenConns 10, got %d", config.MaxOpenConns)
	}
	if config.MaxIdleConns != 5 {
		t.Fatalf("expected MaxIdleConns 5, got %d", config.MaxIdleConns)
	}
	if config.PingTimeout != 5*time.Second {
		t.Fatalf("expected PingTimeout 5s, got %v", config.PingTimeout)
	}
}

func TestApplyConfigMaxOpenConns(t *testing.T) {
	conn := &stubConn{}
	connector := &stubConnector{conn: conn}
	db := sql.OpenDB(connector)
	defer db.Close()

	ApplyConfig(db, Config{MaxOpenConns: 4, MaxIdleConns: 10})
	stats := db.Stats()
	if stats.MaxOpenConnections != 4 {
		t.Fatalf("expected MaxOpenConnections 4, got %d", stats.MaxOpenConnections)
	}
}

func TestOpenWithPing(t *testing.T) {
	pingErr := errors.New("ping failed")
	conn := &stubConn{pingErr: pingErr}
	connector := &stubConnector{conn: conn}

	_, err := OpenWith(Config{
		Driver:      "stub",
		DSN:         "dsn",
		PingTimeout: 50 * time.Millisecond,
	}, func(driver, dsn string) (*sql.DB, error) {
		return sql.OpenDB(connector), nil
	})
	if err == nil || !errors.Is(err, ErrPingFailed) {
		t.Fatalf("expected ErrPingFailed, got %v", err)
	}
}

func TestExecContext(t *testing.T) {
	connector := &stubConnector{conn: &stubConn{}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	_, err := ExecContext(ctx, db, "INSERT INTO test VALUES (?)", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecContextNilDB(t *testing.T) {
	_, err := ExecContext(context.Background(), nil, "INSERT INTO test VALUES (?)", 1)
	if err == nil || !errors.Is(err, ErrQueryFailed) {
		t.Fatalf("expected ErrQueryFailed, got %v", err)
	}
}

func TestExecContextUsesCallerContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), testContextKey{}, "request")
	db := &contextRecorderDB{}

	if _, err := ExecContext(ctx, db, "INSERT INTO test VALUES (?)", 1); err != nil {
		t.Fatalf("ExecContext: %v", err)
	}
	if db.execCtx != ctx {
		t.Fatal("expected ExecContext to receive caller context")
	}
}

func TestQueryContext(t *testing.T) {
	connector := &stubConnector{conn: &stubConn{}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	rows, err := QueryContext(ctx, db, "SELECT * FROM test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows == nil {
		t.Fatal("expected rows")
	}
	rows.Close()
}

func TestQueryContextNilDB(t *testing.T) {
	_, err := QueryContext(context.Background(), nil, "SELECT * FROM test")
	if err == nil || !errors.Is(err, ErrQueryFailed) {
		t.Fatalf("expected ErrQueryFailed, got %v", err)
	}
}

func TestQueryContextUsesCallerContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), testContextKey{}, "request")
	db := &contextRecorderDB{}

	if _, err := QueryContext(ctx, db, "SELECT * FROM test"); err != nil {
		t.Fatalf("QueryContext: %v", err)
	}
	if db.queryCtx != ctx {
		t.Fatal("expected QueryContext to receive caller context")
	}
}

func TestQueryRowContext(t *testing.T) {
	connector := &stubConnector{conn: &stubConn{}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	row := QueryRowContext(ctx, db, "SELECT * FROM test WHERE id = ?", 1)
	if row == nil {
		t.Fatal("expected row")
	}
}

func TestQueryRowContextNilDB(t *testing.T) {
	row := QueryRowContext(context.Background(), nil, "SELECT * FROM test WHERE id = ?", 1)
	if row != nil {
		t.Fatal("expected nil row")
	}
}

func TestQueryRowContextUsesCallerContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), testContextKey{}, "request")
	db := &contextRecorderDB{}

	row := QueryRowContext(ctx, db, "SELECT * FROM test WHERE id = ?", 1)
	if row == nil {
		t.Fatal("expected row")
	}
	if db.queryRowCtx != ctx {
		t.Fatal("expected QueryRowContext to receive caller context")
	}
}

func TestWithTransaction(t *testing.T) {
	connector := &stubConnector{conn: &stubConn{}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	err := WithTransaction(ctx, db, nil, func(tx *sql.Tx) error {
		// Simulate transaction work
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithTransactionNilDB(t *testing.T) {
	err := WithTransaction(context.Background(), nil, nil, func(tx *sql.Tx) error {
		return nil
	})
	if err == nil || !errors.Is(err, ErrTransactionFailed) {
		t.Fatalf("expected ErrTransactionFailed, got %v", err)
	}
}

func TestWithTransactionError(t *testing.T) {
	connector := &stubConnector{conn: &stubConn{}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	txErr := errors.New("transaction error")
	err := WithTransaction(ctx, db, nil, func(tx *sql.Tx) error {
		return txErr
	})
	if err == nil || !errors.Is(err, ErrTransactionFailed) {
		t.Fatalf("expected ErrTransactionFailed, got %v", err)
	}
}

func TestWithTransactionUsesCallerContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), testContextKey{}, "request")
	beginErr := errors.New("begin failed")
	db := &contextRecorderDB{beginErr: beginErr}

	err := WithTransaction(ctx, db, nil, func(tx *sql.Tx) error {
		t.Fatal("function should not run when begin fails")
		return nil
	})
	if err == nil || !errors.Is(err, ErrTransactionFailed) {
		t.Fatalf("expected ErrTransactionFailed, got %v", err)
	}
	if db.beginCtx != ctx {
		t.Fatal("expected BeginTx to receive caller context")
	}
}

func TestScanRow(t *testing.T) {
	connector := &stubConnector{conn: &stubConn{}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	row := QueryRowContext(ctx, db, "SELECT * FROM test WHERE id = ?", 1)

	var id int
	err := ScanRow(row, &id)
	// The stub returns no rows, so we expect ErrNoRows
	if err != nil && !errors.Is(err, ErrNoRows) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScanRowNil(t *testing.T) {
	err := ScanRow(nil, nil)
	if err == nil || !errors.Is(err, ErrQueryFailed) {
		t.Fatalf("expected ErrQueryFailed, got %v", err)
	}
}

func TestScanRows(t *testing.T) {
	connector := &stubConnector{conn: &stubConn{}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	rows, err := QueryContext(ctx, db, "SELECT * FROM test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scanFunc := func(rows *sql.Rows) (int, error) {
		var id int
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		return id, nil
	}

	results, err := ScanRows(rows, scanFunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The stub returns no rows, so results should be nil or empty slice
	// Since rows.Next() returns false immediately, results will be nil
	if results == nil {
		// This is expected when no rows are returned
		return
	}
	// Empty slice is also acceptable
	if len(results) != 0 {
		t.Fatalf("expected empty results, got %d", len(results))
	}
}

func TestScanRowsNil(t *testing.T) {
	_, err := ScanRows(nil, func(rows *sql.Rows) (int, error) {
		return 0, nil
	})
	if err == nil || !errors.Is(err, ErrQueryFailed) {
		t.Fatalf("expected ErrQueryFailed, got %v", err)
	}
}

func TestPing(t *testing.T) {
	connector := &stubConnector{conn: &stubConn{}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	err := Ping(ctx, db, 10*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPingNilDB(t *testing.T) {
	err := Ping(context.Background(), nil, 10*time.Second)
	if err == nil || !errors.Is(err, ErrPingFailed) {
		t.Fatalf("expected ErrPingFailed, got %v", err)
	}
}

func TestPingWithTimeout(t *testing.T) {
	pingErr := errors.New("ping failed")
	connector := &stubConnector{conn: &stubConn{pingErr: pingErr}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	err := Ping(ctx, db, 10*time.Millisecond)
	if err == nil || !errors.Is(err, ErrPingFailed) {
		t.Fatalf("expected ErrPingFailed, got %v", err)
	}
}

func TestQueryRow(t *testing.T) {
	connector := &stubConnector{conn: &stubConn{}}
	db := sql.OpenDB(connector)
	defer db.Close()

	ctx := context.Background()
	row, err := QueryRow(ctx, db, "SELECT * FROM test WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row == nil {
		t.Fatal("expected row")
	}
}

func TestQueryRowNilDB(t *testing.T) {
	_, err := QueryRow(context.Background(), nil, "SELECT * FROM test WHERE id = ?", 1)
	if err == nil || !errors.Is(err, ErrQueryFailed) {
		t.Fatalf("expected ErrQueryFailed, got %v", err)
	}
}

func TestQueryRowUsesCallerContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), testContextKey{}, "request")
	db := &contextRecorderDB{}

	row, err := QueryRow(ctx, db, "SELECT * FROM test WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("QueryRow: %v", err)
	}
	if row == nil {
		t.Fatal("expected row")
	}
	if db.queryRowCtx != ctx {
		t.Fatal("expected QueryRow to receive caller context")
	}
}

func TestQueryRowStrictNoRows(t *testing.T) {
	db := sql.OpenDB(&rowsConnector{rows: &fixedRows{cols: []string{"id"}}})
	defer db.Close()

	err := QueryRowStrict(context.Background(), db, "SELECT id FROM test", func(rows *sql.Rows) error {
		var id int
		return rows.Scan(&id)
	})
	if err == nil || !errors.Is(err, ErrNoRows) {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
}

func TestQueryRowStrictMultipleRows(t *testing.T) {
	db := sql.OpenDB(&rowsConnector{rows: &fixedRows{
		cols:   []string{"id"},
		values: [][]driver.Value{{1}, {2}},
	}})
	defer db.Close()

	err := QueryRowStrict(context.Background(), db, "SELECT id FROM test", func(rows *sql.Rows) error {
		var id int
		return rows.Scan(&id)
	})
	if err == nil || !errors.Is(err, ErrMultipleRows) {
		t.Fatalf("expected ErrMultipleRows, got %v", err)
	}
}

// Stub implementations for testing
type stubConnector struct {
	conn *stubConn
}

func (c *stubConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.conn, nil
}

func (c *stubConnector) Driver() driver.Driver {
	return stubDriver{}
}

type stubDriver struct{}

func (d stubDriver) Open(name string) (driver.Conn, error) {
	return nil, errors.New("not supported")
}

type stubConn struct {
	pingErr error
}

func (c *stubConn) Prepare(query string) (driver.Stmt, error) {
	return stubStmt{}, nil
}

func (c *stubConn) Close() error {
	return nil
}

func (c *stubConn) Begin() (driver.Tx, error) {
	return stubTx{}, nil
}

func (c *stubConn) Ping(ctx context.Context) error {
	// Add a small delay to ensure measurable latency in tests
	time.Sleep(1 * time.Millisecond)
	return c.pingErr
}

type stubStmt struct{}

func (s stubStmt) Close() error {
	return nil
}

func (s stubStmt) NumInput() int {
	return -1
}

func (s stubStmt) Exec(args []driver.Value) (driver.Result, error) {
	return stubResult{}, nil
}

func (s stubStmt) Query(args []driver.Value) (driver.Rows, error) {
	return stubRows{}, nil
}

type stubTx struct{}

func (t stubTx) Commit() error {
	return nil
}

func (t stubTx) Rollback() error {
	return nil
}

type stubResult struct{}

func (r stubResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r stubResult) RowsAffected() (int64, error) {
	return 0, nil
}

type stubRows struct{}

func (r stubRows) Columns() []string {
	return nil
}

func (r stubRows) Close() error {
	return nil
}

func (r stubRows) Next(dest []driver.Value) error {
	return io.EOF
}

type rowsConnector struct {
	rows driver.Rows
}

func (c *rowsConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return &rowsConn{rows: c.rows}, nil
}

func (c *rowsConnector) Driver() driver.Driver {
	return rowsDriver{}
}

type rowsDriver struct{}

func (d rowsDriver) Open(name string) (driver.Conn, error) {
	return nil, errors.New("not supported")
}

type rowsConn struct {
	rows driver.Rows
}

func (c *rowsConn) Prepare(query string) (driver.Stmt, error) {
	return rowsStmt{rows: c.rows}, nil
}

func (c *rowsConn) Close() error {
	return nil
}

func (c *rowsConn) Begin() (driver.Tx, error) {
	return stubTx{}, nil
}

func (c *rowsConn) Ping(ctx context.Context) error {
	return nil
}

type rowsStmt struct {
	rows driver.Rows
}

func (s rowsStmt) Close() error {
	return nil
}

func (s rowsStmt) NumInput() int {
	return -1
}

func (s rowsStmt) Exec(args []driver.Value) (driver.Result, error) {
	return stubResult{}, nil
}

func (s rowsStmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.rows, nil
}

type fixedRows struct {
	cols   []string
	values [][]driver.Value
	idx    int
}

func (r *fixedRows) Columns() []string {
	return r.cols
}

func (r *fixedRows) Close() error {
	return nil
}

func (r *fixedRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.idx])
	r.idx++
	return nil
}

type testContextKey struct{}

type contextRecorderDB struct {
	execCtx     context.Context
	queryCtx    context.Context
	queryRowCtx context.Context
	beginCtx    context.Context
	beginErr    error
}

func (db *contextRecorderDB) ExecContext(ctx context.Context, _ string, _ ...any) (sql.Result, error) {
	db.execCtx = ctx
	return stubResult{}, nil
}

func (db *contextRecorderDB) QueryContext(ctx context.Context, _ string, _ ...any) (*sql.Rows, error) {
	db.queryCtx = ctx
	return nil, nil
}

func (db *contextRecorderDB) QueryRowContext(ctx context.Context, _ string, _ ...any) *sql.Row {
	db.queryRowCtx = ctx
	return &sql.Row{}
}

func (db *contextRecorderDB) BeginTx(ctx context.Context, _ *sql.TxOptions) (*sql.Tx, error) {
	db.beginCtx = ctx
	return nil, db.beginErr
}

func (db *contextRecorderDB) PingContext(context.Context) error {
	return nil
}

func (db *contextRecorderDB) Close() error {
	return nil
}
