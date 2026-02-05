# File Storage Module - Testing Guide

Complete testing documentation for the file storage module.

## Overview

The file storage module has comprehensive test coverage across multiple levels:

1. **Unit Tests** (standard library only) - in `store/file/`
2. **Integration Tests** (with PostgreSQL and MinIO) - in `examples/file-storage-tests/`
3. **Performance Benchmarks** - in both locations

## Unit Tests (Standard Library Only)

Location: `store/file/*_test.go`

### Test Files

| File | Lines | Description |
|------|-------|-------------|
| `utils_test.go` | 181 | ID generation, path safety, MIME conversions |
| `image_test.go` | 244 | Image processing, resizing, thumbnails |
| `local_test.go` | 531 | Local filesystem storage operations |
| `s3_signer_test.go` | 312 | AWS Signature V4 signing algorithm |
| `handler_test.go` | 428 | HTTP API endpoints |
| `example_test.go` | 164 | Usage examples for documentation |

**Total: 1,860 lines of test code**

### Running Unit Tests

```bash
# All tests
go test ./store/file/...

# Verbose output
go test -v ./store/file/...

# With race detector
go test -race ./store/file/...

# Specific test
go test ./store/file/ -run TestGenerateID

# Coverage report
go test ./store/file/... -cover
go test ./store/file/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Benchmark Results

```
BenchmarkGenerateID-16                     5.2µs/op    # UUID generation
BenchmarkIsPathSafe-16                      30ns/op    # Path validation
BenchmarkImageProcessor_GetInfo-16        10.3µs/op    # Image metadata
BenchmarkImageProcessor_Thumbnail-16      15.3ms/op    # Thumbnail generation
BenchmarkImageProcessor_Resize-16         41.7ms/op    # Image resizing
BenchmarkLocalStorage_Put-16               1.0ms/op    # File upload
BenchmarkLocalStorage_Get-16             168.1µs/op    # File download
BenchmarkS3Signer_SignRequest-16           9.5µs/op    # AWS signing
BenchmarkS3Signer_PresignRequest-16       17.7µs/op    # Presigned URL
BenchmarkHmacSHA256-16                   720.0ns/op    # HMAC operation
BenchmarkHandler_Upload-16                15.3µs/op    # HTTP upload
BenchmarkHandler_Download-16               7.9µs/op    # HTTP download
```

### Test Coverage by Module

#### 1. Utils (`utils_test.go`)

**TestGenerateID**
- Generates 1,000 IDs and verifies uniqueness
- Checks 32-character hex format
- Ensures cryptographic randomness

**TestIsPathSafe**
- ✅ Valid relative paths: `tenant/2026/02/05/file.txt`
- ❌ Path traversal: `../../../etc/passwd`
- ❌ Absolute paths: `/etc/passwd`
- ❌ Double slashes: `tenant//file.txt`
- ❌ Relative components: `./tenant/file.txt`

**TestMimeToExt / TestExtToMime**
- JPEG, PNG, GIF, WebP, SVG conversions
- PDF, text, JSON, XML, ZIP conversions
- Unknown type handling

#### 2. Image Processing (`image_test.go`)

**TestImageProcessor_IsImage**
- MIME type validation (case-insensitive)
- Image types: JPEG, PNG, GIF, WebP
- Non-image types rejected

**TestImageProcessor_GetInfo**
- Extract width, height, format
- JPEG and PNG support
- Invalid data error handling

**TestImageProcessor_Resize**
- Exact dimension scaling
- Format preservation (JPEG, PNG, GIF)
- Quality: 85% for JPEG

**TestImageProcessor_Thumbnail**
- Landscape: 1920x1080 → 200x112 (maintains aspect ratio)
- Portrait: 1080x1920 → 112x200 (maintains aspect ratio)
- Square: 500x500 → 100x100
- Small images: Not upscaled (50x50 stays 50x50)

**TestImageProcessor_InvalidData**
- Graceful error handling for non-image data

#### 3. Local Storage (`local_test.go`)

**TestLocalStorage_Put_Get**
- Upload file with SHA256 hash
- Download and verify content
- Metadata generation (ID, path, size, hash)

**TestLocalStorage_Delete**
- File deletion from filesystem
- Error handling for non-existent files

**TestLocalStorage_Exists**
- Check file existence
- Returns false for non-existent files

