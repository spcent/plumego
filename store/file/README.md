# store/file — Pure File Storage Interfaces

Stable, transport-agnostic interfaces and primitives for file storage operations.
No tenant semantics, no HTTP transport, no concrete backend or image-pipeline implementations.

## Package Layout

```
store/file/
├── file.go      # Storage and MetadataManager interfaces
├── types.go     # File, PutOptions, FileStat, Query
├── errors.go    # Error definitions (ErrNotFound, ErrInvalidPath, …)
└── utils.go     # GenerateID, IsPathSafe, MimeToExt, ExtToMime
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
    GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)
    Copy(ctx context.Context, srcPath, dstPath string) error
}
```

### MetadataManager

Persistent metadata store abstraction:

```go
type MetadataManager interface {
    Save(ctx context.Context, file *File) error
    Get(ctx context.Context, id string) (*File, error)
    GetByPath(ctx context.Context, path string) (*File, error)
    GetByHash(ctx context.Context, hash string) (*File, error)
    List(ctx context.Context, query Query) ([]*File, int64, error)
    Delete(ctx context.Context, id string) error
    UpdateAccessTime(ctx context.Context, id string) error
}
```

## Concrete Implementations

- Tenant-aware storage backends (local filesystem, S3) and the database-backed
  metadata manager live in **`x/data/file`**.
- HTTP upload/download handlers and request parsing live in **`x/fileapi`**.
- `store/file` stays responsible for the stable `Storage` / `MetadataManager`
  contracts, shared file types, errors, and pure helpers such as path safety
  checks.
- Backend-specific configuration and thumbnail or image-processing helpers live
  in **`x/data/file`** and **`x/fileapi`**.

## Non-Goals

- Tenant-aware adapters — use `x/data/file`
- HTTP transport — use `x/fileapi`
- Business repositories
- Read-write splitting or sharding topology
