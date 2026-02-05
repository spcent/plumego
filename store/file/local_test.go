package file

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewLocalStorage(t *testing.T) {
	tmpDir := t.TempDir()

	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewLocalStorage failed: %v", err)
	}

	if storage == nil {
		t.Fatal("Storage is nil")
	}

	// Verify directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Base directory was not created")
	}
}

func TestNewLocalStorage_InvalidPath(t *testing.T) {
	// Skip this test as MkdirAll might succeed depending on permissions
	t.Skip("Skipping invalid path test as it's environment-dependent")
}

func TestLocalStorage_Put_Get(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	content := []byte("Hello, World!")

	// Put file
	result, err := storage.Put(ctx, PutOptions{
		TenantID:    "tenant-123",
		Reader:      bytes.NewReader(content),
		FileName:    "test.txt",
		ContentType: "text/plain",
		UploadedBy:  "user-456",
	})
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify result
	if result.ID == "" {
		t.Error("File ID is empty")
	}
	if result.TenantID != "tenant-123" {
		t.Errorf("TenantID = %q, want %q", result.TenantID, "tenant-123")
	}
	if result.Name != "test.txt" {
		t.Errorf("Name = %q, want %q", result.Name, "test.txt")
	}
	if result.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", result.Size, len(content))
	}
	if result.StorageType != "local" {
		t.Errorf("StorageType = %q, want %q", result.StorageType, "local")
	}

	// Get file
	reader, err := storage.Get(ctx, result.Path)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer reader.Close()

	// Verify content
	gotContent, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if !bytes.Equal(gotContent, content) {
		t.Errorf("Content = %q, want %q", gotContent, content)
	}
}

func TestLocalStorage_Put_WithExtension(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	tests := []struct {
		fileName    string
		contentType string
		wantExt     string
	}{
		{"test.jpg", "image/jpeg", ".jpg"},
		{"test.png", "image/png", ".png"},
		{"test", "image/jpeg", ".jpg"}, // No extension, infer from MIME
		{"test.txt", "text/plain", ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			result, err := storage.Put(ctx, PutOptions{
				TenantID:    "tenant-123",
				Reader:      strings.NewReader("test content"),
				FileName:    tt.fileName,
				ContentType: tt.contentType,
			})
			if err != nil {
				t.Fatalf("Put failed: %v", err)
			}

			if result.Extension != tt.wantExt {
				t.Errorf("Extension = %q, want %q", result.Extension, tt.wantExt)
			}

			// Verify file path has correct extension
			if !strings.HasSuffix(result.Path, tt.wantExt) {
				t.Errorf("Path %q does not end with %q", result.Path, tt.wantExt)
			}
		})
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Put file
	result, err := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   strings.NewReader("test content"),
		FileName: "test.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	exists, err := storage.Exists(ctx, result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("File should exist after Put")
	}

	// Delete file
	if err := storage.Delete(ctx, result.Path); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify file no longer exists
	exists, err = storage.Exists(ctx, result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("File should not exist after Delete")
	}
}

func TestLocalStorage_Delete_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Try to delete non-existent file
	err = storage.Delete(ctx, "tenant-123/2026/02/05/nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Check error type
	if fileErr, ok := err.(*Error); ok {
		if fileErr.Err != ErrNotFound {
			t.Errorf("Error = %v, want ErrNotFound", fileErr.Err)
		}
	} else {
		t.Errorf("Error type = %T, want *Error", err)
	}
}

func TestLocalStorage_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Check non-existent file
	exists, err := storage.Exists(ctx, "tenant-123/2026/02/05/test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("Non-existent file should not exist")
	}

	// Put file
	result, err := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   strings.NewReader("test content"),
		FileName: "test.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check existing file
	exists, err = storage.Exists(ctx, result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("Existing file should exist")
	}
}

func TestLocalStorage_Stat(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	content := []byte("test content")

	// Put file
	result, err := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   bytes.NewReader(content),
		FileName: "test.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get stat
	stat, err := storage.Stat(ctx, result.Path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// Verify stat
	if stat.Path != result.Path {
		t.Errorf("Path = %q, want %q", stat.Path, result.Path)
	}
	if stat.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", stat.Size, len(content))
	}
	if stat.IsDir {
		t.Error("IsDir should be false for file")
	}
	if stat.ModifiedTime.IsZero() {
		t.Error("ModifiedTime should not be zero")
	}
}

