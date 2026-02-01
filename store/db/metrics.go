package db

import (
	"context"
	"database/sql"
	"time"
)

// MetricsCollector defines the minimal interface for database metrics collection.
// It is intentionally decoupled from the main metrics package to avoid circular dependencies.
type MetricsCollector interface {
	ObserveDB(ctx context.Context, operation, driver, query string, rows int, duration time.Duration, err error)
}

// InstrumentedDB wraps a sql.DB with metrics collection.
// It implements the DB interface and records metrics for all database operations.
type InstrumentedDB struct {
	db        *sql.DB
	collector MetricsCollector
	driver    string
}

// NewInstrumentedDB creates a new instrumented database connection.
// The collector parameter can be nil to disable metrics.
func NewInstrumentedDB(db *sql.DB, collector MetricsCollector, driver string) *InstrumentedDB {
	return &InstrumentedDB{
		db:        db,
		collector: collector,
		driver:    driver,
	}
}

// ExecContext executes a query with metrics tracking.
func (idb *InstrumentedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := idb.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	// Record metrics
	if idb.collector != nil {
		rows := int64(0)
		if result != nil && err == nil {
			if affected, aErr := result.RowsAffected(); aErr == nil {
				rows = affected
			}
		}
		idb.collector.ObserveDB(ctx, "exec", idb.driver, query, int(rows), duration, err)
	}

	return result, err
}

// QueryContext executes a query with metrics tracking.
func (idb *InstrumentedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := idb.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	// Record metrics (rows count is not known yet)
	if idb.collector != nil {
		idb.collector.ObserveDB(ctx, "query", idb.driver, query, 0, duration, err)
	}

	return rows, err
}

// QueryRowContext executes a query with metrics tracking.
func (idb *InstrumentedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := idb.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	// Record metrics
	if idb.collector != nil {
		idb.collector.ObserveDB(ctx, "query", idb.driver, query, 1, duration, nil)
	}

	return row
}

// BeginTx starts a transaction with metrics tracking.
func (idb *InstrumentedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	start := time.Now()
	tx, err := idb.db.BeginTx(ctx, opts)
	duration := time.Since(start)

	// Record metrics
	if idb.collector != nil {
		idb.collector.ObserveDB(ctx, "transaction", idb.driver, "BEGIN", 0, duration, err)
	}

	return tx, err
}

// PingContext pings the database with metrics tracking.
func (idb *InstrumentedDB) PingContext(ctx context.Context) error {
	start := time.Now()
	err := idb.db.PingContext(ctx)
	duration := time.Since(start)

	// Record metrics
	if idb.collector != nil {
		idb.collector.ObserveDB(ctx, "ping", idb.driver, "", 0, duration, err)
	}

	return err
}

// Close closes the database connection with metrics tracking.
func (idb *InstrumentedDB) Close() error {
	start := time.Now()
	err := idb.db.Close()
	duration := time.Since(start)

	// Record metrics
	if idb.collector != nil {
		idb.collector.ObserveDB(context.Background(), "close", idb.driver, "", 0, duration, err)
	}

	return err
}

// Unwrap returns the underlying sql.DB instance.
// This is useful when you need access to DB-specific features like Stats().
func (idb *InstrumentedDB) Unwrap() *sql.DB {
	return idb.db
}

// Stats returns database statistics from the underlying connection.
func (idb *InstrumentedDB) Stats() sql.DBStats {
	return idb.db.Stats()
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
func (idb *InstrumentedDB) SetMaxOpenConns(n int) {
	idb.db.SetMaxOpenConns(n)
}

// SetMaxIdleConns sets the maximum number of idle connections in the connection pool.
func (idb *InstrumentedDB) SetMaxIdleConns(n int) {
	idb.db.SetMaxIdleConns(n)
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
func (idb *InstrumentedDB) SetConnMaxLifetime(d time.Duration) {
	idb.db.SetConnMaxLifetime(d)
}

// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
func (idb *InstrumentedDB) SetConnMaxIdleTime(d time.Duration) {
	idb.db.SetConnMaxIdleTime(d)
}

// Verify InstrumentedDB implements DB interface at compile time
var _ DB = (*InstrumentedDB)(nil)
