package file_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/spcent/plumego/store/file"
)

// ExampleLocalStorage demonstrates basic local storage usage.
//
// This example shows how to initialize local storage and upload/download files.
// Note: This example requires a PostgreSQL database and filesystem access.
func ExampleLocalStorage() {
	// Initialize metadata manager (optional, can be nil)
	var metadata file.MetadataManager
	// For production, use: db, _ := sql.Open("postgres", "connection-string")
	// metadata = file.NewDBMetadataManager(db)

	// Create local storage
	storage, err := file.NewLocalStorage(
		"/tmp/uploads",              // Base directory
		"http://example.com/files",  // Base URL
		metadata,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Upload a file
	content := bytes.NewReader([]byte("Hello, World!"))
	result, err := storage.Put(ctx, file.PutOptions{
		TenantID:      "tenant-123",
		Reader:        content,
		FileName:      "hello.txt",
		ContentType:   "text/plain",
		UploadedBy:    "user-456",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Uploaded file ID: %s\n", result.ID)
	fmt.Printf("Path: %s\n", result.Path)
}

// ExampleS3Storage demonstrates S3-compatible storage usage.
func ExampleS3Storage() {
	db, _ := sql.Open("postgres", "connection-string")
	metadata := file.NewDBMetadataManager(db)

	// Create S3 storage
	storage, err := file.NewS3Storage(
		file.StorageConfig{
			S3Endpoint:  "s3.amazonaws.com",
			S3Bucket:    "my-bucket",
			S3Region:    "us-east-1",
			S3AccessKey: "access-key",
			S3SecretKey: "secret-key",
			S3UseSSL:    true,
		},
		metadata,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Upload a file
	content := bytes.NewReader([]byte("Hello, S3!"))
	result, err := storage.Put(ctx, file.PutOptions{
		TenantID:    "tenant-123",
		Reader:      content,
		FileName:    "hello.txt",
		ContentType: "text/plain",
		UploadedBy:  "user-456",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Get a presigned URL (valid for 1 hour)
	url, err := storage.GetURL(ctx, result.Path, time.Hour)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Presigned URL: %s\n", url)
}

// Example_imageProcessing demonstrates image thumbnail generation.
func Example_imageProcessing() {
	db, _ := sql.Open("postgres", "connection-string")
	metadata := file.NewDBMetadataManager(db)
	storage, _ := file.NewLocalStorage("/tmp/uploads", "http://example.com/files", metadata)

	ctx := context.Background()

	// Assume imgReader contains JPEG image data
	var imgReader *bytes.Reader

	// Upload image with thumbnail generation
	result, err := storage.Put(ctx, file.PutOptions{
		TenantID:      "tenant-123",
		Reader:        imgReader,
		FileName:      "photo.jpg",
		ContentType:   "image/jpeg",
		GenerateThumb: true,
		ThumbWidth:    200,
		ThumbHeight:   200,
		UploadedBy:    "user-456",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original: %s (%dx%d)\n", result.Path, result.Width, result.Height)
	fmt.Printf("Thumbnail: %s\n", result.ThumbnailPath)
}

// Example_fileDeduplication demonstrates automatic file deduplication.
// Note: Deduplication requires a metadata manager to be configured.
func Example_fileDeduplication() {
	// Deduplication requires a metadata manager
	// db, _ := sql.Open("postgres", "connection-string")
	// metadata := file.NewDBMetadataManager(db)
	// storage, _ := file.NewLocalStorage("/tmp/uploads", "http://example.com/files", metadata)

	// Without metadata manager, files are not deduplicated
	storage, _ := file.NewLocalStorage("/tmp/uploads", "http://example.com/files", nil)

	ctx := context.Background()

	// Upload the same content twice with different names
	content1 := bytes.NewReader([]byte("Duplicate content"))
	file1, _ := storage.Put(ctx, file.PutOptions{
		TenantID:    "tenant-123",
		Reader:      content1,
		FileName:    "file1.txt",
		ContentType: "text/plain",
	})

	content2 := bytes.NewReader([]byte("Duplicate content"))
	file2, _ := storage.Put(ctx, file.PutOptions{
		TenantID:    "tenant-123",
		Reader:      content2,
		FileName:    "file2.txt",
		ContentType: "text/plain",
	})

	// Without metadata, each upload creates a new file
	// With metadata manager, both uploads would return the same file (deduplication)
	fmt.Printf("Different files: %v\n", file1.ID != file2.ID)
}
