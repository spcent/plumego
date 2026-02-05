package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage implements Storage interface using the local filesystem.
type LocalStorage struct {
	basePath  string
	baseURL   string
	metadata  MetadataManager
	imageProc ImageProcessor
}

// NewLocalStorage creates a new local filesystem storage.
func NewLocalStorage(basePath, baseURL string, metadata MetadataManager) (*LocalStorage, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, &Error{
			Op:   "NewLocalStorage",
			Path: basePath,
			Err:  err,
		}
	}

	return &LocalStorage{
		basePath:  basePath,
		baseURL:   baseURL,
		metadata:  metadata,
		imageProc: NewImageProcessor(),
	}, nil
}

// Put uploads a file to local storage.
func (s *LocalStorage) Put(ctx context.Context, opts PutOptions) (*File, error) {
	// Generate file ID
	fileID := generateID()

	// Get file extension
	ext := filepath.Ext(opts.FileName)
	if ext == "" && opts.ContentType != "" {
		ext = mimeToExt(opts.ContentType)
	}

	// Generate storage path: basePath/tenant_id/2026/02/05/uuid.ext
	now := time.Now()
	relativePath := filepath.Join(
		opts.TenantID,
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		fileID+ext,
	)

	fullPath := filepath.Join(s.basePath, relativePath)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, &Error{
			Op:   "Put",
			Path: fullPath,
			Err:  err,
		}
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp(dir, "upload-*")
	if err != nil {
		return nil, &Error{
			Op:   "Put",
			Path: fullPath,
			Err:  err,
		}
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Calculate hash and write to temp file
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(tmpFile, hash), opts.Reader)
	if err != nil {
		tmpFile.Close()
		return nil, &Error{
			Op:   "Put",
			Path: fullPath,
			Err:  err,
		}
	}
	tmpFile.Close()

	hashString := hex.EncodeToString(hash.Sum(nil))

	// Check for deduplication
	if s.metadata != nil {
		existing, err := s.metadata.GetByHash(ctx, hashString)
		if err == nil && existing != nil {
			// File already exists, return existing metadata
			return existing, nil
		}
	}

	// Move to final location
	if err := os.Rename(tmpPath, fullPath); err != nil {
		return nil, &Error{
			Op:   "Put",
			Path: fullPath,
			Err:  err,
		}
	}

	// Build file metadata
	file := &File{
		ID:          fileID,
		TenantID:    opts.TenantID,
		Name:        opts.FileName,
		Path:        relativePath,
		Size:        size,
		MimeType:    opts.ContentType,
		Extension:   ext,
		Hash:        hashString,
		StorageType: "local",
		Metadata:    opts.Metadata,
		UploadedBy:  opts.UploadedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// If image, get dimensions
	if opts.ContentType != "" && s.imageProc.IsImage(opts.ContentType) {
		f, err := os.Open(fullPath)
		if err == nil {
			imgInfo, err := s.imageProc.GetInfo(f)
			f.Close()
			if err == nil {
				file.Width = imgInfo.Width
				file.Height = imgInfo.Height
			}
		}

		// Generate thumbnail if requested
		if opts.GenerateThumb {
			thumbPath, err := s.generateThumbnail(fullPath, relativePath, opts.ThumbWidth, opts.ThumbHeight)
			if err == nil {
				file.ThumbnailPath = thumbPath
			}
		}
	}

	// Save metadata
	if s.metadata != nil {
		if err := s.metadata.Save(ctx, file); err != nil {
			// Rollback: delete uploaded file
			os.Remove(fullPath)
			return nil, err
		}
	}

	return file, nil
}

// Get retrieves a file from local storage.
func (s *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	// Safety check: prevent path traversal
	if !isPathSafe(path) {
		return nil, &Error{
			Op:   "Get",
			Path: path,
			Err:  ErrInvalidPath,
		}
	}

	fullPath := filepath.Join(s.basePath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &Error{
				Op:   "Get",
				Path: path,
				Err:  ErrNotFound,
			}
		}
		return nil, &Error{
			Op:   "Get",
			Path: path,
			Err:  err,
		}
	}

	return file, nil
}

// Delete removes a file from local storage.
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	if !isPathSafe(path) {
		return &Error{
			Op:   "Delete",
			Path: path,
			Err:  ErrInvalidPath,
		}
	}

	fullPath := filepath.Join(s.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return &Error{
				Op:   "Delete",
				Path: path,
				Err:  ErrNotFound,
			}
		}
		return &Error{
			Op:   "Delete",
			Path: path,
			Err:  err,
		}
	}

	return nil
}

// Exists checks if a file exists in local storage.
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	if !isPathSafe(path) {
		return false, ErrInvalidPath
	}

	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Stat returns file information from local storage.
func (s *LocalStorage) Stat(ctx context.Context, path string) (*FileStat, error) {
	if !isPathSafe(path) {
		return nil, ErrInvalidPath
	}

	fullPath := filepath.Join(s.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &FileStat{
		Path:         path,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
		IsDir:        info.IsDir(),
	}, nil
}

// List returns a list of files in local storage matching the prefix.
func (s *LocalStorage) List(ctx context.Context, prefix string, limit int) ([]*FileStat, error) {
	if !isPathSafe(prefix) {
		return nil, ErrInvalidPath
	}

	fullPath := filepath.Join(s.basePath, prefix)

	var results []*FileStat
	count := 0

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if limit > 0 && count >= limit {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(s.basePath, path)
			if err != nil {
				return err
			}

			results = append(results, &FileStat{
				Path:         relPath,
				Size:         info.Size(),
				ModifiedTime: info.ModTime(),
				IsDir:        false,
			})
			count++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetURL returns a static URL for accessing the file.
func (s *LocalStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	// Local storage doesn't support expiring URLs, return static URL
	return s.baseURL + "/" + path, nil
}

// Copy copies a file within local storage.
func (s *LocalStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
	if !isPathSafe(srcPath) || !isPathSafe(dstPath) {
		return ErrInvalidPath
	}

	srcFullPath := filepath.Join(s.basePath, srcPath)
	dstFullPath := filepath.Join(s.basePath, dstPath)

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dstFullPath), 0755); err != nil {
		return err
	}

	// Open source file
	src, err := os.Open(srcFullPath)
	if err != nil {
		return err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(dstFullPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy content
	_, err = io.Copy(dst, src)
	return err
}

// generateThumbnail generates a thumbnail for an image file.
func (s *LocalStorage) generateThumbnail(srcPath, relativePath string, width, height int) (string, error) {
	if width <= 0 {
		width = 200
	}
	if height <= 0 {
		height = 200
	}

	// Open source file
	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate thumbnail
	thumbReader, err := s.imageProc.Thumbnail(src, width, height)
	if err != nil {
		return "", err
	}

	// Generate thumbnail path
	ext := filepath.Ext(relativePath)
	thumbRelPath := strings.TrimSuffix(relativePath, ext) + "_thumb" + ext
	thumbFullPath := filepath.Join(s.basePath, thumbRelPath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(thumbFullPath), 0755); err != nil {
		return "", err
	}

	// Save thumbnail
	thumbFile, err := os.Create(thumbFullPath)
	if err != nil {
		return "", err
	}
	defer thumbFile.Close()

	if _, err := io.Copy(thumbFile, thumbReader); err != nil {
		os.Remove(thumbFullPath)
		return "", err
	}

	return thumbRelPath, nil
}
