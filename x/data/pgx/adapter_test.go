package pgx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	driver "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestQueryRowReturnsScannedValue(t *testing.T) {
	pool := &fakePool{
		row: fakeRow{values: []any{"plumego"}},
	}
	db := newWithRunner(pool)

	var got string
	if err := db.QueryRow(t.Context(), "select name").Scan(&got); err != nil {
		t.Fatalf("scan row: %v", err)
	}
	if got != "plumego" {
		t.Fatalf("got %q, want plumego", got)
	}
	if pool.queryRowSQL != "select name" {
		t.Fatalf("query = %q, want select name", pool.queryRowSQL)
	}
}

func TestQueryReturnsMultipleRows(t *testing.T) {
	pool := &fakePool{
		rows: &fakeRows{values: [][]any{{"alpha"}, {"beta"}}},
	}
	db := newWithRunner(pool)

	rows, err := db.Query(t.Context(), "select name")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got = append(got, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	if want := []string{"alpha", "beta"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rows = %#v, want %#v", got, want)
	}
}

func TestExecAffectsRowCount(t *testing.T) {
	pool := &fakePool{
		execTag: pgconn.NewCommandTag("UPDATE 3"),
	}
	db := newWithRunner(pool)

	tag, err := db.Exec(t.Context(), "update widgets set active = true")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if got := tag.RowsAffected(); got != 3 {
		t.Fatalf("rows affected = %d, want 3", got)
	}
}

func TestBeginTxCommitSucceeds(t *testing.T) {
	tx := &fakeTx{}
	db := newWithRunner(&fakePool{tx: tx})

	started, err := db.BeginTx(t.Context(), driver.TxOptions{})
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := started.Commit(t.Context()); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if !tx.committed {
		t.Fatal("expected transaction to be committed")
	}
}

func TestBeginTxRollbackSucceeds(t *testing.T) {
	tx := &fakeTx{}
	db := newWithRunner(&fakePool{tx: tx})

	started, err := db.BeginTx(t.Context(), driver.TxOptions{})
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := started.Rollback(t.Context()); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if !tx.rolledBack {
		t.Fatal("expected transaction to be rolled back")
	}
}

func TestNewConnectionFailureReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := New(ctx, "postgres://user:pass@127.0.0.1:1/plumego"); err == nil {
		t.Fatal("expected connection error")
	}
}

type fakePool struct {
	row         driver.Row
	rows        driver.Rows
	tx          driver.Tx
	execTag     pgconn.CommandTag
	queryRowSQL string
}

func (p *fakePool) QueryRow(_ context.Context, query string, _ ...any) driver.Row {
	p.queryRowSQL = query
	if p.row == nil {
		return fakeRow{err: errors.New("row not configured")}
	}
	return p.row
}

func (p *fakePool) Query(context.Context, string, ...any) (driver.Rows, error) {
	if p.rows == nil {
		return nil, errors.New("rows not configured")
	}
	return p.rows, nil
}

func (p *fakePool) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return p.execTag, nil
}

func (p *fakePool) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	if p.tx == nil {
		return nil, errors.New("tx not configured")
	}
	return p.tx, nil
}

func (p *fakePool) Ping(context.Context) error {
	return nil
}

func (p *fakePool) Close() {}

type fakeRow struct {
	values []any
	err    error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return scanValues(dest, r.values)
}

type fakeRows struct {
	values [][]any
	index  int
	closed bool
	err    error
}

func (r *fakeRows) Close() {
	r.closed = true
}

func (r *fakeRows) Err() error {
	return r.err
}

func (r *fakeRows) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag(fmt.Sprintf("SELECT %d", len(r.values)))
}

func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (r *fakeRows) Next() bool {
	if r.index >= len(r.values) {
		r.closed = true
		return false
	}
	r.index++
	return true
}

func (r *fakeRows) Scan(dest ...any) error {
	if r.index == 0 || r.index > len(r.values) {
		return errors.New("scan called without current row")
	}
	return scanValues(dest, r.values[r.index-1])
}

func (r *fakeRows) Values() ([]any, error) {
	if r.index == 0 || r.index > len(r.values) {
		return nil, errors.New("values called without current row")
	}
	return append([]any(nil), r.values[r.index-1]...), nil
}

func (r *fakeRows) RawValues() [][]byte {
	return nil
}

func (r *fakeRows) Conn() *driver.Conn {
	return nil
}

type fakeTx struct {
	committed  bool
	rolledBack bool
}

func (t *fakeTx) Begin(context.Context) (driver.Tx, error) {
	return &fakeTx{}, nil
}

func (t *fakeTx) Commit(context.Context) error {
	t.committed = true
	return nil
}

func (t *fakeTx) Rollback(context.Context) error {
	t.rolledBack = true
	return nil
}

func (t *fakeTx) CopyFrom(context.Context, driver.Identifier, []string, driver.CopyFromSource) (int64, error) {
	return 0, nil
}

func (t *fakeTx) SendBatch(context.Context, *driver.Batch) driver.BatchResults {
	return nil
}

func (t *fakeTx) LargeObjects() driver.LargeObjects {
	return driver.LargeObjects{}
}

func (t *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (t *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (t *fakeTx) Query(context.Context, string, ...any) (driver.Rows, error) {
	return &fakeRows{}, nil
}

func (t *fakeTx) QueryRow(context.Context, string, ...any) driver.Row {
	return fakeRow{}
}

func (t *fakeTx) Conn() *driver.Conn {
	return nil
}

func scanValues(dest []any, values []any) error {
	if len(dest) != len(values) {
		return fmt.Errorf("scan destination count %d does not match value count %d", len(dest), len(values))
	}
	for i, value := range values {
		target := reflect.ValueOf(dest[i])
		if target.Kind() != reflect.Ptr || target.IsNil() {
			return fmt.Errorf("destination %d is not a non-nil pointer", i)
		}
		source := reflect.ValueOf(value)
		if !source.Type().AssignableTo(target.Elem().Type()) {
			return fmt.Errorf("cannot assign %T to %T", value, dest[i])
		}
		target.Elem().Set(source)
	}
	return nil
}
