# File Storage Integration Tests

Integration and performance tests for the file storage module using real PostgreSQL and MinIO/S3.

## Overview

These tests verify the file storage module against real infrastructure:
- **PostgreSQL**: Database metadata management with actual SQL queries
- **MinIO/S3**: S3-compatible object storage with real HTTP requests
- **Performance**: Benchmarks with real I/O operations

## Prerequisites

### 1. PostgreSQL Database

Start PostgreSQL using Docker:

```bash
docker run --name postgres-test \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=plumego_test \
  -p 5432:5432 \
  -d postgres:16-alpine
```

Apply migrations:

```bash
psql -h localhost -U postgres -d plumego_test \
  -f ../../store/file/migrations/001_create_files_table.up.sql
```

### 2. MinIO (S3-Compatible Storage)

Start MinIO using Docker:

```bash
docker run --name minio-test \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"
```

Create test bucket using MinIO CLI:

```bash
# Install mc (MinIO Client)
brew install minio/stable/mc  # macOS
# or: wget https://dl.min.io/client/mc/release/linux-amd64/mc

# Configure alias
mc alias set myminio http://localhost:9000 minioadmin minioadmin

# Create bucket
mc mb myminio/test-bucket
```

Or create bucket via web console: http://localhost:9001

### 3. Environment Variables

Set environment variables for tests:

```bash
export POSTGRES_URL="postgres://postgres:postgres@localhost:5432/plumego_test?sslmode=disable"
export S3_ENDPOINT="localhost:9000"
export S3_BUCKET="test-bucket"
export S3_REGION="us-east-1"
export S3_ACCESS_KEY="minioadmin"
export S3_SECRET_KEY="minioadmin"
```

Or create `.env` file:

```bash
cat > .env <<EOF
POSTGRES_URL=postgres://postgres:postgres@localhost:5432/plumego_test?sslmode=disable
S3_ENDPOINT=localhost:9000
S3_BUCKET=test-bucket
S3_REGION=us-east-1
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
EOF

# Load environment
source .env
```

## Running Tests

### All Integration Tests

```bash
go test -v
```

### Specific Tests

```bash
# Database integration only
go test -v -run TestIntegration_DatabaseMetadata

# S3/MinIO integration only
go test -v -run TestIntegration_S3Storage

# Local storage with database
go test -v -run TestIntegration_LocalStorageWithDatabase
```

### Benchmarks

```bash
# All benchmarks
go test -bench=. -benchtime=10s

# Local storage benchmark
go test -bench=BenchmarkIntegration_LocalStoragePut -benchtime=10s

# S3 storage benchmark
go test -bench=BenchmarkIntegration_S3StoragePut -benchtime=10s
```

### With Race Detector

```bash
go test -v -race
```

## Test Coverage

### Database Tests (`TestIntegration_DatabaseMetadata`)

Tests DBMetadataManager with real PostgreSQL:
- ✅ Save file metadata
- ✅ Get file by ID
- ✅ Get file by hash (deduplication lookup)
- ✅ List files with pagination and filters
- ✅ Update access time
- ✅ Soft delete
- ✅ JSONB metadata queries

### Local Storage Tests (`TestIntegration_LocalStorageWithDatabase`)

Tests LocalStorage with DBMetadataManager:
- ✅ Upload files with automatic hash calculation
- ✅ File deduplication via SHA256 hash
- ✅ Download files with streaming I/O
- ✅ Metadata persistence in PostgreSQL
- ✅ Tenant isolation

### S3 Tests (`TestIntegration_S3Storage`)

Tests S3Storage with MinIO:
- ✅ Upload to S3 with AWS Signature V4
- ✅ Check file existence
- ✅ Get file statistics (size, modified time)
- ✅ Download from S3
- ✅ Generate presigned URLs
- ✅ Delete objects

## Performance Benchmarks

Expected performance (reference: i5-8250U, SSD, local PostgreSQL/MinIO):

```
BenchmarkIntegration_LocalStoragePut-8     1000    1.2 ms/op  (with DB)
BenchmarkIntegration_S3StoragePut-8         100   10.5 ms/op  (with MinIO)
```

Performance tips:
- **Database**: Use connection pooling (`db.SetMaxOpenConns(10)`)
- **S3**: Enable HTTP/2 and connection reuse
- **Local**: Use SSD for better IOPS

## Troubleshooting

### PostgreSQL Connection Failed

```bash
# Check if PostgreSQL is running
docker ps | grep postgres-test

# Check logs
docker logs postgres-test

# Test connection
psql -h localhost -U postgres -d plumego_test -c "SELECT 1"
```

### MinIO Connection Failed

```bash
# Check if MinIO is running
docker ps | grep minio-test

# Check logs
docker logs minio-test

# Test with curl
curl http://localhost:9000
```

### Migration Not Applied

```bash
# Check if table exists
psql -h localhost -U postgres -d plumego_test \
  -c "\\d files"

# Re-apply migration
psql -h localhost -U postgres -d plumego_test \
  -f ../../store/file/migrations/001_create_files_table.up.sql
```

### Test Failures Due to Leftover Data

```bash
# Clean up test data
psql -h localhost -U postgres -d plumego_test \
  -c "DELETE FROM files WHERE tenant_id LIKE 'test-%'"

# Clean MinIO bucket
mc rm --recursive --force myminio/test-bucket/test-tenant/
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: plumego_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      minio:
        image: minio/minio
        env:
          MINIO_ROOT_USER: minioadmin
          MINIO_ROOT_PASSWORD: minioadmin
        ports:
          - 9000:9000
        command: server /data

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Apply migrations
        run: |
          psql -h localhost -U postgres -d plumego_test \
            -f ../../store/file/migrations/001_create_files_table.up.sql
        env:
          PGPASSWORD: postgres

      - name: Create MinIO bucket
        run: |
          mc alias set myminio http://localhost:9000 minioadmin minioadmin
          mc mb myminio/test-bucket

      - name: Run integration tests
        run: go test -v
        env:
          POSTGRES_URL: "postgres://postgres:postgres@localhost:5432/plumego_test?sslmode=disable"
          S3_ENDPOINT: "localhost:9000"
          S3_BUCKET: "test-bucket"
          S3_REGION: "us-east-1"
          S3_ACCESS_KEY: "minioadmin"
          S3_SECRET_KEY: "minioadmin"
```

## Development Workflow

1. **Start services**:
   ```bash
   docker-compose up -d  # If using docker-compose
   ```

2. **Apply migrations**:
   ```bash
   make migrate-up
   ```

3. **Run tests during development**:
   ```bash
   # Watch mode (requires: go install github.com/cespare/reflex@latest)
   reflex -r '\.go$' -- go test -v
   ```

4. **Clean up**:
   ```bash
   docker-compose down -v
   ```

## Additional Resources

- [PostgreSQL Docker Image](https://hub.docker.com/_/postgres)
- [MinIO Documentation](https://min.io/docs/minio/linux/index.html)
- [AWS S3 API Compatibility](https://docs.min.io/docs/aws-cli-with-minio.html)
- [Go Database SQL Tutorial](https://go.dev/doc/database/sql-intro)

## License

Part of the Plumego framework.
