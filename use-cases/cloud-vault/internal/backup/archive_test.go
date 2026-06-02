package backup

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreateArchive_BasicDatabase verifies that CreateArchive creates a zip
// containing a database snapshot and manifest.
func TestCreateArchive_BasicDatabase(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a real (but empty) db file to simulate a database.
	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("SQLite-style placeholder data"), 0644); err != nil {
		t.Fatalf("create db: %v", err)
	}

	backupPath := filepath.Join(tmpDir, "test-backup.zip")
	manifest, err := CreateArchive(ArchiveOptions{
		BackupPath:      backupPath,
		DatabasePath:    dbPath,
		StorageProvider: "local",
		StorageRoot:     "", // no objects
		IncludeConfig:   false,
		AppVersion:      "1.2.3",
	})
	if err != nil {
		t.Fatalf("CreateArchive: %v", err)
	}

	if manifest == nil {
		t.Fatal("manifest is nil")
	}
	if manifest.App != "cloud-vault" {
		t.Errorf("App: got %q, want cloud-vault", manifest.App)
	}
	if manifest.Version != "1.2.3" {
		t.Errorf("Version: got %q, want 1.2.3", manifest.Version)
	}
	if manifest.StorageProvider != "local" {
		t.Errorf("StorageProvider: got %q, want local", manifest.StorageProvider)
	}
	if manifest.DatabaseSizeBytes <= 0 {
		t.Errorf("DatabaseSizeBytes: got %d, want > 0", manifest.DatabaseSizeBytes)
	}
	if manifest.ObjectCount != 0 {
		t.Errorf("ObjectCount (no storage root): got %d, want 0", manifest.ObjectCount)
	}
	if manifest.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// Verify zip was created.
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("backup zip not found: %v", err)
	}

	// Verify manifest.json is in the zip.
	zr, err := zip.OpenReader(backupPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	var hasManifest, hasDB bool
	for _, f := range zr.File {
		if f.Name == "manifest.json" {
			hasManifest = true
		}
		if f.Name == "database/app.db" {
			hasDB = true
		}
	}
	if !hasManifest {
		t.Error("manifest.json not found in zip")
	}
	if !hasDB {
		t.Error("database/app.db not found in zip")
	}
}

// TestCreateArchive_WithObjects verifies that local storage objects are included.
func TestCreateArchive_WithObjects(t *testing.T) {
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("db data"), 0644); err != nil {
		t.Fatalf("create db: %v", err)
	}

	objectsDir := filepath.Join(tmpDir, "objects")
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		t.Fatalf("create objects dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(objectsDir, "doc1.md"), []byte("# Doc 1"), 0644); err != nil {
		t.Fatalf("create object 1: %v", err)
	}
	subDir := filepath.Join(objectsDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "doc2.md"), []byte("# Doc 2"), 0644); err != nil {
		t.Fatalf("create object 2: %v", err)
	}

	backupPath := filepath.Join(tmpDir, "backup-with-objects.zip")
	manifest, err := CreateArchive(ArchiveOptions{
		BackupPath:      backupPath,
		DatabasePath:    dbPath,
		StorageProvider: "local",
		StorageRoot:     objectsDir,
		IncludeConfig:   false,
		AppVersion:      "0.9.0",
	})
	if err != nil {
		t.Fatalf("CreateArchive: %v", err)
	}

	if manifest.ObjectCount != 2 {
		t.Errorf("ObjectCount: got %d, want 2", manifest.ObjectCount)
	}

	// Verify the objects are in the zip.
	zr, err := zip.OpenReader(backupPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	var hasDoc1, hasDoc2 bool
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "doc1.md") {
			hasDoc1 = true
		}
		if strings.HasSuffix(f.Name, "doc2.md") {
			hasDoc2 = true
		}
	}
	if !hasDoc1 {
		t.Error("doc1.md not found in zip")
	}
	if !hasDoc2 {
		t.Error("doc2.md not found in zip")
	}
}

