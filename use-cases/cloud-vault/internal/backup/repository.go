package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

// Repository manages backup files in a directory.
type Repository struct {
	backupDir string
}

// NewRepository creates a new backup repository.
func NewRepository(backupDir string) *Repository {
	return &Repository{backupDir: backupDir}
}

// validBackupName matches safe backup filenames.
var validBackupName = regexp.MustCompile(`^[a-zA-Z0-9._-]+\.zip$`)

// ValidateName checks if a backup name is safe.
func ValidateName(name string) error {
	if !validBackupName.MatchString(name) {
		return fmt.Errorf("invalid backup name: %s", name)
	}
	return nil
}

// EnsureDir creates the backup directory if it doesn't exist.
func (r *Repository) EnsureDir() error {
	return os.MkdirAll(r.backupDir, 0755)
}

// List returns all backup files in the directory.
func (r *Repository) List() ([]BackupInfo, error) {
	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, fmt.Errorf("read backup dir: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := ValidateName(entry.Name()); err != nil {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Name:      entry.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// Path returns the full path for a backup file.
func (r *Repository) Path(name string) string {
	return filepath.Join(r.backupDir, name)
}

// Exists checks if a backup file exists.
func (r *Repository) Exists(name string) (bool, error) {
	if err := ValidateName(name); err != nil {
		return false, err
	}
	_, err := os.Stat(r.Path(name))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Delete removes a backup file.
func (r *Repository) Delete(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	return os.Remove(r.Path(name))
}

// ReadManifest reads the manifest from a backup zip file.
func (r *Repository) ReadManifest(name string) (*BackupManifest, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}

	zr, err := zip.OpenReader(r.Path(name))
	if err != nil {
		return nil, fmt.Errorf("open backup zip: %w", err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name == "manifest.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("open manifest: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("read manifest: %w", err)
			}

			var manifest BackupManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				return nil, fmt.Errorf("parse manifest: %w", err)
			}

			return &manifest, nil
		}
	}

	return nil, fmt.Errorf("manifest.json not found in backup")
}

// GenerateName creates a backup filename based on the current time.
func GenerateName() string {
	return fmt.Sprintf("cloud-vault-backup-%s.zip", time.Now().Format("20060102-150405"))
}
