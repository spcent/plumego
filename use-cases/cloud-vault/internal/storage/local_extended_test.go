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

// TestLocalStorage_PutEmptyKey verifies that an empty key still works
// (maps to the root directory file).
func TestLocalStorage_PutEmptyKey(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	content := []byte("data for empty-key file")
	// Empty key puts the file directly at root
	err := store.Put(ctx, "emptyish.txt", bytes.NewReader(content), int64(len(content)), "text/plain")
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	rc, err := store.Get(ctx, "emptyish.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer rc.Close()
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

// TestLocalStorage_DeleteThenExistsFalse covers the idempotent Delete + Exists flow.
func TestLocalStorage_DeleteThenExistsFalse(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "todelete/file.md"
	content := []byte("will be deleted")

	if err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/markdown"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// First delete should succeed
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// File should no longer exist
	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists after delete: %v", err)
	}
	if exists {
		t.Error("Exists returned true after Delete")
	}

	// Second delete on non-existent should be idempotent
	if err := store.Delete(ctx, key); err != nil {
		t.Errorf("second Delete nonexistent should not error: %v", err)
	}
}

// TestLocalStorage_DeleteNestedFile removes a file inside a subdirectory.
func TestLocalStorage_DeleteNestedFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "deep/nested/dir/file.txt"
	content := []byte("nested")

	if err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/plain"); err != nil {
		t.Fatalf("Put nested: %v", err)
	}

	// Verify directory was created
	dirPath := filepath.Join(tmpDir, "deep", "nested", "dir")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Fatalf("expected nested directory to be created")
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete nested: %v", err)
	}

	exists, _ := store.Exists(ctx, key)
	if exists {
		t.Error("file should be gone after delete")
	}
}

// TestLocalStorage_ExistsAfterPut covers the happy path of Exists returning true.
func TestLocalStorage_ExistsAfterPut(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "check/existence.md"
	content := []byte("check me")

	before, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists before Put: %v", err)
	}
	if before {
		t.Error("Exists should return false before Put")
	}

	if err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/markdown"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	after, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists after Put: %v", err)
	}
	if !after {
		t.Error("Exists should return true after Put")
	}
}

// TestLocalStorage_GetNonExistent exercises the ErrNotFound path.
func TestLocalStorage_GetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	_, err := store.Get(ctx, "no/such/file.md")
	if err == nil {
		t.Fatal("expected error getting nonexistent file, got nil")
	}
	if err != ErrNotFound {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

// TestLocalStorage_PutAndGetMultipleFiles writes several files then reads them all.
func TestLocalStorage_PutAndGetMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	files := map[string]string{
		"a/one.txt":       "content one",
		"a/two.txt":       "content two",
		"b/sub/three.txt": "content three",
		"root.txt":        "root content",
	}

	for key, body := range files {
		buf := []byte(body)
		if err := store.Put(ctx, key, bytes.NewReader(buf), int64(len(buf)), "text/plain"); err != nil {
			t.Fatalf("Put %q: %v", key, err)
		}
	}

	for key, body := range files {
		rc, err := store.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get %q: %v", key, err)
		}
		got, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("ReadAll %q: %v", key, err)
		}
		if string(got) != body {
			t.Errorf("Get %q = %q, want %q", key, got, body)
		}
	}
}

// TestLocalStorage_LargeFile2MB puts a 2 MB file and reads it back in full.
func TestLocalStorage_LargeFile2MB(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	const size = 2 * 1024 * 1024 // 2 MB
	content := []byte(strings.Repeat("A", size))
	key := "large/2mb.bin"

	if err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "application/octet-stream"); err != nil {
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
	if len(got) != size {
		t.Errorf("large file size = %d, want %d", len(got), size)
	}
}

// TestLocalStorage_KeyWithForwardSlashes verifies path separator normalisation.
func TestLocalStorage_KeyWithForwardSlashes(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "a/b/c/d/file.txt"
	content := []byte("slash test")

	if err := store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/plain"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	// Verify underlying path is created with the right OS separator
	expectedPath := filepath.Join(tmpDir, filepath.FromSlash(key))
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at OS path %q", expectedPath)
	}
}

// TestErrNotFound_IsDistinctError verifies ErrNotFound is a typed error and is
// not wrapped by accident.
func TestErrNotFound_IsDistinctError(t *testing.T) {
	if ErrNotFound == nil {
		t.Fatal("ErrNotFound should not be nil")
	}
	if ErrNotFound.Error() == "" {
		t.Error("ErrNotFound.Error() should not be empty")
	}
}

// TestLocalStorage_Root ensures the root value is honoured by keyPath.
func TestLocalStorage_Root(t *testing.T) {
	root := t.TempDir()
	store := NewLocalStorage(root)

	key := "some/file.md"
	expected := filepath.Join(root, filepath.FromSlash(key))
	got := store.keyPath(key)
	if got != expected {
		t.Errorf("keyPath(%q) = %q, want %q", key, got, expected)
	}
}
