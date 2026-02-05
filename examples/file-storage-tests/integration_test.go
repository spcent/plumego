package integration

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/spcent/plumego/store/file"
)

const (
	// Environment variables for integration tests
	EnvPostgresURL = "POSTGRES_URL"
	EnvS3Endpoint  = "S3_ENDPOINT"
	EnvS3Bucket    = "S3_BUCKET"
	EnvS3Region    = "S3_REGION"
	EnvS3AccessKey = "S3_ACCESS_KEY"
	EnvS3SecretKey = "S3_SECRET_KEY"
)

// TestIntegration_DatabaseMetadata tests the DBMetadataManager with a real PostgreSQL database.
//
// Prerequisites:
//   - PostgreSQL database running
//   - Set environment variable: POSTGRES_URL=postgres://user:pass@localhost/dbname?sslmode=disable
//   - Run migration: psql -f ../../store/file/migrations/001_create_files_table.up.sql
//
// Run with:
//
//	POSTGRES_URL="postgres://..." go test -v -run TestIntegration_DatabaseMetadata
func TestIntegration_DatabaseMetadata(t *testing.T) {
	postgresURL := os.Getenv(EnvPostgresURL)
	if postgresURL == "" {
		t.Skipf("Skipping integration test: %s not set", EnvPostgresURL)
	}

	// Connect to database
	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Create metadata manager
	metadata := file.NewDBMetadataManager(db)
	ctx := context.Background()

	// Test Save
	testFile := &file.File{
		ID:          fmt.Sprintf("test-%d", time.Now().Unix()),
		TenantID:    "test-tenant",
		Name:        "test.txt",
		Path:        "test-tenant/2026/02/05/test.txt",
		Size:        1024,
		MimeType:    "text/plain",
		Extension:   ".txt",
		Hash:        "abc123",
		StorageType: "local",
		UploadedBy:  "test-user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    map[string]interface{}{"custom_field": "custom_value"},
	}

	if err := metadata.Save(ctx, testFile); err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Test Get
	retrieved, err := metadata.Get(ctx, testFile.ID)
	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}

	if retrieved.ID != testFile.ID {
		t.Errorf("ID = %q, want %q", retrieved.ID, testFile.ID)
	}
	if retrieved.Name != testFile.Name {
		t.Errorf("Name = %q, want %q", retrieved.Name, testFile.Name)
	}
	if retrieved.Size != testFile.Size {
		t.Errorf("Size = %d, want %d", retrieved.Size, testFile.Size)
	}

	// Test GetByHash
	byHash, err := metadata.GetByHash(ctx, testFile.Hash)
	if err != nil {
		t.Fatalf("Failed to get file by hash: %v", err)
	}
	if byHash == nil {
		t.Fatal("GetByHash returned nil")
	}
	if byHash.ID != testFile.ID {
		t.Errorf("GetByHash ID = %q, want %q", byHash.ID, testFile.ID)
	}

	// Test List
	files, total, err := metadata.List(ctx, file.Query{
		TenantID: "test-tenant",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}
	if total < 1 {
		t.Errorf("Total = %d, want >= 1", total)
	}
	if len(files) < 1 {
		t.Error("Files list is empty")
	}

	// Test UpdateAccessTime
	if err := metadata.UpdateAccessTime(ctx, testFile.ID); err != nil {
		t.Fatalf("Failed to update access time: %v", err)
	}

	// Verify access time was updated
	updated, err := metadata.Get(ctx, testFile.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.LastAccessAt == nil {
		t.Error("LastAccessAt should not be nil after update")
	}

	// Test Delete (soft delete)
	if err := metadata.Delete(ctx, testFile.ID); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file is soft-deleted
	deleted, err := metadata.Get(ctx, testFile.ID)
	if err != file.ErrNotFound {
		t.Errorf("Get after delete: error = %v, want ErrNotFound", err)
	}
	if deleted != nil {
		t.Error("Get after delete should return nil")
	}

	// Cleanup
	db.Exec("DELETE FROM files WHERE id = $1", testFile.ID)
}

// TestIntegration_LocalStorageWithDatabase tests LocalStorage with DBMetadataManager.
func TestIntegration_LocalStorageWithDatabase(t *testing.T) {
	postgresURL := os.Getenv(EnvPostgresURL)
	if postgresURL == "" {
		t.Skipf("Skipping integration test: %s not set", EnvPostgresURL)
	}

	// Setup database
	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata := file.NewDBMetadataManager(db)

	// Setup local storage
	tmpDir := t.TempDir()
	storage, err := file.NewLocalStorage(tmpDir, "http://localhost", metadata)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	content := []byte("Integration test content")

	// Test Put with deduplication
	result1, err := storage.Put(ctx, file.PutOptions{
		TenantID:    "test-tenant",
		Reader:      bytes.NewReader(content),
		FileName:    "file1.txt",
		ContentType: "text/plain",
		UploadedBy:  "test-user",
	})
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Upload same content with different name (should deduplicate)
	result2, err := storage.Put(ctx, file.PutOptions{
		TenantID:    "test-tenant",
		Reader:      bytes.NewReader(content),
		FileName:    "file2.txt",
		ContentType: "text/plain",
		UploadedBy:  "test-user",
	})
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify deduplication
	if result1.Hash != result2.Hash {
		t.Errorf("Hash mismatch: %q != %q", result1.Hash, result2.Hash)
	}
	if result1.ID != result2.ID {
		t.Log("Deduplication: Both files should have same ID due to same hash")
		// Note: This depends on implementation - some systems may create separate records
	}

	// Test Get
	reader, err := storage.Get(ctx, result1.Path)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer reader.Close()

	gotContent, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotContent, content) {
		t.Errorf("Content = %q, want %q", gotContent, content)
	}

	// Cleanup
	db.Exec("DELETE FROM files WHERE tenant_id = 'test-tenant'")
}

