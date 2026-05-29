package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RestoreOptions configures backup restoration.
type RestoreOptions struct {
	BackupPath string
	DataDir    string
}

// Restore extracts a backup archive to the data directory.
func Restore(opts RestoreOptions) error {
	// Validate backup name
	if err := ValidateName(filepath.Base(opts.BackupPath)); err != nil {
		return err
	}

	// Open the zip file
	zr, err := zip.OpenReader(opts.BackupPath)
	if err != nil {
		return fmt.Errorf("open backup zip: %w", err)
	}
	defer zr.Close()

	// Create data directory if needed
	if err := os.MkdirAll(opts.DataDir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	// Extract each file
	for _, f := range zr.File {
		if err := extractFile(f, opts.DataDir); err != nil {
			return fmt.Errorf("extract %s: %w", f.Name, err)
		}
	}

	return nil
}

// extractFile extracts a single file from the zip archive.
func extractFile(f *zip.File, dataDir string) error {
	// Skip manifest.json - it's metadata only
	if f.Name == "manifest.json" {
		return nil
	}

	// Path traversal protection: reject paths that escape dataDir
	cleanName := filepath.Clean(f.Name)
	if strings.HasPrefix(cleanName, "..") || strings.Contains(cleanName, "/..") {
		return fmt.Errorf("path traversal detected: %s", f.Name)
	}

	// Determine the target path
	var targetPath string
	switch {
	case f.Name == "database/app.db":
		targetPath = filepath.Join(dataDir, "app.db")
	case len(f.Name) > 8 && f.Name[:8] == "objects/":
		targetPath = filepath.Join(dataDir, f.Name)
	case f.Name == "config.toml":
		targetPath = filepath.Join(dataDir, "config.toml")
	default:
		// Skip unknown files
		return nil
	}

	// Additional path traversal check on final target
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("resolve target path: %w", err)
	}
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return fmt.Errorf("resolve data dir: %w", err)
	}
	if !strings.HasPrefix(absTarget, absDataDir) {
		return fmt.Errorf("path traversal detected: %s escapes %s", f.Name, dataDir)
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Open the file in the zip
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create the target file
	dst, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the contents
	if _, err := io.Copy(dst, rc); err != nil {
		return err
	}

	return nil
}

// RestoreCLI implements the CLI restore subcommand.
func RestoreCLI(backupFile, dataDir string) error {
	// Validate the backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	fmt.Printf("Restoring from %s to %s\n", backupFile, dataDir)

	// Extract to temporary locations first
	dbPath := filepath.Join(dataDir, "app.db")
	dbNewPath := dbPath + ".new"
	objectsDir := filepath.Join(dataDir, "objects")
	objectsNewDir := objectsDir + ".new"

	// Create a temporary restore directory
	tmpDir, err := os.MkdirTemp("", "cloud-vault-restore-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract to temp directory
	if err := Restore(RestoreOptions{
		BackupPath: backupFile,
		DataDir:    tmpDir,
	}); err != nil {
		return fmt.Errorf("extract backup: %w", err)
	}

	// Read manifest to check storage provider
	manifest, err := NewRepository(filepath.Dir(backupFile)).ReadManifest(filepath.Base(backupFile))
	if err != nil {
		fmt.Printf("Warning: could not read manifest: %v\n", err)
	}

	// Move database
	tmpDB := filepath.Join(tmpDir, "app.db")
	if _, err := os.Stat(tmpDB); err == nil {
		fmt.Println("Restoring database...")
		if err := os.Rename(tmpDB, dbNewPath); err != nil {
			return fmt.Errorf("move database: %w", err)
		}
		// Atomic rename
		if err := os.Rename(dbNewPath, dbPath); err != nil {
			return fmt.Errorf("rename database: %w", err)
		}
	}

	// Move objects (only if local storage)
	tmpObjects := filepath.Join(tmpDir, "objects")
	if _, err := os.Stat(tmpObjects); err == nil {
		fmt.Println("Restoring objects...")
		if err := os.Rename(tmpObjects, objectsNewDir); err != nil {
			return fmt.Errorf("move objects: %w", err)
		}
		// Remove old objects directory if it exists
		if _, err := os.Stat(objectsDir); err == nil {
			if err := os.RemoveAll(objectsDir); err != nil {
				return fmt.Errorf("remove old objects: %w", err)
			}
		}
		// Atomic rename
		if err := os.Rename(objectsNewDir, objectsDir); err != nil {
			return fmt.Errorf("rename objects: %w", err)
		}
	}

	// Warn about Qiniu
	if manifest != nil && manifest.StorageProvider == "qiniu" {
		fmt.Println("\nWarning: This backup was created with Qiniu storage.")
		fmt.Println("Objects must be restored from Qiniu separately.")
	}

	fmt.Println("\nRestore complete. Run 'cloud-vault doctor' to verify.")
	return nil
}
