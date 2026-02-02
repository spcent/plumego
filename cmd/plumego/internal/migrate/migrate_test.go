package migrate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateMigrationFiles(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	migration, err := CreateMigrationFiles(dir, "Add Users Table", now)
	if err != nil {
		t.Fatalf("CreateMigrationFiles() error = %v", err)
	}

	if migration.Version != "20240102030405" {
		t.Errorf("Version = %q, want %q", migration.Version, "20240102030405")
	}

	if migration.Name != "add_users_table" {
		t.Errorf("Name = %q, want %q", migration.Name, "add_users_table")
	}

	if _, err := os.Stat(migration.UpPath); err != nil {
		t.Fatalf("up migration missing: %v", err)
	}

	if _, err := os.Stat(migration.DownPath); err != nil {
		t.Fatalf("down migration missing: %v", err)
	}

	if _, err := CreateMigrationFiles(dir, "!!!", now); err == nil {
		t.Fatalf("expected error for invalid name")
	}
}

func TestLoadMigrations(t *testing.T) {
	dir := t.TempDir()

	files := []string{
		"20240102030405_init_schema.up.sql",
		"20240102030405_init_schema.down.sql",
		"20240102030505_add_users.up.sql",
		"20240102030505_add_users.down.sql",
		"README.md",
	}

	for _, name := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("-- test"), 0o644); err != nil {
			t.Fatalf("failed to write file %s: %v", name, err)
		}
	}

	migrations, err := LoadMigrations(dir)
	if err != nil {
		t.Fatalf("LoadMigrations() error = %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("LoadMigrations() returned %d migrations, want 2", len(migrations))
	}

	if migrations[0].Version != "20240102030405" {
		t.Errorf("first migration version = %q, want %q", migrations[0].Version, "20240102030405")
	}

	if migrations[1].Version != "20240102030505" {
		t.Errorf("second migration version = %q, want %q", migrations[1].Version, "20240102030505")
	}

	if migrations[0].UpPath == "" || migrations[0].DownPath == "" {
		t.Errorf("expected up/down paths for first migration")
	}
}
