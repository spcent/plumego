# File Storage Module

A production-ready file storage module for the Plumego framework, providing unified file management with support for local filesystem and S3-compatible cloud storage.

## Features

### Core Capabilities
- **Multi-backend storage**: Local filesystem and S3-compatible cloud storage
- **Automatic deduplication**: SHA256 hash-based duplicate detection
- **Multi-tenancy**: Built-in tenant isolation
- **Image processing**: Automatic thumbnail generation with aspect ratio preservation
- **Soft delete**: Safe file deletion with timestamp tracking
- **Streaming I/O**: Efficient handling of large files
- **Metadata management**: PostgreSQL-backed searchable metadata
- **Access tracking**: Last access time for cleanup strategies
- **Presigned URLs**: Temporary access URLs for S3 storage

### Security
- **Path traversal protection**: Validated file paths
- **AWS Signature V4**: Complete authentication implementation
- **Tenant isolation**: Context-based access control
- **Content type validation**: MIME type checking

### Standards Compliance
- **Zero dependencies**: Standard library only
- **S3 compatibility**: Works with AWS S3, MinIO, DigitalOcean Spaces, etc.
- **RESTful API**: Standard HTTP endpoints
- **PostgreSQL storage**: ACID-compliant metadata

## Architecture

```
store/file/
├── file.go          # Core interfaces (Storage, MetadataManager, ImageProcessor)
├── types.go         # Data structures (File, PutOptions, Query, etc.)
├── errors.go        # Error definitions
├── utils.go         # Utility functions (ID generation, path safety, MIME types)
├── local.go         # Local filesystem storage implementation
├── s3.go            # S3-compatible cloud storage implementation
├── s3_signer.go     # AWS Signature V4 signing algorithm
├── image.go         # Image processing (resize, thumbnail, metadata)
├── metadata.go      # Database metadata manager
├── handler.go       # HTTP REST API handler
└── example_test.go  # Usage examples
```

## Quick Start

### 1. Local Storage

```go
import (
    "context"
    "database/sql"
    "github.com/spcent/plumego/store/file"
)

// Initialize metadata manager
db, _ := sql.Open("postgres", "postgres://localhost/plumego")
metadata := file.NewDBMetadataManager(db)

// Create local storage
storage, err := file.NewLocalStorage(
    "/var/uploads",              // Base directory
    "http://example.com/files",  // Base URL
    metadata,
)

// Upload a file
result, err := storage.Put(context.Background(), file.PutOptions{
    TenantID:      "tenant-123",
    Reader:        fileReader,
    FileName:      "photo.jpg",
    ContentType:   "image/jpeg",
    GenerateThumb: true,
    ThumbWidth:    200,
    ThumbHeight:   200,
    UploadedBy:    "user-456",
})

// Download a file
reader, err := storage.Get(context.Background(), result.Path)
defer reader.Close()
```

### 2. S3 Storage

```go
// Create S3 storage
storage, err := file.NewS3Storage(
    "s3.amazonaws.com",  // Endpoint
    "my-bucket",         // Bucket
    "us-east-1",         // Region
    "access-key",        // Access key
    "secret-key",        // Secret key
    metadata,
)

// Upload to S3
result, err := storage.Put(context.Background(), file.PutOptions{
    TenantID:    "tenant-123",
    Reader:      fileReader,
    FileName:    "document.pdf",
    ContentType: "application/pdf",
    UploadedBy:  "user-456",
})

// Get presigned URL (valid for 1 hour)
url, err := storage.GetURL(context.Background(), result.Path, time.Hour)
```

### 3. HTTP Handler

```go
// Initialize handler
handler := file.NewHandler(storage, metadata).
    WithMaxSize(100 << 20) // 100 MiB

// Register routes (standard library)
mux := http.NewServeMux()
handler.RegisterRoutes(mux)

// Start server
http.ListenAndServe(":8080", mux)
```

## API Endpoints

### Upload File
```http
POST /files
Content-Type: multipart/form-data

file: <file data>
generate_thumb: true
thumb_width: 200
thumb_height: 200
```

