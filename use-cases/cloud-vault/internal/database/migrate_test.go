package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrate_CreatesAllTables(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Expected tables from all migrations
	expectedTables := []string{
		"schema_migrations",
		"documents",
		"document_versions",
		"tags",
		"document_tags",
		"sync_jobs",
		"import_jobs",
		"import_job_items",
		"document_metadata",
		"document_fts",
		"document_index_status",
		"search_history",
		"document_similarity",
		"collections",
		"collection_documents",
		"document_sources",
		"tag_suggestions",
		"topics",
		"topic_documents",
		"organize_jobs",
		"document_fingerprints",
		"ai_tasks",
		"document_ai_summaries",
		"prompts",
		"document_chunks",
	}

	for _, table := range expectedTables {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&name)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Errorf("table %q not created", table)
			} else {
				t.Errorf("query table %q: %v", table, err)
			}
		}
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// First migration
	if err := db.Migrate(); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}

	// Second migration should succeed without error
	if err := db.Migrate(); err != nil {
		t.Fatalf("second Migrate (idempotent): %v", err)
	}

	// Verify all migrations recorded
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 migrations, got %d", count)
	}
}

func TestMigrate_SchemaMigrationsTable(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Check schema_migrations structure
	rows, err := db.Query("SELECT version, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		var appliedAt string
		if err := rows.Scan(&version, &appliedAt); err != nil {
			t.Fatalf("scan: %v", err)
		}
		versions = append(versions, version)
		if appliedAt == "" {
			t.Errorf("version %d has empty applied_at", version)
		}
	}

	if len(versions) != 5 {
		t.Errorf("expected 5 versions, got %d", len(versions))
	}

	// Verify versions are 1-5
	for i, v := range versions {
		if v != i+1 {
			t.Errorf("version %d at index %d", v, i)
		}
	}
}

func TestOpen_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Directory should exist
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("directory %q not created", dir)
	}

	// Database file should exist
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file %q not created", dbPath)
	}
}