**TestLocalStorage_Stat**
- File metadata (size, modified time)
- Directory detection

**TestLocalStorage_List**
- Directory listing with tenant prefix
- Limit enforcement
- Sorted by modification time

**TestLocalStorage_Copy**
- File copying within storage
- Content verification

**TestLocalStorage_PathSafety**
- Rejects `../../../etc/passwd`
- Rejects `/etc/passwd`
- Protects against path traversal attacks

#### 4. S3 Signing (`s3_signer_test.go`)

**TestS3Signer_SignRequest**
- Adds required headers (x-amz-date, x-amz-content-sha256)
- Generates Authorization header
- Format: `AWS4-HMAC-SHA256 Credential=..., SignedHeaders=..., Signature=...`

**TestS3Signer_PresignRequest**
- Generates presigned URLs
- Query parameters: X-Amz-Algorithm, Credential, Date, Expires, SignedHeaders, Signature
- Expiry validation (15min-7days, default 15min)

**TestS3Signer_BuildCanonicalQueryString**
- Sorts parameters alphabetically
- URL-encodes values
- Handles multiple values for same key

**TestS3Signer_BuildCanonicalHeaders**
- Lowercases header names
- Sorts headers alphabetically
- Includes host + x-amz-* headers

**TestEmptyStringSHA256**
- Verifies SHA256 of empty string
- 64-character hex output

**TestHmacSHA256**
- HMAC-SHA256 determinism
- Key sensitivity
- 32-byte output

#### 5. HTTP Handler (`handler_test.go`)

Uses mock implementations for Storage and MetadataManager.

**TestHandler_Upload**
- Multipart form file upload
- Tenant ID from context
- User ID from context
- Returns file metadata

**TestHandler_Download**
- Streaming file download
- Sets Content-Type and Content-Disposition headers
- Updates access time asynchronously

**TestHandler_GetInfo**
- Returns file metadata (name, size, MIME type)
- Error handling for non-existent files

**TestHandler_Delete**
- Soft delete via metadata manager
- 404 for non-existent files

**TestHandler_List**
- Paginated file listing
- Supports filters (tenant, MIME type, uploader, date range)
- Returns total count and page info

**TestHandler_GetURL**
- Generates presigned URLs
- Configurable expiry duration

## Integration Tests (With External Dependencies)

Location: `examples/file-storage-tests/`

### Prerequisites

1. **PostgreSQL** (via Docker):
   ```bash
   docker run -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:16-alpine
   psql -f ../../store/file/migrations/001_create_files_table.up.sql
   ```

2. **MinIO** (via Docker):
   ```bash
   docker run -p 9000:9000 \
     -e MINIO_ROOT_USER=minioadmin \
     -e MINIO_ROOT_PASSWORD=minioadmin \
     minio/minio server /data

   mc alias set myminio http://localhost:9000 minioadmin minioadmin
   mc mb myminio/test-bucket
   ```

3. **Environment Variables**:
   ```bash
   export POSTGRES_URL="postgres://postgres:postgres@localhost:5432/plumego_test?sslmode=disable"
   export S3_ENDPOINT="localhost:9000"
   export S3_BUCKET="test-bucket"
   export S3_REGION="us-east-1"
   export S3_ACCESS_KEY="minioadmin"
   export S3_SECRET_KEY="minioadmin"
   ```

### Running Integration Tests

```bash
cd examples/file-storage-tests

# All tests
go test -v

# Specific test
go test -v -run TestIntegration_DatabaseMetadata
go test -v -run TestIntegration_S3Storage

# Benchmarks
go test -bench=. -benchtime=10s

# With race detector
go test -v -race
```

### Integration Test Coverage

**TestIntegration_DatabaseMetadata**
- Real PostgreSQL queries
- CRUD operations with actual SQL
- Deduplication via hash lookup
- Pagination and filtering
- Soft delete verification
- JSONB metadata queries

**TestIntegration_LocalStorageWithDatabase**
- LocalStorage + DBMetadataManager
- File upload with hash calculation
- Automatic deduplication (same content = same file)
- Metadata persistence
- Streaming download

**TestIntegration_S3Storage**
- MinIO/S3 compatibility
- Upload with AWS Signature V4
- File existence checks
- Object statistics
- Streaming download
- Presigned URL generation
- Object deletion

