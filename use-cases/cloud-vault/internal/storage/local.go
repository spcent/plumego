package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage stores objects on the local filesystem under a root directory.
// Object keys map to files: root/key (with forward slashes converted to OS separators).
type LocalStorage struct {
	root string
}

// NewLocalStorage creates a LocalStorage rooted at root.
func NewLocalStorage(root string) *LocalStorage {
	return &LocalStorage{root: root}
}

func (s *LocalStorage) Put(_ context.Context, key string, body io.Reader, _ int64, _ string) error {
	path, err := s.keyPath(key)
	if err != nil {
		return fmt.Errorf("local storage put %q: %w", key, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("local storage put %q: create directory: %w", key, err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("local storage put %q: create file: %w", key, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("local storage put %q: write: %w", key, err)
	}
	return nil
}

func (s *LocalStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	path, err := s.keyPath(key)
	if err != nil {
		return nil, fmt.Errorf("local storage get %q: %w", key, err)
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("local storage get %q: %w", key, err)
	}
	return f, nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	path, err := s.keyPath(key)
	if err != nil {
		return fmt.Errorf("local storage delete %q: %w", key, err)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local storage delete %q: %w", key, err)
	}
	return nil
}

func (s *LocalStorage) Exists(_ context.Context, key string) (bool, error) {
	path, err := s.keyPath(key)
	if err != nil {
		return false, fmt.Errorf("local storage exists %q: %w", key, err)
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("local storage exists %q: %w", key, err)
}

// Ping verifies the storage root directory is accessible.
func (s *LocalStorage) Ping(_ context.Context) error {
	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return fmt.Errorf("local storage: resolve root: %w", err)
	}
	_, err = os.Stat(absRoot)
	if err != nil {
		return fmt.Errorf("local storage: root not accessible: %w", err)
	}
	return nil
}

// keyPath resolves a storage key to an absolute filesystem path, validating
// that the result stays within the storage root (prevents path traversal).
func (s *LocalStorage) keyPath(key string) (string, error) {
	if strings.ContainsRune(key, 0) {
		return "", fmt.Errorf("key contains null byte")
	}
	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return "", fmt.Errorf("resolve storage root: %w", err)
	}
	resolved := filepath.Clean(filepath.Join(absRoot, filepath.FromSlash(key)))
	// Ensure the resolved path is within root (add separator to prevent prefix matching e.g. /data vs /data2)
	if resolved != absRoot && !strings.HasPrefix(resolved, absRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("key escapes storage root")
	}
	return resolved, nil
}
