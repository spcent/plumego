package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// TestMigration_UpgradePaths verifies that databases created at earlier schema
// versions can be upgraded to the latest version without errors.
func TestMigration_UpgradePaths(t *testing.T) {
	tests := []struct {
		name       string
		fixtureSQL string // SQL to initialize the database
		fromDesc   string // description of starting version
	}{
		{
			name:       "empty_to_latest",
			fixtureSQL: "", // empty database
			fromDesc:   "empty",
		},
		{
			name:       "v01_to_latest",
			fixtureSQL: loadFixture(t, "v01_schema.sql"),
			fromDesc:   "v01",
		},
		{
			name:       "v03_to_latest",
			fixtureSQL: loadFixture(t, "v03_schema.sql"),
			fromDesc:   "v03",
		},
		{
			name:       "v05_to_latest",
			fixtureSQL: loadFixture(t, "v05_schema.sql"),
			fromDesc:   "v05",
		},
		{
			name:       "v07_to_latest",
			fixtureSQL: loadFixture(t, "v07_schema.sql"),
			fromDesc:   "v07",
		},
		{
			name:       "latest_to_latest_idempotent",
			fixtureSQL: "", // will be migrated to latest first
			fromDesc:   "latest",
		},
	}

	// Expected tables in the latest schema
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
		"users",
		"user_sessions",
		"login_attempts",
		"security_events",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test.db")

			db, err := Open(dbPath)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			defer db.Close()

			// Initialize with fixture if provided
			if tt.fixtureSQL != "" {
				if _, err := db.Exec(tt.fixtureSQL); err != nil {
					t.Fatalf("load fixture %s: %v", tt.fromDesc, err)
				}
			}

			// For idempotent test, migrate once first
			if tt.fromDesc == "latest" {
				if err := db.Migrate(); err != nil {
					t.Fatalf("first Migrate (setup): %v", err)
				}
			}

			// Upgrade to latest
			if err := db.Migrate(); err != nil {
				t.Fatalf("Migrate from %s: %v", tt.fromDesc, err)
			}

			// Verify all expected tables exist
			for _, table := range expectedTables {
				var name string
				err := db.QueryRow(
					"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
					table,
				).Scan(&name)
				if err != nil {
					if err == sql.ErrNoRows {
						t.Errorf("table %q missing after upgrade from %s", table, tt.fromDesc)
					} else {
						t.Errorf("query table %q: %v", table, err)
					}
				}
			}

			// Verify critical columns exist in documents table
			docColumns := []string{
				"id", "title", "status", "storage_key", "content_hash",
				"source_type", "import_job_id", "review_status", "quality_score",
			}
			for _, col := range docColumns {
				if !columnExists(t, db, "documents", col) {
					t.Errorf("column documents.%s missing after upgrade from %s", col, tt.fromDesc)
				}
			}

			// Verify critical columns exist in users table
			userColumns := []string{
				"id", "username", "email", "password_hash", "role", "status",
			}
			for _, col := range userColumns {
				if !columnExists(t, db, "users", col) {
					t.Errorf("column users.%s missing after upgrade from %s", col, tt.fromDesc)
				}
			}

			// Verify migrations are recorded
			var migrationCount int
			if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount); err != nil {
				t.Fatalf("count migrations: %v", err)
			}
			if migrationCount != 9 {
				t.Errorf("expected 9 migrations, got %d (upgrade from %s)", migrationCount, tt.fromDesc)
			}

			// Test idempotency: run migrate again
			if err := db.Migrate(); err != nil {
				t.Fatalf("second Migrate (idempotent check): %v", err)
			}

			// Verify migration count unchanged
			var newCount int
			if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&newCount); err != nil {
				t.Fatalf("count migrations after idempotent run: %v", err)
			}
			if newCount != migrationCount {
				t.Errorf("migration count changed after idempotent run: %d -> %d", migrationCount, newCount)
			}
		})
	}
}

// loadFixture reads a SQL fixture from testdata/.
func loadFixture(t *testing.T, filename string) string {
	t.Helper()
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return string(data)
}

// columnExists checks if a column exists in a table.
func columnExists(t *testing.T, db *DB, table, column string) bool {
	t.Helper()
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		t.Fatalf("PRAGMA table_info(%s): %v", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("scan table_info: %v", err)
		}
		if name == column {
			return true
		}
	}
	return false
}
