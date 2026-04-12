# store/file — Pure File Storage Interfaces

Stable, transport-agnostic interfaces and primitives for file storage operations.
No tenant semantics, no HTTP transport, no concrete backend or signed-URL or image-pipeline implementations.

## Package Layout

```
store/file/
├── file.go      # Storage interface
├── types.go     # File, PutOptions, FileStat, Query
├── errors.go    # Error definitions (ErrNotFound, ErrInvalidPath, …)
```

## Interfaces

### Storage

Path-based object storage abstraction:

```go
type Storage interface {
    Put(ctx context.Context, opts PutOptions) (*File, error)
    Get(ctx context.Context, path string) (io.ReadCloser, error)
    Delete(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
    Stat(ctx context.Context, path string) (*FileStat, error)
    List(ctx context.Context, prefix string, limit int) ([]*FileStat, error)
    Copy(ctx context.Context, srcPath, dstPath string) error
}
```

## Concrete Implementations

- Tenant-aware storage backends (local filesystem, S3) and the database-backed
  metadata manager live in **`x/data/file`**.
- HTTP upload/download handlers and request parsing live in **`x/fileapi`**.
- `store/file` stays responsible for the stable `Storage` contract, shared file
  types, and file operation errors.
- Backend-specific configuration, metadata persistence, signed URLs, path/id
  helpers, and thumbnail or image-processing helpers live in **`x/data/file`**
  and **`x/fileapi`**.

## Non-Goals

- Tenant-aware adapters — use `x/data/file`
- HTTP transport — use `x/fileapi`
- Business repositories
- Read-write splitting or sharding topology

## Testing

- Stable package tests live in `store/file/coverage_test.go`.
- Backend-specific, metadata-manager, image-processing, and HTTP transport tests belong in `x/data/file` or `x/fileapi`, not in the stable root.
