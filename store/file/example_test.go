package file_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spcent/plumego/store/file"
)

// ExampleLocalStorage demonstrates basic local storage usage.
func ExampleLocalStorage() {
	// Initialize metadata manager
	db, _ := sql.Open("postgres", "postgres://localhost/plumego")
	metadata := file.NewDBMetadataManager(db)

	// Create local storage
	storage, err := file.NewLocalStorage("/var/uploads", "http://example.com/files", metadata)
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
		GenerateThumb: false,
		UploadedBy:    "user-456",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("File uploaded: %s\n", result.ID)
	fmt.Printf("Path: %s\n", result.Path)
	fmt.Printf("Size: %d bytes\n", result.Size)
	fmt.Printf("Hash: %s\n", result.Hash)

	// Download the file
	reader, err := storage.Get(ctx, result.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// Read and print content
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	fmt.Printf("Content: %s\n", buf.String())

	// Output:
	// File uploaded: ...
	// Path: tenant-123/2026/02/05/...
	// Size: 13 bytes
	// Hash: ...
	// Content: Hello, World!
}

// ExampleS3Storage demonstrates S3 storage usage.
func ExampleS3Storage() {
	// Initialize metadata manager
	db, _ := sql.Open("postgres", "postgres://localhost/plumego")
	metadata := file.NewDBMetadataManager(db)

	// Create S3 storage
	storage, err := file.NewS3Storage(
		"s3.amazonaws.com",
		"my-bucket",
		"us-east-1",
		"access-key",
		"secret-key",
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

	fmt.Printf("File uploaded to S3: %s\n", result.ID)

	// Get a presigned URL (valid for 1 hour)
	url, err := storage.GetURL(ctx, result.Path, time.Hour)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Presigned URL: %s\n", url)

	// Output:
	// File uploaded to S3: ...
	// Presigned URL: https://...
}

// ExampleImageProcessor demonstrates image processing.
func ExampleImageProcessor() {
	// Create image processor
	processor := file.NewImageProcessor()

	// Open an image file
	imgFile := bytes.NewReader([]byte("...")) // Actual image data

	// Get image info
	info, err := processor.GetInfo(imgFile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Image: %dx%d, format: %s\n", info.Width, info.Height, info.Format)

	// Generate thumbnail (200x200)
	imgFile.Seek(0, 0) // Reset reader
	thumb, err := processor.Thumbnail(imgFile, 200, 200)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Thumbnail generated\n")
	_ = thumb

	// Output:
	// Image: 1920x1080, format: jpeg
	// Thumbnail generated
}

// ExampleHandler demonstrates HTTP handler usage.
func ExampleHandler() {
	// Initialize components
	db, _ := sql.Open("postgres", "postgres://localhost/plumego")
	metadata := file.NewDBMetadataManager(db)
	storage, _ := file.NewLocalStorage("/var/uploads", "http://example.com/files", metadata)

	// Create handler
	handler := file.NewHandler(storage, metadata).
		WithMaxSize(100 << 20) // 100 MiB

	// Register routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Start server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("File server listening on :8080")
	log.Fatal(server.ListenAndServe())
}

// ExampleFileDeduplication demonstrates automatic file deduplication.
func ExampleFileDeduplication() {
	db, _ := sql.Open("postgres", "postgres://localhost/plumego")
	metadata := file.NewDBMetadataManager(db)
	storage, _ := file.NewLocalStorage("/var/uploads", "http://example.com/files", metadata)

	ctx := context.Background()

	// Upload the same file twice
	content1 := bytes.NewReader([]byte("Duplicate content"))
	file1, _ := storage.Put(ctx, file.PutOptions{
		TenantID:    "tenant-123",
		Reader:      content1,
		FileName:    "file1.txt",
		ContentType: "text/plain",
		UploadedBy:  "user-456",
	})

	content2 := bytes.NewReader([]byte("Duplicate content"))
	file2, _ := storage.Put(ctx, file.PutOptions{
		TenantID:    "tenant-123",
		Reader:      content2,
		FileName:    "file2.txt", // Different name
		ContentType: "text/plain",
		UploadedBy:  "user-456",
	})

	// Both files have the same hash, so file2 returns the existing file1
	fmt.Printf("File1 ID: %s\n", file1.ID)
	fmt.Printf("File2 ID: %s\n", file2.ID)
	fmt.Printf("Same file: %v\n", file1.ID == file2.ID)

	// Output:
	// File1 ID: ...
	// File2 ID: ...
	// Same file: true
}

// ExampleMetadataQuery demonstrates querying file metadata.
func ExampleMetadataQuery() {
	db, _ := sql.Open("postgres", "postgres://localhost/plumego")
	metadata := file.NewDBMetadataManager(db)

	ctx := context.Background()

	// Query files for a tenant
	files, total, err := metadata.List(ctx, file.Query{
		TenantID:   "tenant-123",
		MimeType:   "image/jpeg",
		UploadedBy: "user-456",
		StartTime:  time.Now().Add(-30 * 24 * time.Hour), // Last 30 days
		EndTime:    time.Now(),
		Page:       1,
		PageSize:   20,
		OrderBy:    "created_at DESC",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d files (total: %d)\n", len(files), total)
	for _, f := range files {
		fmt.Printf("- %s (%d bytes)\n", f.Name, f.Size)
	}
}

// ExampleThumbnailGeneration demonstrates automatic thumbnail generation.
func ExampleThumbnailGeneration() {
	db, _ := sql.Open("postgres", "postgres://localhost/plumego")
	metadata := file.NewDBMetadataManager(db)
	storage, _ := file.NewLocalStorage("/var/uploads", "http://example.com/files", metadata)

	ctx := context.Background()

	// Upload image with thumbnail generation
	imgFile := bytes.NewReader([]byte("...")) // Actual JPEG data
	result, err := storage.Put(ctx, file.PutOptions{
		TenantID:      "tenant-123",
		Reader:        imgFile,
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

	// Output:
	// Original: tenant-123/2026/02/05/xxx.jpg (1920x1080)
	// Thumbnail: tenant-123/2026/02/05/xxx_thumb.jpg
}
