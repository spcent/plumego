// Package sqlx adapts github.com/jmoiron/sqlx behind a small explicit query
// and transaction surface for x/data callers.
package sqlx

import (
	"context"
	"database/sql"
	"fmt"

	upstream "github.com/jmoiron/sqlx"
	storedb "github.com/spcent/plumego/store/db"
)

// Querier is the context-aware query surface exposed by DB and Tx.
type Querier interface {
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Transactor starts sqlx transactions.
type Transactor interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

// Tx is a sqlx transaction with the same query surface as DB.
type Tx interface {
	Querier
	Commit() error
	Rollback() error
}

// DB wraps a sqlx.DB and exposes explicit query and transaction methods.
type DB struct {
	db *upstream.DB
}

var (
	_ Querier    = (*DB)(nil)
	_ Transactor = (*DB)(nil)
	_ Tx         = (*tx)(nil)
)

// New opens a sqlx database and verifies connectivity with Ping.
func New(driverName, dataSourceName string) (*DB, error) {
	opened, err := upstream.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("%w: open sqlx database: %w", storedb.ErrConnectionFailed, err)
	}
	if err := opened.Ping(); err != nil {
		_ = opened.Close()
		return nil, fmt.Errorf("%w: ping sqlx database: %w", storedb.ErrConnectionFailed, err)
	}
	return NewWithDB(opened), nil
}

// NewWithDB wraps an existing sqlx DB.
func NewWithDB(db *upstream.DB) *DB {
	return &DB{db: db}
}

// SQLX returns the wrapped sqlx DB.
func (d *DB) SQLX() *upstream.DB {
	if d == nil {
		return nil
	}
	return d.db
}

// QueryRow executes a query expected to return at most one row.
func (d *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.QueryRowContext(ctx, query, args...)
}

// Query executes a query returning rows.
func (d *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("%w: sqlx database is nil", storedb.ErrQueryFailed)
	}
	return d.db.QueryContext(ctx, query, args...)
}

// Exec executes a statement.
func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("%w: sqlx database is nil", storedb.ErrQueryFailed)
	}
	return d.db.ExecContext(ctx, query, args...)
}

// BeginTx begins a sqlx transaction.
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("%w: sqlx database is nil", storedb.ErrTransactionFailed)
	}
	started, err := d.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: begin sqlx transaction: %w", storedb.ErrTransactionFailed, err)
	}
	return &tx{tx: started}, nil
}

// NamedExec executes a named query using sqlx struct or map binding.
func (d *DB) NamedExec(ctx context.Context, query string, arg any) (sql.Result, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("%w: sqlx database is nil", storedb.ErrQueryFailed)
	}
	return d.db.NamedExecContext(ctx, query, arg)
}

// NamedQuery executes a named query using sqlx struct or map binding.
func (d *DB) NamedQuery(ctx context.Context, query string, arg any) (*upstream.Rows, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("%w: sqlx database is nil", storedb.ErrQueryFailed)
	}
	return d.db.NamedQueryContext(ctx, query, arg)
}

// Close closes the wrapped database.
func (d *DB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

type tx struct {
	tx *upstream.Tx
}

func (t *tx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if t == nil || t.tx == nil {
		return nil
	}
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *tx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if t == nil || t.tx == nil {
		return nil, fmt.Errorf("%w: sqlx transaction is nil", storedb.ErrQueryFailed)
	}
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if t == nil || t.tx == nil {
		return nil, fmt.Errorf("%w: sqlx transaction is nil", storedb.ErrQueryFailed)
	}
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *tx) Commit() error {
	if t == nil || t.tx == nil {
		return fmt.Errorf("%w: sqlx transaction is nil", storedb.ErrTransactionFailed)
	}
	return t.tx.Commit()
}

func (t *tx) Rollback() error {
	if t == nil || t.tx == nil {
		return fmt.Errorf("%w: sqlx transaction is nil", storedb.ErrTransactionFailed)
	}
	return t.tx.Rollback()
}
