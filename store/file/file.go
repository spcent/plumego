// Package file provides a unified interface for file storage operations.
// It supports both local filesystem storage and S3-compatible cloud storage,
// with built-in image processing, file deduplication, and metadata management.
package file

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for file storage operations.
// Implementations must be safe for concurrent use.
type Storage interface {
	// Put uploads a file with the given options.
	// Returns file metadata on success.
	Put(ctx context.Context, opts PutOptions) (*File, error)

	// Get retrieves a file by its path.
	// Returns a ReadCloser that must be closed by the caller.
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file by its path.
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists at the given path.
	Exists(ctx context.Context, path string) (bool, error)

	// Stat returns metadata about a file.
	Stat(ctx context.Context, path string) (*FileStat, error)

	// List returns a list of files matching the prefix.
	// If limit is 0, all files are returned.
	List(ctx context.Context, prefix string, limit int) ([]*FileStat, error)

	// GetURL returns a URL for accessing the file.
	// For local storage, returns a static URL.
	// For S3 storage, returns a presigned URL valid for the given duration.
	GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// Copy copies a file from srcPath to dstPath.
	Copy(ctx context.Context, srcPath, dstPath string) error
}

// MetadataManager manages file metadata in a persistent store.
type MetadataManager interface {
	// Save stores file metadata.
	Save(ctx context.Context, file *File) error

	// Get retrieves file metadata by ID.
	Get(ctx context.Context, id string) (*File, error)

	// GetByPath retrieves file metadata by path.
	GetByPath(ctx context.Context, path string) (*File, error)

	// GetByHash retrieves file metadata by hash for deduplication.
	// Returns nil if no file with the hash exists.
	GetByHash(ctx context.Context, hash string) (*File, error)

	// List retrieves file metadata matching the query.
	List(ctx context.Context, query Query) ([]*File, int64, error)

	// Delete soft-deletes file metadata.
	Delete(ctx context.Context, id string) error

	// UpdateAccessTime updates the last access timestamp.
	UpdateAccessTime(ctx context.Context, id string) error
}

// ImageProcessor handles image processing operations.
type ImageProcessor interface {
	// Resize scales an image to the specified dimensions.
	Resize(src io.Reader, width, height int) (io.Reader, error)

	// Thumbnail generates a thumbnail maintaining aspect ratio.
	// The thumbnail will fit within maxWidth x maxHeight.
	Thumbnail(src io.Reader, maxWidth, maxHeight int) (io.Reader, error)

	// GetInfo extracts image metadata (width, height, format).
	GetInfo(src io.Reader) (*ImageInfo, error)

	// IsImage checks if the MIME type represents an image.
	IsImage(mimeType string) bool
}