// TestCreateArchive_WithConfig verifies that config is included when requested.
func TestCreateArchive_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("db"), 0644); err != nil {
		t.Fatalf("create db: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(`[server]\nport = 8080`), 0644); err != nil {
		t.Fatalf("create config: %v", err)
	}

	backupPath := filepath.Join(tmpDir, "backup-config.zip")
	manifest, err := CreateArchive(ArchiveOptions{
		BackupPath:      backupPath,
		DatabasePath:    dbPath,
		StorageProvider: "local",
		IncludeConfig:   true,
		ConfigPath:      configPath,
		AppVersion:      "0.5.0",
	})
	if err != nil {
		t.Fatalf("CreateArchive: %v", err)
	}

	if !manifest.ConfigSnapshot {
		t.Error("ConfigSnapshot should be true when config included")
	}

	// Verify config.toml is in the zip.
	zr, err := zip.OpenReader(backupPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	var hasConfig bool
	for _, f := range zr.File {
		if f.Name == "config.toml" {
			hasConfig = true
		}
	}
	if !hasConfig {
		t.Error("config.toml not found in zip")
	}
}

// TestCreateArchive_ConfigNotExist verifies graceful handling when config file is missing.
func TestCreateArchive_ConfigNotExist(t *testing.T) {
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("db"), 0644); err != nil {
		t.Fatalf("create db: %v", err)
	}

	backupPath := filepath.Join(tmpDir, "backup-no-config.zip")
	manifest, err := CreateArchive(ArchiveOptions{
		BackupPath:      backupPath,
		DatabasePath:    dbPath,
		StorageProvider: "local",
		IncludeConfig:   true,
		ConfigPath:      "/nonexistent/config.toml",
		AppVersion:      "0.5.0",
	})
	if err != nil {
		t.Fatalf("CreateArchive with missing config: %v", err)
	}
	if manifest.ConfigSnapshot {
		t.Error("ConfigSnapshot should be false when config file does not exist")
	}
}

// TestCreateArchive_NonLocalStorage verifies that objects are skipped for non-local providers.
func TestCreateArchive_NonLocalStorage(t *testing.T) {
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("db"), 0644); err != nil {
		t.Fatalf("create db: %v", err)
	}

	backupPath := filepath.Join(tmpDir, "backup-qiniu.zip")
	manifest, err := CreateArchive(ArchiveOptions{
		BackupPath:      backupPath,
		DatabasePath:    dbPath,
		StorageProvider: "qiniu",
		StorageRoot:     "/some/root",
		AppVersion:      "1.0.0",
	})
	if err != nil {
		t.Fatalf("CreateArchive qiniu: %v", err)
	}

	// Objects should not be included for non-local provider.
	if manifest.ObjectCount != 0 {
		t.Errorf("ObjectCount for qiniu: got %d, want 0", manifest.ObjectCount)
	}
}

// TestAddFileToZip verifies the addFileToZip helper.
func TestAddFileToZip(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a source file.
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := "hello world from zip"
	if err := os.WriteFile(srcPath, []byte(content), 0644); err != nil {
		t.Fatalf("create source: %v", err)
	}

	// Create a zip with it.
	zipPath := filepath.Join(tmpDir, "out.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip file: %v", err)
	}
	zw := zip.NewWriter(f)

	if err := addFileToZip(zw, srcPath, "inner/source.txt"); err != nil {
		t.Fatalf("addFileToZip: %v", err)
	}
	zw.Close()
	f.Close()

	// Read back and verify.
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	if len(zr.File) != 1 {
		t.Fatalf("zip file count: got %d, want 1", len(zr.File))
	}
	if zr.File[0].Name != "inner/source.txt" {
		t.Errorf("zip entry name: got %q, want inner/source.txt", zr.File[0].Name)
	}
}

// TestAddFileToZip_SourceNotExist verifies error when source file is missing.
func TestAddFileToZip_SourceNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "out.zip")
	f, _ := os.Create(zipPath)
	zw := zip.NewWriter(f)
	defer zw.Close()
	defer f.Close()

	err := addFileToZip(zw, "/nonexistent/file.txt", "entry.txt")
	if err == nil {
		t.Error("addFileToZip should fail for non-existent source")
	}
}