func TestLocalStorage_List(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Put multiple files
	for i := 0; i < 5; i++ {
		_, err := storage.Put(ctx, PutOptions{
			TenantID: "tenant-123",
			Reader:   strings.NewReader("test content"),
			FileName: "test.txt",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// List files
	files, err := storage.List(ctx, "tenant-123", 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify count
	if len(files) != 5 {
		t.Errorf("File count = %d, want 5", len(files))
	}

	// Verify all paths start with tenant prefix
	for _, file := range files {
		if !strings.HasPrefix(file.Path, "tenant-123") {
			t.Errorf("Path %q does not start with tenant-123", file.Path)
		}
	}
}

func TestLocalStorage_List_WithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Put 10 files
	for i := 0; i < 10; i++ {
		_, err := storage.Put(ctx, PutOptions{
			TenantID: "tenant-123",
			Reader:   strings.NewReader("test content"),
			FileName: "test.txt",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// List with limit of 5
	files, err := storage.List(ctx, "tenant-123", 5)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify limit was applied
	if len(files) != 5 {
		t.Errorf("File count = %d, want 5", len(files))
	}
}

func TestLocalStorage_GetURL(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Put file
	result, err := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   strings.NewReader("test content"),
		FileName: "test.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get URL
	url, err := storage.GetURL(ctx, result.Path, time.Hour)
	if err != nil {
		t.Fatalf("GetURL failed: %v", err)
	}

	// Verify URL format
	expected := "http://example.com/" + result.Path
	if url != expected {
		t.Errorf("URL = %q, want %q", url, expected)
	}
}

func TestLocalStorage_Copy(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	content := []byte("test content")

	// Put source file
	result, err := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   bytes.NewReader(content),
		FileName: "source.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Copy file
	dstPath := "tenant-123/2026/02/05/copy.txt"
	if err := storage.Copy(ctx, result.Path, dstPath); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Verify copy exists
	exists, err := storage.Exists(ctx, dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("Copied file should exist")
	}

	// Verify copied content
	reader, err := storage.Get(ctx, dstPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	copiedContent, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(copiedContent, content) {
		t.Errorf("Copied content = %q, want %q", copiedContent, content)
	}
}

func TestLocalStorage_PathSafety(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Try path traversal attacks
	attacks := []string{
		"../../../etc/passwd",
		"/etc/passwd",
		"tenant/../../../etc/passwd",
	}

	for _, attack := range attacks {
		t.Run(attack, func(t *testing.T) {
			// Get should fail
			_, err := storage.Get(ctx, attack)
			if err == nil {
				t.Error("Expected error for path traversal attack")
			}

			// Delete should fail
			err = storage.Delete(ctx, attack)
			if err == nil {
				t.Error("Expected error for path traversal attack")
			}

			// Exists should return error
			_, err = storage.Exists(ctx, attack)
			if err == nil {
				t.Error("Expected error for path traversal attack")
			}
		})
	}
}

// Benchmark local storage operations
func BenchmarkLocalStorage_Put(b *testing.B) {
	tmpDir := b.TempDir()
	storage, _ := NewLocalStorage(tmpDir, "http://example.com", nil)
	ctx := context.Background()
	content := bytes.Repeat([]byte("a"), 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Put(ctx, PutOptions{
			TenantID: "tenant-123",
			Reader:   bytes.NewReader(content),
			FileName: "test.txt",
		})
	}
}

func BenchmarkLocalStorage_Get(b *testing.B) {
	tmpDir := b.TempDir()
	storage, _ := NewLocalStorage(tmpDir, "http://example.com", nil)
	ctx := context.Background()

	// Put a file first
	result, _ := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   strings.NewReader("test content"),
		FileName: "test.txt",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader, _ := storage.Get(ctx, result.Path)
		io.Copy(io.Discard, reader)
		reader.Close()
	}
}
