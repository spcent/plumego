package system

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"cloud-vault/internal/config"
	"cloud-vault/internal/database"
	"cloud-vault/internal/storage"
)

func openTestDB(t *testing.T) *database.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func setupSystem(t *testing.T) (*Service, *database.DB, storage.ObjectStorage) {
	t.Helper()
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	cfg := config.Config{
		AI: config.AIConfig{Enabled: false},
	}
	svc := NewService(db.DB, store, cfg)
	return svc, db, store
}

func TestSystem_Health_AllOK(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Health(ctx)

	if result.Status != StatusOK {
		t.Errorf("Status = %q, want %q", result.Status, StatusOK)
	}
	if result.Database != StatusOK {
		t.Errorf("Database = %q, want %q", result.Database, StatusOK)
	}
	if result.Storage != StatusOK {
		t.Errorf("Storage = %q, want %q", result.Storage, StatusOK)
	}
	if result.Search != StatusOK {
		t.Errorf("Search = %q, want %q", result.Search, StatusOK)
	}
}

func TestSystem_Stats_EmptyDB(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	stats, err := svc.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}

	// All counts should be 0 for empty DB
	if stats.Documents != 0 {
		t.Errorf("Documents = %d, want 0", stats.Documents)
	}
	if stats.Versions != 0 {
		t.Errorf("Versions = %d, want 0", stats.Versions)
	}
	if stats.Collections != 0 {
		t.Errorf("Collections = %d, want 0", stats.Collections)
	}
	if stats.Tags != 0 {
		t.Errorf("Tags = %d, want 0", stats.Tags)
	}
}

func TestSystem_Doctor_AllChecks(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Doctor(ctx, DoctorRequest{})

	// Status may not be OK due to deployment checks (config, backup, etc.)
	// Just verify we got results

	// Should run all 18 checks (9 data + auth + 8 deployment)
	if len(result.Checks) != 18 {
		t.Errorf("Checks count = %d, want 18", len(result.Checks))
	}

	// Verify all checks present
	expectedChecks := map[string]bool{
		"storage_objects":        false,
		"document_versions":      false,
		"document_hash":          false,
		"fts_index":              false,
		"tags":                   false,
		"collections":            false,
		"sources":                false,
		"imports":                false,
		"ai":                     false,
		"auth":                   false,
		"config_check":           false,
		"data_dir_check":         false,
		"storage_writable_check": false,
		"auth_security_check":    false,
		"cookie_security_check":  false,
		"qiniu_config_check":     false,
		"backup_check":           false,
		"migration_check":        false,
	}

	for _, check := range result.Checks {
		if _, ok := expectedChecks[check.Name]; ok {
			expectedChecks[check.Name] = true
		}
	}

	for name, found := range expectedChecks {
		if !found {
			t.Errorf("Check %q not found", name)
		}
	}
}

func TestSystem_Doctor_SpecificChecks(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Doctor(ctx, DoctorRequest{
		Checks: []string{"storage_objects", "fts_index"},
	})

	if len(result.Checks) != 2 {
		t.Errorf("Checks count = %d, want 2", len(result.Checks))
	}
}

func TestSystem_Doctor_DetectsDanglingRefs(t *testing.T) {
	svc, db, _ := setupSystem(t)
	ctx := context.Background()

	// Insert a document
	_, err := db.Exec(`
		INSERT INTO documents (id, title, storage_key, content_hash, status, created_at, updated_at)
		VALUES ('doc-1', 'Test', 'key', 'hash', 'active', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
	`)
	if err != nil {
		t.Fatalf("Insert doc: %v", err)
	}

	// Insert a tag
	_, err = db.Exec(`
		INSERT INTO tags (id, name, created_at) VALUES ('tag-1', 'test', '2026-01-01T00:00:00Z')
	`)
	if err != nil {
		t.Fatalf("Insert tag: %v", err)
	}

	// Temporarily disable FK constraints to create a dangling reference
	_, err = db.Exec(`PRAGMA foreign_keys = OFF`)
	if err != nil {
		t.Fatalf("Disable FK: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO document_tags (document_id, tag_id) VALUES ('nonexistent-doc', 'tag-1')
	`)
	if err != nil {
		t.Fatalf("Insert dangling ref: %v", err)
	}
	_, err = db.Exec(`PRAGMA foreign_keys = ON`)
	if err != nil {
		t.Fatalf("Enable FK: %v", err)
	}

	// Run tags check
	result := svc.Doctor(ctx, DoctorRequest{
		Checks: []string{"tags"},
	})

	if len(result.Checks) != 1 {
		t.Fatalf("Checks count = %d, want 1", len(result.Checks))
	}

	check := result.Checks[0]
	if check.Name != "tags" {
		t.Errorf("Check name = %q, want 'tags'", check.Name)
	}
	if check.Status != StatusWarning {
		t.Errorf("Status = %q, want %q", check.Status, StatusWarning)
	}
	if check.Failed != 1 {
		t.Errorf("Failed = %d, want 1", check.Failed)
	}
	if len(check.Items) != 1 {
		t.Errorf("Items count = %d, want 1", len(check.Items))
	}
	if len(check.Items) > 0 && check.Items[0].Issue != "dangling_document_id" {
		t.Errorf("Issue = %q, want 'dangling_document_id'", check.Items[0].Issue)
	}
}

func TestSystem_Doctor_StorageCheck(t *testing.T) {
	svc, db, store := setupSystem(t)
	ctx := context.Background()

	// Insert a document
	_, err := db.Exec(`
		INSERT INTO documents (id, title, storage_key, content_hash, status, created_at, updated_at)
		VALUES ('doc-1', 'Test', 'docs/doc-1.md', 'hash', 'active', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
	`)
	if err != nil {
		t.Fatalf("Insert doc: %v", err)
	}

	// Don't create the storage object — should detect missing
	result := svc.Doctor(ctx, DoctorRequest{
		Checks: []string{"storage_objects"},
	})

	if len(result.Checks) != 1 {
		t.Fatalf("Checks count = %d, want 1", len(result.Checks))
	}

	check := result.Checks[0]
	if check.Status != StatusWarning {
		t.Errorf("Status = %q, want %q", check.Status, StatusWarning)
	}
	if check.Failed != 1 {
		t.Errorf("Failed = %d, want 1", check.Failed)
	}

	// Now create the storage object
	err = store.Put(ctx, "docs/doc-1.md", strings.NewReader(""), 0, "text/markdown")
	if err != nil {
		t.Fatalf("Put storage: %v", err)
	}

	// Re-run check
	result2 := svc.Doctor(ctx, DoctorRequest{
		Checks: []string{"storage_objects"},
	})

	check2 := result2.Checks[0]
	if check2.Status != StatusOK {
		t.Errorf("After put: Status = %q, want %q", check2.Status, StatusOK)
	}
	if check2.Failed != 0 {
		t.Errorf("After put: Failed = %d, want 0", check2.Failed)
	}
}