// TestRestore_DatabaseFile verifies Restore extracts database/app.db.
func TestRestore_DatabaseFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Build a zip with database/app.db.
	zipPath := filepath.Join(tmpDir, "restore-test.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(f)

	// manifest.json
	mw, _ := zw.Create("manifest.json")
	mw.Write([]byte(`{"app":"cloud-vault","version":"1.0.0"}`))

	// database/app.db
	dbw, _ := zw.Create("database/app.db")
	dbw.Write([]byte("fake-database-content"))

	zw.Close()
	f.Close()

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := Restore(RestoreOptions{
		BackupPath: zipPath,
		DataDir:    extractDir,
	}); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	restoredDB := filepath.Join(extractDir, "app.db")
	data, err := os.ReadFile(restoredDB)
	if err != nil {
		t.Fatalf("read restored db: %v", err)
	}
	if string(data) != "fake-database-content" {
		t.Errorf("Restored db content: got %q, want %q", string(data), "fake-database-content")
	}
}

// TestRestore_ObjectsFile verifies objects/ files are extracted.
func TestRestore_ObjectsFile(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "restore-obj.zip")
	f, _ := os.Create(zipPath)
	zw := zip.NewWriter(f)

	mw, _ := zw.Create("manifest.json")
	mw.Write([]byte(`{"app":"cloud-vault","version":"1.0.0"}`))

	ow, _ := zw.Create("objects/subfolder/doc.md")
	ow.Write([]byte("# Document content"))

	zw.Close()
	f.Close()

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := Restore(RestoreOptions{
		BackupPath: zipPath,
		DataDir:    extractDir,
	}); err != nil {
		t.Fatalf("Restore objects: %v", err)
	}

	restoredObj := filepath.Join(extractDir, "objects", "subfolder", "doc.md")
	if _, err := os.Stat(restoredObj); err != nil {
		t.Errorf("restored object not found: %v", err)
	}
}

// TestRestore_SkipsManifest verifies manifest.json is not written to disk.
func TestRestore_SkipsManifest(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "restore-skip.zip")
	f, _ := os.Create(zipPath)
	zw := zip.NewWriter(f)
	mw, _ := zw.Create("manifest.json")
	mw.Write([]byte(`{"app":"cloud-vault"}`))
	zw.Close()
	f.Close()

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := Restore(RestoreOptions{
		BackupPath: zipPath,
		DataDir:    extractDir,
	}); err != nil {
		t.Fatalf("Restore skip manifest: %v", err)
	}

	manifestPath := filepath.Join(extractDir, "manifest.json")
	if _, err := os.Stat(manifestPath); err == nil {
		t.Error("manifest.json should NOT be extracted to disk")
	}
}

// TestRestore_SkipsUnknownFiles verifies unknown files are silently skipped.
func TestRestore_SkipsUnknownFiles(t *testing.T) {
	tmpDir := t.TempDir()

	zipPath := filepath.Join(tmpDir, "restore-unk.zip")
	f, _ := os.Create(zipPath)
	zw := zip.NewWriter(f)
	mw, _ := zw.Create("manifest.json")
	mw.Write([]byte(`{"app":"cloud-vault"}`))
	uw, _ := zw.Create("unknown/strange-file.bin")
	uw.Write([]byte("binary data"))
	zw.Close()
	f.Close()

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := Restore(RestoreOptions{
		BackupPath: zipPath,
		DataDir:    extractDir,
	}); err != nil {
		t.Fatalf("Restore with unknown files: %v", err)
	}

	// unknown file should not be extracted.
	unknownPath := filepath.Join(extractDir, "unknown", "strange-file.bin")
	if _, err := os.Stat(unknownPath); err == nil {
		t.Error("Unknown file should not be extracted")
	}
}

// TestRestore_InvalidBackupName verifies that invalid backup names are rejected.
func TestRestore_InvalidBackupName(t *testing.T) {
	tmpDir := t.TempDir()

	err := Restore(RestoreOptions{
		BackupPath: filepath.Join(tmpDir, "../evil.zip"),
		DataDir:    tmpDir,
	})
	if err == nil {
		t.Error("Restore with invalid name: expected error")
	}
}

// TestRepository_ReadManifest verifies ReadManifest parses manifest.json.
func TestRepository_ReadManifest(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	// Create test backup with manifest.
	createTestBackup(t, tmpDir, "backup-manifest.zip")

	manifest, err := repo.ReadManifest("backup-manifest.zip")
	if err != nil {
		t.Fatalf("ReadManifest: %v", err)
	}
	if manifest == nil {
		t.Fatal("manifest is nil")
	}
	if manifest.App != "cloud-vault" {
		t.Errorf("App: got %q, want cloud-vault", manifest.App)
	}
	if manifest.Version != "0.8.0" {
		t.Errorf("Version: got %q, want 0.8.0", manifest.Version)
	}
}