**Response:**
```json
{
  "id": "abc123",
  "tenant_id": "tenant-123",
  "name": "photo.jpg",
  "path": "tenant-123/2026/02/05/abc123.jpg",
  "size": 1048576,
  "mime_type": "image/jpeg",
  "hash": "sha256...",
  "width": 1920,
  "height": 1080,
  "thumbnail_path": "tenant-123/2026/02/05/abc123_thumb.jpg",
  "created_at": "2026-02-05T10:30:00Z"
}
```

### Download File
```http
GET /files/{id}
```

**Response:**
- Streams file content with proper `Content-Type` and `Content-Disposition` headers
- Updates last access time asynchronously

### Get File Info
```http
GET /files/{id}/info
```

**Response:**
```json
{
  "id": "abc123",
  "name": "photo.jpg",
  "size": 1048576,
  "mime_type": "image/jpeg",
  "created_at": "2026-02-05T10:30:00Z"
}
```

### Get Presigned URL
```http
GET /files/{id}/url?expiry=3600
```

**Response:**
```json
{
  "url": "https://s3.amazonaws.com/bucket/key?X-Amz-Signature=...",
  "expires_in": "3600"
}
```

### Delete File
```http
DELETE /files/{id}
```

**Response:**
```json
{
  "message": "file deleted"
}
```

### List Files
```http
GET /files?page=1&page_size=20&tenant_id=tenant-123&mime_type=image/jpeg
```

**Query Parameters:**
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)
- `tenant_id`: Filter by tenant
- `uploaded_by`: Filter by uploader
- `mime_type`: Filter by MIME type
- `start_time`: Filter by creation time (RFC3339)
- `end_time`: Filter by creation time (RFC3339)
- `order_by`: Sort order (e.g., "created_at DESC")

**Response:**
```json
{
  "items": [...],
  "total": 150,
  "page": 1,
  "page_size": 20,
  "total_page": 8
}
```

## Database Schema

```sql
CREATE TABLE files (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    size BIGINT NOT NULL,
    mime_type VARCHAR(128),
    extension VARCHAR(32),
    hash VARCHAR(64) NOT NULL,
    width INTEGER,
    height INTEGER,
    thumbnail_path VARCHAR(1024),
    storage_type VARCHAR(32) NOT NULL,
    metadata JSONB,
    uploaded_by VARCHAR(64),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    last_access_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_files_tenant_id ON files(tenant_id);
CREATE INDEX idx_files_hash ON files(hash);
CREATE UNIQUE INDEX idx_files_tenant_hash ON files(tenant_id, hash) WHERE deleted_at IS NULL;
CREATE INDEX idx_files_created_at ON files(created_at);
CREATE INDEX idx_files_deleted_at ON files(deleted_at);
CREATE INDEX idx_files_mime_type ON files(mime_type);
```

## File Organization

### Local Storage
Files are organized in a date-based directory structure:

```
/var/uploads/
└── tenant-123/
    └── 2026/
        └── 02/
            └── 05/
                ├── abc123.jpg
                ├── abc123_thumb.jpg
                ├── def456.pdf
                └── ghi789.txt
```

### S3 Storage
Objects are stored with the same path structure:

```
s3://my-bucket/
└── tenant-123/
    └── 2026/
        └── 02/
            └── 05/
                └── abc123.jpg
```

## Deduplication

Files with identical content (SHA256 hash) are automatically deduplicated:

1. During upload, file content is hashed
2. Database is checked for existing file with same hash
3. If found, existing metadata is returned
4. If not found, file is stored and new metadata created

This saves storage space and reduces redundant uploads.

## Image Processing

The module includes built-in image processing using Go's standard library:

### Supported Formats
- JPEG
- PNG
- GIF
- WebP (detection only, no processing)

### Features
- **Resize**: Scale to exact dimensions
- **Thumbnail**: Scale with aspect ratio preservation
- **Metadata**: Extract width, height, format

