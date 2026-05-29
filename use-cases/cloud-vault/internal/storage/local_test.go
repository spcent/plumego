package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorage_PutGetDelete(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "test/file.md"
	content := []byte("# Hello World\n\nThis is a test.")

	// Put
	err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/markdown")
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Exists
	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Error("Exists returned false after Put")
	}

	// Get
	rc, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("Get content = %q, want %q", got, content)
	}

	// Delete
	err = store.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Should not exist after delete
	exists, err = store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists after Delete: %v", err)
	}
	if exists {
		t.Error("Exists returned true after Delete")
	}
}

func TestLocalStorage_GetNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	_, err := store.Get(ctx, "nonexistent.txt")
	if err != ErrNotFound {
		t.Errorf("Get nonexistent: got %v, want ErrNotFound", err)
	}
}

func TestLocalStorage_ExistsNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	exists, err := store.Exists(ctx, "nonexistent.txt")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Error("Exists returned true for nonexistent key")
	}
}

func TestLocalStorage_DeleteNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	// Delete nonexistent should succeed (idempotent)
	err := store.Delete(ctx, "nonexistent.txt")
	if err != nil {
		t.Errorf("Delete nonexistent: %v", err)
	}
}

func TestLocalStorage_NestedKeys(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "documents/2026/01/test.md"
	content := []byte("nested content")

	err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/markdown")
	if err != nil {
		t.Fatalf("Put nested: %v", err)
	}

	// Verify file exists at correct path
	filePath := filepath.Join(tmpDir, filepath.FromSlash(key))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("nested file not created at %q", filePath)
	}

	// Get should work
	rc, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get nested: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("nested content = %q, want %q", got, content)
	}
}

func TestLocalStorage_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "overwrite.txt"
	content1 := []byte("version 1")
	content2 := []byte("version 2 updated")

	// Put first version
	err := store.Put(ctx, key, bytes.NewReader(content1), int64(len(content1)), "text/plain")
	if err != nil {
		t.Fatalf("Put v1: %v", err)
	}

	// Put second version (overwrite)
	err = store.Put(ctx, key, bytes.NewReader(content2), int64(len(content2)), "text/plain")
	if err != nil {
		t.Fatalf("Put v2: %v", err)
	}

	// Get should return v2
	rc, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !bytes.Equal(got, content2) {
		t.Errorf("Get after overwrite = %q, want %q", got, content2)
	}
}

func TestLocalStorage_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "empty.txt"
	content := []byte{}

	err := store.Put(ctx, key, bytes.NewReader(content), 0, "text/plain")
	if err != nil {
		t.Fatalf("Put empty: %v", err)
	}

	rc, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get empty: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Get empty = %d bytes, want 0", len(got))
	}
}

func TestLocalStorage_KeyWithSpaces(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "path with spaces/file name.md"
	content := []byte("content with spaces in path")

	err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/markdown")
	if err != nil {
		t.Fatalf("Put with spaces: %v", err)
	}

	rc, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get with spaces: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch")
	}
}

func TestLocalStorage_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	// 1MB file
	key := "large.bin"
	content := []byte(strings.Repeat("x", 1024*1024))

	err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "application/octet-stream")
	if err != nil {
		t.Fatalf("Put large: %v", err)
	}

	rc, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get large: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll large: %v", err)
	}
	if len(got) != len(content) {
		t.Errorf("large file size = %d, want %d", len(got), len(content))
	}
}