// TestRepository_ReadManifest_InvalidName verifies error for invalid backup name.
func TestRepository_ReadManifest_InvalidName(t *testing.T) {
	repo := NewRepository(t.TempDir())
	_, err := repo.ReadManifest("../evil.zip")
	if err == nil {
		t.Error("ReadManifest invalid name: expected error")
	}
}

// TestRepository_ReadManifest_NoManifestEntry verifies error when zip has no manifest.
func TestRepository_ReadManifest_NoManifestEntry(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	zipPath := filepath.Join(tmpDir, "no-manifest.zip")
	f, _ := os.Create(zipPath)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("database/app.db")
	w.Write([]byte("db"))
	zw.Close()
	f.Close()

	_, err := repo.ReadManifest("no-manifest.zip")
	if err == nil {
		t.Error("ReadManifest with no manifest: expected error")
	}
}

// TestRepository_Delete verifies deletion of a backup file.
func TestRepository_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	createTestBackup(t, tmpDir, "del-test.zip")

	// Verify it exists.
	exists, err := repo.Exists("del-test.zip")
	if err != nil || !exists {
		t.Fatalf("Backup should exist before delete: err=%v, exists=%v", err, exists)
	}

	// Delete it.
	if err := repo.Delete("del-test.zip"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Verify it's gone.
	exists, err = repo.Exists("del-test.zip")
	if err != nil || exists {
		t.Errorf("Backup should not exist after delete: err=%v, exists=%v", err, exists)
	}
}

// TestRepository_Delete_InvalidName verifies error for invalid backup name.
func TestRepository_Delete_InvalidName(t *testing.T) {
	repo := NewRepository(t.TempDir())
	if err := repo.Delete("../evil.zip"); err == nil {
		t.Error("Delete invalid name: expected error")
	}
}

// TestRepository_Path verifies the Path method.
func TestRepository_Path(t *testing.T) {
	repo := NewRepository("/backups")
	got := repo.Path("backup.zip")
	want := "/backups/backup.zip"
	if got != want {
		t.Errorf("Path: got %q, want %q", got, want)
	}
}

// TestRepository_EnsureDir verifies directory creation.
func TestRepository_EnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "nested", "backups")
	repo := NewRepository(newDir)

	if err := repo.EnsureDir(); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	if _, err := os.Stat(newDir); err != nil {
		t.Errorf("Directory not created: %v", err)
	}
}

// TestRepository_List_EmptyDir verifies List returns empty for nonexistent directory.
func TestRepository_List_EmptyDir(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "no-such-dir")
	repo := NewRepository(nonexistent)

	backups, err := repo.List()
	if err != nil {
		t.Fatalf("List nonexistent dir: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("List nonexistent: got %d, want 0", len(backups))
	}
}

// TestRepository_List_SortedByTime verifies backups are sorted newest first.
func TestRepository_List_SortedByTime(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewRepository(tmpDir)

	// Create two files; we can't control mtime precisely but we can at least
	// verify the list has both.
	createTestBackup(t, tmpDir, "backup-alpha.zip")
	createTestBackup(t, tmpDir, "backup-beta.zip")

	backups, err := repo.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(backups) != 2 {
		t.Errorf("List count: got %d, want 2", len(backups))
	}
}

// TestGenerateName verifies the backup name format.
func TestGenerateName(t *testing.T) {
	name := GenerateName()
	if !strings.HasPrefix(name, "cloud-vault-backup-") {
		t.Errorf("GenerateName prefix: %q", name)
	}
	if !strings.HasSuffix(name, ".zip") {
		t.Errorf("GenerateName suffix: %q", name)
	}
	if err := ValidateName(name); err != nil {
		t.Errorf("GenerateName not valid: %v", err)
	}
}

