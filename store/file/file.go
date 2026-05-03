// Package file provides stable, transport-agnostic contracts, shared types, and
// errors for file storage operations. Concrete storage backends, metadata
// implementations, and helper policy live outside the stable root; HTTP
// handlers live in x/fileapi.
package file

import (
	"context"
	"io"
)

// Storage defines the interface for file storage operations.
// Implementations must be safe for concurrent use.
type Storage interface {
	// Put stores content from opts.Reader and returns file metadata on success.
	// Implementations define path generation and validation policy.
	Put(ctx context.Context, opts PutOptions) (*File, error)

	// Get retrieves a file by its path.
	// Returns a ReadCloser that must be closed by the caller.
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file by its path. Missing-path behavior is
	// implementation-defined and should be documented by concrete backends.
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists at the given path.
	Exists(ctx context.Context, path string) (bool, error)

	// Stat returns metadata about a file.
	Stat(ctx context.Context, path string) (*FileStat, error)

	// List returns files matching the prefix. If limit is 0, all files are
	// returned; negative limits are invalid and should return ErrInvalidSize.
	// Ordering and pagination consistency are implementation-defined.
	List(ctx context.Context, prefix string, limit int) ([]*FileStat, error)

	// Copy copies a file from srcPath to dstPath. Overwrite behavior, metadata
	// preservation, and atomicity are implementation-defined.
	Copy(ctx context.Context, srcPath, dstPath string) error
}
