package storage

import (
	"context"
	"errors"
	"io"
)

// ErrNotFound is returned when a requested object does not exist.
var ErrNotFound = errors.New("object not found")

// ObjectStorage is the interface for cloud object storage backends.
type ObjectStorage interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	// Ping verifies connectivity to the storage backend.
	Ping(ctx context.Context) error
}
