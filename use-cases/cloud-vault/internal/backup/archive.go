package backup

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DeriveEncryptionKey hashes passphrase down to a fixed 32-byte AES-256 key,
// regardless of the input string's length.
func DeriveEncryptionKey(passphrase string) []byte {
	sum := sha256.Sum256([]byte(passphrase))
	return sum[:]
}

// ArchiveOptions configures backup archive creation.
type ArchiveOptions struct {
	BackupPath      string
	DatabasePath    string
	StorageProvider string
	StorageRoot     string // only used when provider is "local"
	IncludeConfig   bool
	ConfigPath      string // path to config.toml
	AppVersion      string
	// EncryptionKey enables AES-256-GCM encryption when set (must be exactly 32 bytes).
	EncryptionKey []byte
}

// CreateArchive creates a backup zip archive.
func CreateArchive(opts ArchiveOptions) (*BackupManifest, error) {
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(opts.BackupPath), 0755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}

	// Create the zip file
	outFile, err := os.Create(opts.BackupPath)
	if err != nil {
		return nil, fmt.Errorf("create backup file: %w", err)
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	manifest := &BackupManifest{
		App:             "cloud-vault",
		Version:         opts.AppVersion,
		CreatedAt:       time.Now(),
		DatabasePath:    opts.DatabasePath,
		StorageProvider: opts.StorageProvider,
	}

	// 1. Backup database using VACUUM INTO
	dbSnapshot, err := createDatabaseSnapshot(opts.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("create db snapshot: %w", err)
	}
	defer os.Remove(dbSnapshot)

	// Get DB size
	dbInfo, err := os.Stat(dbSnapshot)
	if err != nil {
		return nil, fmt.Errorf("stat db snapshot: %w", err)
	}
	manifest.DatabaseSizeBytes = dbInfo.Size()

	// Add DB to zip
	if err := addFileToZip(zw, dbSnapshot, "database/app.db"); err != nil {
		return nil, fmt.Errorf("add database to zip: %w", err)
	}

	// 2. Backup objects (only for local storage)
	if opts.StorageProvider == "local" && opts.StorageRoot != "" {
		count, err := addDirectoryToZip(zw, opts.StorageRoot, "objects")
		if err != nil {
			return nil, fmt.Errorf("add objects to zip: %w", err)
		}
		manifest.ObjectCount = count
	}

	// 3. Optionally include config
	if opts.IncludeConfig && opts.ConfigPath != "" {
		if _, err := os.Stat(opts.ConfigPath); err == nil {
			if err := addFileToZip(zw, opts.ConfigPath, "config.toml"); err != nil {
				return nil, fmt.Errorf("add config to zip: %w", err)
			}
			manifest.ConfigSnapshot = true
		}
	}

	// 4. Write manifest as the first entry (we'll prepend it)
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}

	manifestWriter, err := zw.Create("manifest.json")
	if err != nil {
		return nil, fmt.Errorf("create manifest in zip: %w", err)
	}
	if _, err := manifestWriter.Write(manifestData); err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}

	// Optionally encrypt the backup in-place.
	if len(opts.EncryptionKey) > 0 {
		if err := encryptFile(opts.BackupPath, opts.EncryptionKey); err != nil {
			return nil, fmt.Errorf("encrypt backup: %w", err)
		}
		manifest.Encrypted = true
	}

	return manifest, nil
}

// encryptFile encrypts a file in-place using AES-256-GCM.
// The encrypted format is: 12-byte nonce || GCM ciphertext.
func encryptFile(path string, key []byte) error {
	plaintext, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return os.WriteFile(path, ciphertext, 0600)
}

// DecryptBackup decrypts a backup file previously encrypted with encryptFile.
// key must be exactly 32 bytes (AES-256). The decrypted ZIP is written to dst.
func DecryptBackup(src, dst string, key []byte) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read encrypted backup: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create GCM: %w", err)
	}
	if len(data) < gcm.NonceSize() {
		return fmt.Errorf("encrypted backup too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decrypt backup: wrong key or corrupted data")
	}
	return os.WriteFile(dst, plaintext, 0600)
}

// createDatabaseSnapshot creates a consistent snapshot of the SQLite database.
func createDatabaseSnapshot(dbPath string) (string, error) {
	// Create a temporary file for the snapshot
	tmpFile, err := os.CreateTemp("", "cloud-vault-backup-*.db")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("open database for backup: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA schema_version"); err != nil {
		if copyErr := copyFile(dbPath, tmpPath); copyErr != nil {
			os.Remove(tmpPath)
			return "", fmt.Errorf("snapshot database: %w; fallback copy: %w", err, copyErr)
		}
		return tmpPath, nil
	}

	if _, err := db.Exec("VACUUM INTO ?", tmpPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("vacuum database snapshot: %w", err)
	}

	return tmpPath, nil
}

// addFileToZip adds a single file to the zip archive.
func addFileToZip(zw *zip.Writer, srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := zw.Create(dstPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	return err
}

// addDirectoryToZip adds a directory recursively to the zip archive.
func addDirectoryToZip(zw *zip.Writer, srcDir, dstDir string) (int, error) {
	count := 0
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, relPath)
		if err := addFileToZip(zw, path, dstPath); err != nil {
			return err
		}
		count++
		return nil
	})

	return count, err
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