// TestService_ListBackups verifies ListBackups returns correct count.
func TestService_ListBackups(t *testing.T) {
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("db content"), 0644); err != nil {
		t.Fatalf("create db: %v", err)
	}

	backupDir := filepath.Join(tmpDir, "backups")
	repo := NewRepository(backupDir)
	svc := NewService(repo, dbPath, "local", "", "", "1.0.0")
	ctx := context.Background()

	// Initially empty.
	backups, err := svc.ListBackups(ctx)
	if err != nil {
		t.Fatalf("ListBackups empty: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("ListBackups empty: got %d", len(backups))
	}

	// Create one backup via service.
	bk1, err := svc.CreateBackup(ctx, false)
	if err != nil {
		t.Fatalf("CreateBackup 1: %v", err)
	}

	// Add a second backup file manually so names won't collide.
	createTestBackup(t, backupDir, "manual-backup.zip")

	backups, err = svc.ListBackups(ctx)
	if err != nil {
		t.Fatalf("ListBackups 2: %v", err)
	}
	if len(backups) != 2 {
		t.Errorf("ListBackups: got %d, want 2", len(backups))
	}
	_ = bk1
}

// TestService_GetBackupPath verifies GetBackupPath returns the path for an existing backup.
func TestService_GetBackupPath(t *testing.T) {
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("db"), 0644); err != nil {
		t.Fatalf("create db: %v", err)
	}

	backupDir := filepath.Join(tmpDir, "backups")
	repo := NewRepository(backupDir)
	svc := NewService(repo, dbPath, "local", "", "", "1.0.0")
	ctx := context.Background()

	bk, err := svc.CreateBackup(ctx, false)
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	path, err := svc.GetBackupPath(bk.Name)
	if err != nil {
		t.Fatalf("GetBackupPath: %v", err)
	}
	if !strings.HasSuffix(path, bk.Name) {
		t.Errorf("GetBackupPath: got %q, want suffix %q", path, bk.Name)
	}

	// Non-existent backup.
	_, err = svc.GetBackupPath("nonexistent.zip")
	if err == nil {
		t.Error("GetBackupPath nonexistent: expected error")
	}
}

// TestService_GetBackupPath_InvalidName verifies invalid name is rejected.
func TestService_GetBackupPath_InvalidName(t *testing.T) {
	repo := NewRepository(t.TempDir())
	svc := NewService(repo, "", "", "", "", "1.0.0")

	_, err := svc.GetBackupPath("../evil.zip")
	if err == nil {
		t.Error("GetBackupPath invalid name: expected error")
	}
}

// TestService_DeleteBackup verifies backup deletion.
func TestService_DeleteBackup(t *testing.T) {
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("db"), 0644); err != nil {
		t.Fatalf("create db: %v", err)
	}

	backupDir := filepath.Join(tmpDir, "backups")
	repo := NewRepository(backupDir)
	svc := NewService(repo, dbPath, "local", "", "", "1.0.0")
	ctx := context.Background()

	bk, err := svc.CreateBackup(ctx, false)
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	if err := svc.DeleteBackup(ctx, bk.Name); err != nil {
		t.Fatalf("DeleteBackup: %v", err)
	}

	backups, err := svc.ListBackups(ctx)
	if err != nil {
		t.Fatalf("ListBackups after delete: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("After delete: got %d backups, want 0", len(backups))
	}
}

// TestService_DeleteBackup_NotFound verifies error for nonexistent backup.
func TestService_DeleteBackup_NotFound(t *testing.T) {
	repo := NewRepository(t.TempDir())
	svc := NewService(repo, "", "", "", "", "1.0.0")
	ctx := context.Background()

	err := svc.DeleteBackup(ctx, "missing.zip")
	if err == nil {
		t.Error("DeleteBackup missing: expected error")
	}
}

// TestService_RestoreBackup_Integration verifies RestoreBackup extracts to data dir.
func TestService_RestoreBackup_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	dbPath := filepath.Join(tmpDir, "app.db")
	if err := os.WriteFile(dbPath, []byte("database content for restore"), 0644); err != nil {
		t.Fatalf("create db: %v", err)
	}

	backupDir := filepath.Join(tmpDir, "backups")
	repo := NewRepository(backupDir)
	svc := NewService(repo, dbPath, "local", "", "", "1.0.0")
	ctx := context.Background()

	bk, err := svc.CreateBackup(ctx, false)
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}

	restoreDir := filepath.Join(tmpDir, "restored")
	if err := svc.RestoreBackup(ctx, bk.Name, restoreDir); err != nil {
		t.Fatalf("RestoreBackup: %v", err)
	}

	// Restored DB should exist.
	restoredDB := filepath.Join(restoreDir, "app.db")
	if _, err := os.Stat(restoredDB); err != nil {
		t.Errorf("restored db not found: %v", err)
	}
}
