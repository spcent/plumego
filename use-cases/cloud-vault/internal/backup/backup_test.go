package backup

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"backup.zip", false},
		{"cloud-vault-backup-20260529-120000.zip", false},
		{"test_backup.zip", false},
		{"../evil.zip", true},
		{"path/to/backup.zip", true},
		{"backup.exe", true},
		{"backup", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestRepository_List(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	// Empty directory
	backups, err := repo.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("List() returned %d backups, want 0", len(backups))
	}

	// Create some backup files
	createTestBackup(t, tmpDir, "backup1.zip")
	createTestBackup(t, tmpDir, "backup2.zip")

	backups, err = repo.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(backups) != 2 {
		t.Errorf("List() returned %d backups, want 2", len(backups))
	}
}

func TestRepository_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	// Non-existent
	exists, err := repo.Exists("missing.zip")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() = true for missing file")
	}

	// Create file
	createTestBackup(t, tmpDir, "backup.zip")

	exists, err = repo.Exists("backup.zip")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false for existing file")
	}

	// Invalid name
	_, err = repo.Exists("../evil.zip")
	if err == nil {
		t.Error("Exists() should reject invalid name")
	}
}

func TestCreateArchive_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test database
	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("test database content"), 0644); err != nil {
		t.Fatalf("create test db: %v", err)
	}

	// Create test objects
	objectsDir := filepath.Join(tmpDir, "objects")
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		t.Fatalf("create objects dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(objectsDir, "file1.md"), []byte("# Test"), 0644); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	// Create backup
	backupPath := filepath.Join(tmpDir, "backup.zip")
	manifest, err := CreateArchive(ArchiveOptions{
		BackupPath:      backupPath,
		DatabasePath:    dbPath,
		StorageProvider: "local",
		StorageRoot:     objectsDir,
		IncludeConfig:   false,
		AppVersion:      "0.8.0",
	})
	if err != nil {
		t.Fatalf("CreateArchive() error = %v", err)
	}

	// Verify manifest
	if manifest.App != "cloud-vault" {
		t.Errorf("manifest.App = %q, want cloud-vault", manifest.App)
	}
	if manifest.Version != "0.8.0" {
		t.Errorf("manifest.Version = %q, want 0.8.0", manifest.Version)
	}
	if manifest.ObjectCount != 1 {
		t.Errorf("manifest.ObjectCount = %d, want 1", manifest.ObjectCount)
	}

	// Extract and verify
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := Restore(RestoreOptions{
		BackupPath: backupPath,
		DataDir:    extractDir,
	}); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	// Check database was restored
	restoredDB := filepath.Join(extractDir, "app.db")
	if _, err := os.Stat(restoredDB); err != nil {
		t.Errorf("restored database not found: %v", err)
	}

	// Check objects were restored
	restoredObj := filepath.Join(extractDir, "objects", "file1.md")
	if _, err := os.Stat(restoredObj); err != nil {
		t.Errorf("restored object not found: %v", err)
	}
}

func TestRestore_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a malicious zip with path traversal
	maliciousZip := filepath.Join(tmpDir, "malicious.zip")
	f, err := os.Create(maliciousZip)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	zw := zip.NewWriter(f)
	w, err := zw.Create("../../../etc/passwd")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	w.Write([]byte("malicious"))
	zw.Close()
	f.Close()

	// Try to restore - should fail
	extractDir := filepath.Join(tmpDir, "extracted")
	err = Restore(RestoreOptions{
		BackupPath: maliciousZip,
		DataDir:    extractDir,
	})
	if err == nil {
		t.Error("Restore() should reject path traversal")
	}
}

func TestService_CreateBackup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test database
	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("test db"), 0644); err != nil {
		t.Fatalf("create test db: %v", err)
	}

	// Create test objects
	objectsDir := filepath.Join(tmpDir, "objects")
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		t.Fatalf("create objects dir: %v", err)
	}

	backupDir := filepath.Join(tmpDir, "backups")
	repo := NewRepository(backupDir)
	svc := NewService(repo, dbPath, "local", objectsDir, "", "0.8.0")

	ctx := context.Background()
	backup, err := svc.CreateBackup(ctx, false)
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	if !strings.HasPrefix(backup.Name, "cloud-vault-backup-") {
		t.Errorf("backup.Name = %q, want prefix cloud-vault-backup-", backup.Name)
	}

	// List backups
	backups, err := svc.ListBackups(ctx)
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("ListBackups() returned %d backups, want 1", len(backups))
	}
}

// createTestBackup creates a minimal test backup file.
func createTestBackup(t *testing.T, dir, name string) {
	t.Helper()

	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create backup file: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	w, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatalf("create manifest: %v", err)
	}
	w.Write([]byte(`{"app":"cloud-vault","version":"0.8.0"}`))
}
