// Package backup provides backup and restore functionality for Cloud Vault.
package backup

import (
	"time"
)

// BackupManifest describes the contents of a backup archive.
type BackupManifest struct {
	App               string    `json:"app"`
	Version           string    `json:"version"`
	CreatedAt         time.Time `json:"created_at"`
	DatabasePath      string    `json:"database_path"`
	StorageProvider   string    `json:"storage_provider"`
	ObjectCount       int       `json:"object_count"`
	DatabaseSizeBytes int64     `json:"database_size_bytes"`
	ConfigSnapshot    bool      `json:"config_snapshot"`
}

// BackupInfo represents a backup file in the backup directory.
type BackupInfo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}
