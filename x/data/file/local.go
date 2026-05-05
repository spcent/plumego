package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	storefile "github.com/spcent/plumego/store/file"
)

// LocalStorage implements Storage using the local filesystem.
// Files are organised as: basePath/{tenantID}/{YYYY}/{MM}/{DD}/{id}{ext}
type LocalStorage struct {
	basePath      string
	baseURL       string
	metadata      MetadataManager
	imageProc     *imageProcessor
	maxUploadSize int64
}

// NewLocalStorage creates a new local filesystem storage.
func NewLocalStorage(basePath, baseURL string, metadata MetadataManager) (*LocalStorage, error) {
	return NewLocalStorageWithConfig(basePath, baseURL, metadata, LocalConfig{})
}

// NewLocalStorageWithConfig creates a new local filesystem storage with provider-specific config.
func NewLocalStorageWithConfig(basePath, baseURL string, metadata MetadataManager, config LocalConfig) (*LocalStorage, error) {
	if config.MaxUploadSize <= 0 {
		config.MaxUploadSize = DefaultLocalMaxUploadSize
	}
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, &storefile.Error{
			Op:   "NewLocalStorage",
			Path: basePath,
			Err:  err,
		}
	}

	return &LocalStorage{
		basePath:      basePath,
		baseURL:       baseURL,
		metadata:      metadata,
		imageProc:     newImageProcessor(),
		maxUploadSize: config.MaxUploadSize,
	}, nil
}

// Put uploads a file to local storage under the tenant's directory tree.
func (s *LocalStorage) Put(ctx context.Context, opts PutOptions) (*File, error) {
	tenantID := strings.TrimSpace(opts.TenantID)
	if !isPathComponentSafe(tenantID) {
		return nil, &storefile.Error{Op: "Put", Path: opts.TenantID, Err: storefile.ErrInvalidPath}
	}
	if err := contextError(ctx); err != nil {
		return nil, &storefile.Error{Op: "Put", Path: tenantID, Err: err}
	}

	fileID, err := generateID()
	if err != nil {
		return nil, &storefile.Error{Op: "Put", Path: tenantID, Err: err}
	}

	ext := filepath.Ext(opts.FileName)
	if ext == "" && opts.ContentType != "" {
		ext = mimeToExt(opts.ContentType)
	}

	// Path: {tenantID}/{YYYY}/{MM}/{DD}/{id}{ext}
	now := time.Now()
	relativePath := filepath.Join(
		tenantID,
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		fileID+ext,
	)

	fullPath, err := safeLocalPath(s.basePath, relativePath)
	if err != nil {
		return nil, &storefile.Error{Op: "Put", Path: relativePath, Err: err}
	}
	if opts.Size > s.maxUploadSize {
		return nil, &storefile.Error{Op: "Put", Path: relativePath, Err: storefile.ErrInvalidSize}
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, &storefile.Error{Op: "Put", Path: fullPath, Err: err}
	}

	// Write to temp file, compute hash atomically
	tmpFile, err := os.CreateTemp(dir, "upload-*")
	if err != nil {
		return nil, &storefile.Error{Op: "Put", Path: fullPath, Err: err}
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	hash := sha256.New()
	reader := &contextReader{ctx: ctx, reader: opts.Reader}
	size, err := io.Copy(io.MultiWriter(tmpFile, hash), io.LimitReader(reader, s.maxUploadSize+1))
	if err != nil {
		tmpFile.Close()
		return nil, &storefile.Error{Op: "Put", Path: fullPath, Err: err}
	}
	tmpFile.Close()
	if size > s.maxUploadSize {
		return nil, &storefile.Error{Op: "Put", Path: relativePath, Err: storefile.ErrInvalidSize}
	}

	hashString := hex.EncodeToString(hash.Sum(nil))

	existing, err := getByTenantHash(ctx, s.metadata, tenantID, hashString)
	if err == nil && existing != nil {
		return existing, nil
	}

	if err := os.Rename(tmpPath, fullPath); err != nil {
		return nil, &storefile.Error{Op: "Put", Path: fullPath, Err: err}
	}

	file := &File{
		ID:          fileID,
		TenantID:    tenantID,
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

		if opts.GenerateThumb && s.imageProc.SupportsThumbnail(opts.ContentType) {
			thumbPath, err := s.generateThumbnail(fullPath, relativePath, opts.ThumbWidth, opts.ThumbHeight)
			if err == nil {
				file.ThumbnailPath = thumbPath
			}
		}
	}

	if s.metadata != nil {
		if err := s.metadata.Save(ctx, file); err != nil {
			os.Remove(fullPath)
			return nil, err
		}
	}

	return file, nil
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := contextError(r.ctx); err != nil {
		return 0, err
	}
	n, err := r.reader.Read(p)
	if err == nil {
		if ctxErr := contextError(r.ctx); ctxErr != nil {
			return n, ctxErr
		}
	}
	return n, err
}

// Get retrieves a file from local storage.
func (s *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath, err := safeLocalPath(s.basePath, path)
	if err != nil {
		return nil, &storefile.Error{Op: "Get", Path: path, Err: storefile.ErrInvalidPath}
	}

	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &storefile.Error{Op: "Get", Path: path, Err: storefile.ErrNotFound}
		}
		return nil, &storefile.Error{Op: "Get", Path: path, Err: err}
	}

	return f, nil
}