### Integration Benchmarks

```
BenchmarkIntegration_LocalStoragePut     1.2ms/op  (with PostgreSQL)
BenchmarkIntegration_S3StoragePut       10.5ms/op  (with MinIO)
```

## Database Migrations

Location: `store/file/migrations/`

### Files

1. **001_create_files_table.up.sql** (75 lines)
   - CREATE TABLE with comprehensive schema
   - 11 strategic indexes
   - JSONB metadata support
   - Unique constraint on (tenant_id, hash)

2. **001_create_files_table.down.sql** (14 lines)
   - Rollback migration
   - DROP TABLE and indexes

### Schema Overview

```sql
CREATE TABLE files (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    hash VARCHAR(64) NOT NULL,         -- SHA256 for deduplication
    deleted_at TIMESTAMP,               -- Soft delete
    last_access_at TIMESTAMP,           -- LRU cleanup
    -- ... other fields
);

-- Indexes for performance
CREATE INDEX idx_files_tenant_id ON files(tenant_id);
CREATE INDEX idx_files_hash ON files(hash);
CREATE UNIQUE INDEX idx_files_tenant_hash ON files(tenant_id, hash);
-- ... 8 more indexes
```

### Applying Migrations

```bash
# Up migration
psql -d plumego -f store/file/migrations/001_create_files_table.up.sql

# Down migration
psql -d plumego -f store/file/migrations/001_create_files_table.down.sql
```

## Test Quality Metrics

### Coverage

- **Unit Tests**: All core functionality covered
- **Integration Tests**: Real infrastructure validation
- **Race Conditions**: None detected (tested with -race)
- **Edge Cases**: Path traversal, invalid data, missing files
- **Performance**: All operations benchmarked

### Test Characteristics

- ✅ **Fast**: Unit tests complete in < 250ms
- ✅ **Isolated**: Each test uses temporary directories/mocks
- ✅ **Deterministic**: No flaky tests
- ✅ **Portable**: Works on Linux, macOS, Windows
- ✅ **Zero Dependencies**: Unit tests use standard library only
- ✅ **Well Documented**: Examples for all major features

### Continuous Integration

Tests are suitable for CI/CD:
- GitHub Actions workflow included in `examples/file-storage-tests/README.md`
- Docker service containers for PostgreSQL and MinIO
- Health checks for services
- Automated cleanup after tests

## Development Workflow

### Adding New Tests

1. **Unit Test** (for core logic):
   ```bash
   # Create test file
   touch store/file/myfeature_test.go

   # Write tests
   # Run tests
   go test ./store/file/ -run TestMyFeature
   ```

2. **Integration Test** (for external systems):
   ```bash
   # Add to examples/file-storage-tests/integration_test.go

   # Run with services
   go test -v -run TestIntegration_MyFeature
   ```

### Before Commit

```bash
# Run all tests
go test ./store/file/...

# Check race conditions
go test -race ./store/file/...

# Format code
gofmt -w store/file/

# Vet code
go vet ./store/file/...
```

## Troubleshooting

### Unit Tests Failing

1. Check Go version (1.24+ required)
2. Ensure no external services needed
3. Check filesystem permissions for temp directories

### Integration Tests Failing

1. Verify PostgreSQL is running: `docker ps`
2. Verify MinIO is running: `curl http://localhost:9000`
3. Check environment variables are set
4. Apply database migrations
5. Check for leftover test data

### Performance Issues

1. Use SSD for local storage tests
2. Use connection pooling for PostgreSQL
3. Enable HTTP/2 for S3 connections
4. Check network latency for MinIO tests

## Summary

| Category | Coverage | Files | Lines |
|----------|----------|-------|-------|
| Unit Tests | All core functionality | 6 files | 1,860 |
| Integration Tests | PostgreSQL + MinIO | 1 file | 400+ |
| Migrations | PostgreSQL schema | 2 files | 89 |
| Benchmarks | 12 operations | - | - |
| Documentation | Complete | 3 READMEs | 750+ |

**Total: 3,100+ lines of tests and documentation**

## References

- Unit Tests: `store/file/*_test.go`
- Integration Tests: `examples/file-storage-tests/`
- Migrations: `store/file/migrations/`
- Module Documentation: `store/file/README.md`
- Integration Setup: `examples/file-storage-tests/README.md`
