package backup

import (
	"context"
	"fmt"
)

// Service provides backup and restore operations.
type Service struct {
	repo            *Repository
	databasePath    string
	storageProvider string
	storageRoot     string
	configPath      string
	appVersion      string
	encryptionKey   []byte // nil means no encryption
}

// NewService creates a new backup service.
// encryptionKey is optional; when non-nil, backups are AES-256-GCM encrypted.
// The caller is responsible for ensuring it is exactly 32 bytes.
func NewService(repo *Repository, databasePath, storageProvider, storageRoot, configPath, appVersion string, encryptionKey []byte) *Service {
	return &Service{
		repo:            repo,
		databasePath:    databasePath,
		storageProvider: storageProvider,
		storageRoot:     storageRoot,
		configPath:      configPath,
		appVersion:      appVersion,
		encryptionKey:   encryptionKey,
	}
}

// CreateBackup creates a new backup archive.
func (s *Service) CreateBackup(ctx context.Context, includeConfig bool) (*BackupInfo, error) {
	if err := s.repo.EnsureDir(); err != nil {
		return nil, fmt.Errorf("ensure backup dir: %w", err)
	}

	name := GenerateName()
	backupPath := s.repo.Path(name)

	manifest, err := CreateArchive(ArchiveOptions{
		BackupPath:      backupPath,
		DatabasePath:    s.databasePath,
		StorageProvider: s.storageProvider,
		StorageRoot:     s.storageRoot,
		IncludeConfig:   includeConfig,
		ConfigPath:      s.configPath,
		AppVersion:      s.appVersion,
		EncryptionKey:   s.encryptionKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create backup: %w", err)
	}

	return &BackupInfo{
		Name:      name,
		Size:      manifest.DatabaseSizeBytes, // approximate
		CreatedAt: manifest.CreatedAt,
	}, nil
}

// ListBackups returns all available backups.
func (s *Service) ListBackups(ctx context.Context) ([]BackupInfo, error) {
	return s.repo.List()
}

// GetBackupPath returns the full path for a backup file.
func (s *Service) GetBackupPath(name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", err
	}

	exists, err := s.repo.Exists(name)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("backup not found: %s", name)
	}

	return s.repo.Path(name), nil
}

// DeleteBackup removes a backup file.
func (s *Service) DeleteBackup(ctx context.Context, name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	exists, err := s.repo.Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("backup not found: %s", name)
	}

	return s.repo.Delete(name)
}

// RestoreBackup restores from a backup archive.
func (s *Service) RestoreBackup(ctx context.Context, name string, dataDir string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	exists, err := s.repo.Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("backup not found: %s", name)
	}

	backupPath := s.repo.Path(name)
	return Restore(RestoreOptions{
		BackupPath:    backupPath,
		DataDir:       dataDir,
		EncryptionKey: s.encryptionKey,
	})
}
