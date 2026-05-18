// Package pgx adapts pgx v5 pools and transactions behind a small explicit
// query and transaction surface for x/data callers.
package pgx

import (
	"context"
	"fmt"

	driver "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	storedb "github.com/spcent/plumego/store/db"
)

// Querier is the pgx-native query surface exposed by DB and Tx.
type Querier interface {
	QueryRow(ctx context.Context, query string, args ...any) driver.Row
	Query(ctx context.Context, query string, args ...any) (driver.Rows, error)
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

// Transactor starts pgx transactions.
type Transactor interface {
	BeginTx(ctx context.Context, opts driver.TxOptions) (Tx, error)
}

// Tx is a pgx transaction with the same query surface as DB.
type Tx interface {
	Querier
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type poolRunner interface {
	Querier
	BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error)
	Ping(ctx context.Context) error
	Close()
}

// DB wraps a pgx pool and exposes explicit query and transaction methods.
type DB struct {
	pool poolRunner
}

var (
	_ poolRunner = (*pgxpool.Pool)(nil)
	_ Querier    = (*DB)(nil)
	_ Transactor = (*DB)(nil)
	_ Tx         = (*tx)(nil)
)

// New creates a pgx pool, verifies connectivity with Ping, and wraps it.
func New(ctx context.Context, connStr string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("%w: parse pgx config: %w", storedb.ErrInvalidConfig, err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: create pgx pool: %w", storedb.ErrConnectionFailed, err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: ping pgx pool: %w", storedb.ErrConnectionFailed, err)
	}
	return NewWithPool(pool), nil
}

// NewWithPool wraps an existing pgx pool.
func NewWithPool(pool *pgxpool.Pool) *DB {
	if pool == nil {
		return &DB{}
	}
	return &DB{pool: pool}
}

// Pool returns the wrapped pgx pool when DB was built from a real pgxpool.Pool.
func (d *DB) Pool() *pgxpool.Pool {
	if d == nil {
		return nil
	}
	pool, _ := d.pool.(*pgxpool.Pool)
	return pool
}

func newWithRunner(pool poolRunner) *DB {
	return &DB{pool: pool}
}

// QueryRow executes a query expected to return at most one row.
func (d *DB) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	if d == nil || d.pool == nil {
		return rowError{err: fmt.Errorf("%w: pgx pool is nil", storedb.ErrQueryFailed)}
	}
	return d.pool.QueryRow(ctx, query, args...)
}

// Query executes a query returning rows.
func (d *DB) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	if d == nil || d.pool == nil {
		return nil, fmt.Errorf("%w: pgx pool is nil", storedb.ErrQueryFailed)
	}
	return d.pool.Query(ctx, query, args...)
}

// Exec executes a statement and returns the pgx command tag.
func (d *DB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if d == nil || d.pool == nil {
		return pgconn.CommandTag{}, fmt.Errorf("%w: pgx pool is nil", storedb.ErrQueryFailed)
	}
	return d.pool.Exec(ctx, query, args...)
}

// BeginTx begins a pgx transaction.
func (d *DB) BeginTx(ctx context.Context, opts driver.TxOptions) (Tx, error) {
	if d == nil || d.pool == nil {
		return nil, fmt.Errorf("%w: pgx pool is nil", storedb.ErrTransactionFailed)
	}
	started, err := d.pool.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: begin pgx transaction: %w", storedb.ErrTransactionFailed, err)
	}
	return &tx{tx: started}, nil
}

// Close closes the wrapped pool. pgxpool.Close does not return an error.
func (d *DB) Close() error {
	if d == nil || d.pool == nil {
		return nil
	}
	d.pool.Close()
	return nil
}

type tx struct {
	tx driver.Tx
}

func (t *tx) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	if t == nil || t.tx == nil {
		return rowError{err: fmt.Errorf("%w: pgx transaction is nil", storedb.ErrQueryFailed)}
	}
	return t.tx.QueryRow(ctx, query, args...)
}

func (t *tx) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	if t == nil || t.tx == nil {
		return nil, fmt.Errorf("%w: pgx transaction is nil", storedb.ErrQueryFailed)
	}
	return t.tx.Query(ctx, query, args...)
}

func (t *tx) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if t == nil || t.tx == nil {
		return pgconn.CommandTag{}, fmt.Errorf("%w: pgx transaction is nil", storedb.ErrQueryFailed)
	}
	return t.tx.Exec(ctx, query, args...)
}

func (t *tx) Commit(ctx context.Context) error {
	if t == nil || t.tx == nil {
		return fmt.Errorf("%w: pgx transaction is nil", storedb.ErrTransactionFailed)
	}
	return t.tx.Commit(ctx)
}

func (t *tx) Rollback(ctx context.Context) error {
	if t == nil || t.tx == nil {
		return fmt.Errorf("%w: pgx transaction is nil", storedb.ErrTransactionFailed)
	}
	return t.tx.Rollback(ctx)
}

type rowError struct {
	err error
}

func (r rowError) Scan(...any) error {
	return r.err
}