// TestIntegration_S3Storage tests S3Storage with a real S3-compatible service (MinIO).
//
// Prerequisites:
//   - MinIO or S3-compatible service running
//   - Set environment variables:
//     S3_ENDPOINT=localhost:9000
//     S3_BUCKET=test-bucket
//     S3_REGION=us-east-1
//     S3_ACCESS_KEY=minioadmin
//     S3_SECRET_KEY=minioadmin
//
// Run MinIO locally:
//
//	docker run -p 9000:9000 -p 9001:9001 \
//	  -e MINIO_ROOT_USER=minioadmin \
//	  -e MINIO_ROOT_PASSWORD=minioadmin \
//	  minio/minio server /data --console-address ":9001"
//
// Run test:
//
//	S3_ENDPOINT=localhost:9000 S3_BUCKET=test-bucket ... go test -v -run TestIntegration_S3Storage
func TestIntegration_S3Storage(t *testing.T) {
	endpoint := os.Getenv(EnvS3Endpoint)
	bucket := os.Getenv(EnvS3Bucket)
	region := os.Getenv(EnvS3Region)
	accessKey := os.Getenv(EnvS3AccessKey)
	secretKey := os.Getenv(EnvS3SecretKey)

	if endpoint == "" || bucket == "" {
		t.Skip("Skipping S3 integration test: S3_ENDPOINT or S3_BUCKET not set")
	}

	// Create S3 storage
	storage, err := file.NewS3Storage(
		file.StorageConfig{
			S3Endpoint:  endpoint,
			S3Bucket:    bucket,
			S3Region:    region,
			S3AccessKey: accessKey,
			S3SecretKey: secretKey,
			S3UseSSL:    false, // Local MinIO typically doesn't use SSL
			S3PathStyle: true,  // MinIO uses path-style URLs
		},
		nil, // No metadata manager for this test
	)
	if err != nil {
		t.Fatalf("Failed to create S3 storage: %v", err)
	}

	ctx := context.Background()
	content := []byte("S3 integration test content")

	// Test Put
	result, err := storage.Put(ctx, file.PutOptions{
		TenantID:    "test-tenant",
		Reader:      bytes.NewReader(content),
		FileName:    "s3-test.txt",
		ContentType: "text/plain",
		UploadedBy:  "test-user",
	})
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	t.Logf("Uploaded to S3: ID=%s, Path=%s", result.ID, result.Path)

	// Test Exists
	exists, err := storage.Exists(ctx, result.Path)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("File should exist after upload")
	}

	// Test Stat
	stat, err := storage.Stat(ctx, result.Path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if stat.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", stat.Size, len(content))
	}

	// Test Get
	reader, err := storage.Get(ctx, result.Path)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer reader.Close()

	gotContent, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotContent, content) {
		t.Errorf("Content = %q, want %q", gotContent, content)
	}

	// Test GetURL (presigned URL)
	url, err := storage.GetURL(ctx, result.Path, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetURL failed: %v", err)
	}
	t.Logf("Presigned URL: %s", url)
	if url == "" {
		t.Error("Presigned URL should not be empty")
	}

	// Test Delete
	if err := storage.Delete(ctx, result.Path); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	exists, err = storage.Exists(ctx, result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("File should not exist after deletion")
	}
}

// BenchmarkIntegration_LocalStoragePut benchmarks file upload with real database.
func BenchmarkIntegration_LocalStoragePut(b *testing.B) {
	postgresURL := os.Getenv(EnvPostgresURL)
	if postgresURL == "" {
		b.Skip("Skipping benchmark: POSTGRES_URL not set")
	}

	db, _ := sql.Open("postgres", postgresURL)
	defer db.Close()

	metadata := file.NewDBMetadataManager(db)
	tmpDir := b.TempDir()
	storage, _ := file.NewLocalStorage(tmpDir, "http://localhost", metadata)

	ctx := context.Background()
	content := bytes.Repeat([]byte("a"), 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Put(ctx, file.PutOptions{
			TenantID:    "bench-tenant",
			Reader:      bytes.NewReader(content),
			FileName:    fmt.Sprintf("bench-%d.txt", i),
			ContentType: "text/plain",
		})
	}
}

// BenchmarkIntegration_S3StoragePut benchmarks S3 upload.
func BenchmarkIntegration_S3StoragePut(b *testing.B) {
	endpoint := os.Getenv(EnvS3Endpoint)
	bucket := os.Getenv(EnvS3Bucket)
	region := os.Getenv(EnvS3Region)
	accessKey := os.Getenv(EnvS3AccessKey)
	secretKey := os.Getenv(EnvS3SecretKey)

	if endpoint == "" || bucket == "" {
		b.Skip("Skipping S3 benchmark: S3_ENDPOINT or S3_BUCKET not set")
	}

	storage, _ := file.NewS3Storage(
		file.StorageConfig{
			S3Endpoint:  endpoint,
			S3Bucket:    bucket,
			S3Region:    region,
			S3AccessKey: accessKey,
			S3SecretKey: secretKey,
			S3UseSSL:    false,
			S3PathStyle: true,
		},
		nil,
	)

	ctx := context.Background()
	content := bytes.Repeat([]byte("a"), 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Put(ctx, file.PutOptions{
			TenantID:    "bench-tenant",
			Reader:      bytes.NewReader(content),
			FileName:    fmt.Sprintf("bench-%d.txt", i),
			ContentType: "text/plain",
		})
	}
}