### Algorithm
- Nearest-neighbor scaling for thumbnails
- JPEG quality: 85%
- Default thumbnail size: 200x200

## AWS Signature V4

The S3 signer implements complete AWS Signature Version 4 authentication:

### Signing Process
1. **Canonical Request**: Normalize HTTP request
   - Method, URI, query string
   - Sorted headers (host + x-amz-*)
   - Payload hash

2. **String to Sign**: Create signing payload
   - Algorithm identifier
   - Timestamp
   - Credential scope
   - Hashed canonical request

3. **Signature**: HMAC-SHA256 chain
   - kDate = HMAC("AWS4" + secret, date)
   - kRegion = HMAC(kDate, region)
   - kService = HMAC(kRegion, "s3")
   - kSigning = HMAC(kService, "aws4_request")
   - signature = HMAC(kSigning, stringToSign)

4. **Authorization**: Add to request
   - Header: `Authorization: AWS4-HMAC-SHA256 Credential=..., SignedHeaders=..., Signature=...`
   - Or URL parameter: `X-Amz-Signature=...`

### Presigned URLs
Generate temporary access URLs without exposing credentials:

```go
url, err := storage.GetURL(ctx, path, time.Hour)
```

Valid for 15 minutes to 7 days.

## Security Considerations

### Path Traversal Protection
```go
// isPathSafe prevents "../../../etc/passwd" attacks
if !isPathSafe(path) {
    return ErrInvalidPath
}
```

### Tenant Isolation
```go
// Extract tenant ID from context
tenantID := ctx.Value("tenant_id").(string)

// All queries are tenant-scoped
SELECT * FROM files WHERE tenant_id = $1 AND deleted_at IS NULL
```

### Signature Verification
- AWS Signature V4 uses HMAC-SHA256
- Timestamp validation prevents replay attacks
- Presigned URLs have expiry enforcement

## Performance Optimizations

### Streaming I/O
Files are streamed directly to/from storage:

```go
// No in-memory buffering
io.Copy(destination, source)
```

### Async Access Updates
Last access time is updated asynchronously:

```go
go metadata.UpdateAccessTime(context.Background(), fileID)
```

### Database Indexes
Strategic indexes for common queries:
- `tenant_id` for tenant isolation
- `hash` for deduplication
- `created_at` for temporal queries
- `deleted_at` for soft delete filtering

## Error Handling

```go
var (
    ErrNotFound    = errors.New("file: not found")
    ErrInvalidPath = errors.New("file: invalid path")
)

type Error struct {
    Op   string  // Operation (Put, Get, Delete)
    Path string  // File path
    Err  error   // Underlying error
}
```

## Testing

Run tests:
```bash
go test ./store/file/...
```

Run with race detector:
```bash
go test -race ./store/file/...
```

## Integration with Plumego

### Component Registration
```go
type FileStorageComponent struct {
    storage  file.Storage
    metadata file.MetadataManager
    handler  *file.Handler
}

func (c *FileStorageComponent) RegisterRoutes(r *router.Router) {
    r.HandleFunc("POST /api/files", c.handler.Upload)
    r.HandleFunc("GET /api/files/{id}", c.handler.Download)
    // ... other routes
}

// Register with app
app := core.New(
    core.WithComponent(&FileStorageComponent{...}),
)
```

### Middleware Integration
```go
// Add tenant extraction middleware
app.Use(middleware.TenantID("X-Tenant-ID"))

// Add authentication
app.Use(middleware.JWT(secret))

// File routes automatically receive tenant_id and user_id from context
```

## Production Checklist

- [ ] Configure database connection pool
- [ ] Set up S3 bucket with proper IAM policies
- [ ] Configure CDN for file delivery
- [ ] Set up backup strategy for local storage
- [ ] Implement cleanup job for old files (using `last_access_at`)
- [ ] Configure file size limits per tenant
- [ ] Set up monitoring for storage usage
- [ ] Implement rate limiting for uploads
- [ ] Configure CORS for direct browser uploads
- [ ] Set up virus scanning for uploaded files

## License

Part of the Plumego framework.
