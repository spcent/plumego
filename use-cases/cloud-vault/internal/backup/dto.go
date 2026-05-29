package backup

// CreateBackupRequest represents a request to create a backup.
type CreateBackupRequest struct {
	IncludeConfig bool `json:"include_config"`
}

// CreateBackupResponse represents the response after creating a backup.
type CreateBackupResponse struct {
	Backup BackupInfo `json:"backup"`
}

// ListBackupsResponse represents the response when listing backups.
type ListBackupsResponse struct {
	Backups []BackupInfo `json:"backups"`
}

// RestoreRequest represents a request to restore from a backup.
type RestoreRequest struct {
	BackupName string `json:"backup_name"`
	Confirm    string `json:"confirm"`
}

// RestoreResponse represents the response after a restore operation.
type RestoreResponse struct {
	Message string `json:"message"`
}