// Delete removes a file from local storage.
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath, err := safeLocalPath(s.basePath, path)
	if err != nil {
		return &storefile.Error{Op: "Delete", Path: path, Err: storefile.ErrInvalidPath}
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return &storefile.Error{Op: "Delete", Path: path, Err: storefile.ErrNotFound}
		}
		return &storefile.Error{Op: "Delete", Path: path, Err: err}
	}

	return nil
}

// Exists checks if a file exists in local storage.
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath, err := safeLocalPath(s.basePath, path)
	if err != nil {
		return false, &storefile.Error{Op: "Exists", Path: path, Err: storefile.ErrInvalidPath}
	}

	_, err = os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, &storefile.Error{Op: "Exists", Path: path, Err: err}
}

// Stat returns file information from local storage.
func (s *LocalStorage) Stat(ctx context.Context, path string) (*storefile.FileStat, error) {
	fullPath, err := safeLocalPath(s.basePath, path)
	if err != nil {
		return nil, &storefile.Error{Op: "Stat", Path: path, Err: storefile.ErrInvalidPath}
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &storefile.Error{Op: "Stat", Path: path, Err: storefile.ErrNotFound}
		}
		return nil, &storefile.Error{Op: "Stat", Path: path, Err: err}
	}

	return &storefile.FileStat{
		Path:         path,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
		IsDir:        info.IsDir(),
	}, nil
}

// List returns files in local storage matching the prefix.
func (s *LocalStorage) List(ctx context.Context, prefix string, limit int) ([]*storefile.FileStat, error) {
	if limit < 0 {
		return nil, storefile.ErrInvalidSize
	}
	baseAbs, err := filepath.Abs(s.basePath)
	if err != nil {
		return nil, err
	}
	rootPath := baseAbs
	if prefix != "" {
		rootPath, err = safeLocalPath(s.basePath, prefix)
		if err != nil {
			return nil, storefile.ErrInvalidPath
		}
	}

	var results []*storefile.FileStat

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(baseAbs, path)
			if err != nil {
				return err
			}
			results = append(results, &storefile.FileStat{
				Path:         relPath,
				Size:         info.Size(),
				ModifiedTime: info.ModTime(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

// GetURL returns a static URL for accessing the file.
func (s *LocalStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	if _, err := safeLocalPath(s.basePath, path); err != nil {
		return "", &storefile.Error{Op: "GetURL", Path: path, Err: storefile.ErrInvalidPath}
	}
	return strings.TrimRight(s.baseURL, "/") + "/" + escapeObjectKey(path), nil
}

// Copy copies a file within local storage.
func (s *LocalStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
	srcFullPath, err := safeLocalPath(s.basePath, srcPath)
	if err != nil {
		return &storefile.Error{Op: "Copy", Path: srcPath, Err: storefile.ErrInvalidPath}
	}
	dstFullPath, err := safeLocalPath(s.basePath, dstPath)
	if err != nil {
		return &storefile.Error{Op: "Copy", Path: dstPath, Err: storefile.ErrInvalidPath}
	}

	if err := os.MkdirAll(filepath.Dir(dstFullPath), 0755); err != nil {
		return &storefile.Error{Op: "Copy", Path: dstPath, Err: err}
	}

	src, err := os.Open(srcFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &storefile.Error{Op: "Copy", Path: srcPath, Err: storefile.ErrNotFound}
		}
		return &storefile.Error{Op: "Copy", Path: srcPath, Err: err}
	}
	defer src.Close()

	dst, err := os.Create(dstFullPath)
	if err != nil {
		return &storefile.Error{Op: "Copy", Path: dstPath, Err: err}
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return &storefile.Error{Op: "Copy", Path: dstPath, Err: err}
	}
	return nil
}

func (s *LocalStorage) generateThumbnail(srcPath, relativePath string, width, height int) (string, error) {
	if width <= 0 {
		width = 200
	}
	if height <= 0 {
		height = 200
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	thumbReader, err := s.imageProc.Thumbnail(src, width, height)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(relativePath)
	thumbRelPath := strings.TrimSuffix(relativePath, ext) + "_thumb" + ext
	thumbFullPath := filepath.Join(s.basePath, thumbRelPath)

	if err := os.MkdirAll(filepath.Dir(thumbFullPath), 0755); err != nil {
		return "", err
	}

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
