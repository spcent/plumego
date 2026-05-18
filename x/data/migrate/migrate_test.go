package migrate

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewMissingMigrationsDirectoryReturnsError(t *testing.T) {
	db := openSQLite(t)
	defer db.Close()

	_, err := New(db, filepath.Join(t.TempDir(), "missing"))
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestEmptyMigrationsStatusReturnsEmptySlice(t *testing.T) {
	db := openSQLite(t)
	defer db.Close()

	migrator := newMigrator(t, db, t.TempDir())
	defer migrator.Close()

	status, err := migrator.Status(t.Context())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if len(status) != 0 {
		t.Fatalf("status length = %d, want 0", len(status))
	}
}

func TestUpAppliesPendingMigration(t *testing.T) {
	db := openSQLite(t)
	defer db.Close()
	dir := t.TempDir()
	writeMigration(t, dir)
	migrator := newMigrator(t, db, dir)
	defer migrator.Close()

	if err := migrator.Up(t.Context()); err != nil {
		t.Fatalf("up: %v", err)
	}

	status := mustStatus(t, migrator)
	if len(status) != 1 {
		t.Fatalf("status length = %d, want 1", len(status))
	}
	if status[0].State != "applied" || status[0].AppliedAt == nil {
		t.Fatalf("status = %#v, want applied with AppliedAt", status[0])
	}
	if !tableExists(t, db, "widgets") {
		t.Fatal("expected widgets table to exist")
	}
}

func TestDownRollsBackAppliedMigration(t *testing.T) {
	db := openSQLite(t)
	defer db.Close()
	dir := t.TempDir()
	writeMigration(t, dir)
	migrator := newMigrator(t, db, dir)
	defer migrator.Close()

	if err := migrator.Up(t.Context()); err != nil {
		t.Fatalf("up: %v", err)
	}
	if err := migrator.Down(t.Context()); err != nil {
		t.Fatalf("down: %v", err)
	}

	status := mustStatus(t, migrator)
	if len(status) != 1 {
		t.Fatalf("status length = %d, want 1", len(status))
	}
	if status[0].State != "pending" || status[0].AppliedAt != nil {
		t.Fatalf("status = %#v, want pending without AppliedAt", status[0])
	}
	if tableExists(t, db, "widgets") {
		t.Fatal("expected widgets table to be dropped")
	}
}

func openSQLite(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func newMigrator(t *testing.T, db *sql.DB, dir string) *Migrator {
	t.Helper()

	migrator, err := New(db, dir)
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}
	return migrator
}

func writeMigration(t *testing.T, dir string) {
	t.Helper()

	body := `-- +goose Up
CREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT NOT NULL);

-- +goose Down
DROP TABLE widgets;
`
	path := filepath.Join(dir, "00001_create_widgets.sql")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write migration: %v", err)
	}
}

func mustStatus(t *testing.T, migrator *Migrator) []MigrationStatus {
	t.Helper()

	status, err := migrator.Status(t.Context())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	return status
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()

	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&count)
	if err != nil {
		t.Fatalf("check table: %v", err)
	}
	return count > 0
}
