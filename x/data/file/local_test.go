package file

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	storefile "github.com/spcent/plumego/store/file"
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

	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Base directory was not created")
	}
}

func TestLocalStorage_Put_Get(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	content := []byte("Hello, World!")

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

	reader, err := storage.Get(ctx, result.Path)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer reader.Close()

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
		{"test", "image/jpeg", ".jpg"},
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

	result, err := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   strings.NewReader("test content"),
		FileName: "test.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	exists, err := storage.Exists(ctx, result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("File should exist after Put")
	}

	if err := storage.Delete(ctx, result.Path); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

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

	err = storage.Delete(context.Background(), "tenant-123/2026/02/05/nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	var fileErr *storefile.Error
	if errors.As(err, &fileErr) {
		if fileErr.Err != storefile.ErrNotFound {
			t.Errorf("Error = %v, want ErrNotFound", fileErr.Err)
		}
	} else {
		t.Errorf("Error type = %T, want *storefile.Error", err)
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

	result, err := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   bytes.NewReader(content),
		FileName: "test.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	stat, err := storage.Stat(ctx, result.Path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

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

	files, err := storage.List(ctx, "tenant-123", 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(files) != 5 {
		t.Errorf("File count = %d, want 5", len(files))
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Path, "tenant-123") {
			t.Errorf("Path %q does not start with tenant-123", file.Path)
		}
	}
}

func TestLocalStorage_GetURL(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewLocalStorage(tmpDir, "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	result, err := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   strings.NewReader("test content"),
		FileName: "test.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	url, err := storage.GetURL(ctx, result.Path, time.Hour)
	if err != nil {
		t.Fatalf("GetURL failed: %v", err)
	}

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

	result, err := storage.Put(ctx, PutOptions{
		TenantID: "tenant-123",
		Reader:   bytes.NewReader(content),
		FileName: "source.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	dstPath := "tenant-123/2026/02/05/copy.txt"
	if err := storage.Copy(ctx, result.Path, dstPath); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	exists, err := storage.Exists(ctx, dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("Copied file should exist")
	}

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
	attacks := []string{
		"../../../etc/passwd",
		"/etc/passwd",
		"tenant/../../../etc/passwd",
	}

	for _, attack := range attacks {
		t.Run(attack, func(t *testing.T) {
			if _, err := storage.Get(ctx, attack); err == nil {
				t.Error("Expected error for path traversal attack")
			}
			if err := storage.Delete(ctx, attack); err == nil {
				t.Error("Expected error for path traversal attack")
			}
			if _, err := storage.Exists(ctx, attack); err == nil {
				t.Error("Expected error for path traversal attack")
			}
		})
	}
}

func TestLocalStorage_Get_NotFound(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = storage.Get(context.Background(), "tenant/2026/01/01/nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
	var fileErr *storefile.Error
	if !errors.As(err, &fileErr) {
		t.Fatalf("expected *storefile.Error, got %T", err)
	}
	if !errors.Is(fileErr.Err, storefile.ErrNotFound) {
		t.Errorf("inner error = %v, want ErrNotFound", fileErr.Err)
	}
}

func TestLocalStorage_Stat_NotFound(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = storage.Stat(context.Background(), "tenant/2026/01/01/nonexistent.txt")
	if !errors.Is(err, storefile.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLocalStorage_Stat_InvalidPath(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = storage.Stat(context.Background(), "../etc/passwd")
	if !errors.Is(err, storefile.ErrInvalidPath) {
		t.Errorf("expected ErrInvalidPath, got %v", err)
	}
}

func TestLocalStorage_Put_Deduplication(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir(), "http://example.com", &mockMetadata{})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	content := []byte("deduplicated content")

	first, err := storage.Put(ctx, PutOptions{
		TenantID: "t1",
		Reader:   bytes.NewReader(content),
		FileName: "dup.txt",
	})
	if err != nil {
		t.Fatalf("first Put: %v", err)
	}

	second, err := storage.Put(ctx, PutOptions{
		TenantID: "t1",
		Reader:   bytes.NewReader(content),
		FileName: "dup.txt",
	})
	if err != nil {
		t.Fatalf("second Put: %v", err)
	}
	if second.Hash != first.Hash {
		t.Errorf("expected same hash for duplicate content, first=%q second=%q", first.Hash, second.Hash)
	}
}

// --- mock MetadataManager ---

type mockMetadata struct {
	store map[string]*File
}

func (m *mockMetadata) Save(_ context.Context, f *File) error {
	if m.store == nil {
		m.store = make(map[string]*File)
	}
	m.store[f.Hash] = f
	return nil
}

func (m *mockMetadata) Get(_ context.Context, id string) (*File, error) {
	if m.store == nil {
		return nil, storefile.ErrNotFound
	}
	for _, f := range m.store {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, storefile.ErrNotFound
}

func (m *mockMetadata) GetByPath(_ context.Context, path string) (*File, error) {
	if m.store == nil {
		return nil, storefile.ErrNotFound
	}
	for _, f := range m.store {
		if f.Path == path {
			return f, nil
		}
	}
	return nil, storefile.ErrNotFound
}

func (m *mockMetadata) GetByHash(_ context.Context, hash string) (*File, error) {
	if m.store == nil {
		return nil, storefile.ErrNotFound
	}
	if f, ok := m.store[hash]; ok {
		return f, nil
	}
	return nil, storefile.ErrNotFound
}

func (m *mockMetadata) List(_ context.Context, _ Query) ([]*File, int64, error) {
	return nil, 0, nil
}

func (m *mockMetadata) Delete(_ context.Context, _ string) error { return nil }

func (m *mockMetadata) UpdateAccessTime(_ context.Context, _ string) error { return nil }

// Compile-time check
var _ MetadataManager = (*mockMetadata)(nil)

// Benchmarks

func BenchmarkLocalStorage_Put(b *testing.B) {
	tmpDir := b.TempDir()
	storage, _ := NewLocalStorage(tmpDir, "http://example.com", nil)
	ctx := context.Background()
	content := bytes.Repeat([]byte("a"), 1024)

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
