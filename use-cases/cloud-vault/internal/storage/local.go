package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	path := s.keyPath(key)
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
	path := s.keyPath(key)
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
	path := s.keyPath(key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local storage delete %q: %w", key, err)
	}
	return nil
}

func (s *LocalStorage) Exists(_ context.Context, key string) (bool, error) {
	_, err := os.Stat(s.keyPath(key))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("local storage exists %q: %w", key, err)
}

func (s *LocalStorage) keyPath(key string) string {
	return filepath.Join(s.root, filepath.FromSlash(key))
}
