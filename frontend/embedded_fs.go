package frontend

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/spcent/plumego/router"
)

// embedded contains optional frontend assets placed under the embedded/ directory.
//
//go:embed embedded/*
var embedded embed.FS

// RegisterEmbedded mounts the embedded frontend bundle.
// Returns an error if no embedded frontend assets are found.
func RegisterEmbedded(r *router.Router, opts ...Option) error {
	mount, err := NewMountEmbedded(opts...)
	if err != nil {
		return err
	}
	return mount.Register(r)
}

// NewMountEmbedded constructs a frontend mount from embedded assets.
func NewMountEmbedded(opts ...Option) (*Mount, error) {
	f, err := embeddedSub()
	if err != nil {
		return nil, err
	}
	return NewMountFS(http.FS(f), opts...)
}

// embeddedSub returns the embedded filesystem subdirectory containing frontend assets.
// Returns an error if no assets are found or if the embedded directory is inaccessible.
func embeddedSub() (fs.FS, error) {
	const root = "embedded"

	entries, err := fs.ReadDir(embedded, root)
	if err != nil {
		return nil, fmt.Errorf("embedded frontend directory %q not found: %w", root, err)
	}

	// Consider the directory non-empty only when it contains real files (not .keep)
	// or non-empty subdirectories.
	hasContent := false
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != ".keep" {
			hasContent = true
			break
		}
		if entry.IsDir() {
			subEntries, err := fs.ReadDir(embedded, root+"/"+entry.Name())
			if err == nil && len(subEntries) > 0 {
				hasContent = true
				break
			}
		}
	}

	if !hasContent {
		return nil, errors.New("no embedded frontend assets found (embedded/ directory is empty)")
	}

	subFS, err := fs.Sub(embedded, root)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedded subdirectory: %w", err)
	}
	return subFS, nil
}
