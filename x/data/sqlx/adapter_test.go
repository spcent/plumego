package sqlx

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	upstream "github.com/jmoiron/sqlx"
)

func TestQueryRowReturnsScannedValue(t *testing.T) {
	db, mock, done := newMockDB(t)
	defer done()

	mock.ExpectQuery("select name").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("plumego"))

	var got string
	if err := db.QueryRow(t.Context(), "select name").Scan(&got); err != nil {
		t.Fatalf("scan row: %v", err)
	}
	if got != "plumego" {
		t.Fatalf("got %q, want plumego", got)
	}
}

func TestQueryReturnsMultipleRows(t *testing.T) {
	db, mock, done := newMockDB(t)
	defer done()

	mock.ExpectQuery("select name").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("alpha").AddRow("beta"))

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
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("rows = %#v, want [alpha beta]", got)
	}
}

func TestExecAffectsRowCount(t *testing.T) {
	db, mock, done := newMockDB(t)
	defer done()

	mock.ExpectExec("update widgets").WillReturnResult(sqlmock.NewResult(0, 3))

	result, err := db.Exec(t.Context(), "update widgets set active = true")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if got, err := result.RowsAffected(); err != nil || got != 3 {
		t.Fatalf("rows affected = %d, %v; want 3, nil", got, err)
	}
}

func TestBeginTxCommitSucceeds(t *testing.T) {
	db, mock, done := newMockDB(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func TestBeginTxRollbackSucceeds(t *testing.T) {
	db, mock, done := newMockDB(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectRollback()

	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
}

func TestQueryRowScanError(t *testing.T) {
	db, mock, done := newMockDB(t)
	defer done()

	mock.ExpectQuery("select count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow("not-an-int"))

	var got int
	if err := db.QueryRow(t.Context(), "select count").Scan(&got); err == nil {
		t.Fatal("expected scan error")
	}
}

func newMockDB(t *testing.T) (*DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	wrapped := upstream.NewDb(raw, "sqlmock")
	db := NewWithDB(wrapped)

	return db, mock, func() {
		mock.ExpectClose()
		if err := db.Close(); err != nil && err != sql.ErrConnDone {
			t.Fatalf("close db: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sql expectations: %v", err)
		}
	}
}
