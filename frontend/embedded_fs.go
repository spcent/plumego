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

// embeddedFS allows tests to swap the source filesystem.
var embeddedFS fs.FS = embedded

// embeddedRoot is the directory that holds the embedded assets.
var embeddedRoot = "embedded"

// HasEmbedded reports whether any embedded frontend assets are available.
func HasEmbedded() bool {
	_, err := embeddedSub()
	return err == nil
}

// RegisterEmbedded mounts an embedded frontend bundle if present.
func RegisterEmbedded(r *router.Router, opts ...Option) error {
	f, err := embeddedSub()
	if err != nil {
		return err
	}

	return RegisterFS(r, http.FS(f), opts...)
}

// embeddedSub returns the embedded filesystem subdirectory containing frontend assets.
// It returns an error if no assets are found or if the embedded directory is inaccessible.
func embeddedSub() (fs.FS, error) {
	// Check if embedded directory exists and has content
	entries, err := fs.ReadDir(embeddedFS, embeddedRoot)
	if err != nil {
		return nil, fmt.Errorf("embedded frontend directory %q not found: %w", embeddedRoot, err)
	}

	// Check for non-hidden files (excluding .keep and other dotfiles)
	hasContent := false
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != ".keep" {
			hasContent = true
			break
		}
		// Also check for subdirectories that might contain assets
		if entry.IsDir() {
			subEntries, err := fs.ReadDir(embeddedFS, embeddedRoot+"/"+entry.Name())
			if err == nil && len(subEntries) > 0 {
				hasContent = true
				break
			}
		}
	}

	if !hasContent {
		return nil, errors.New("no embedded frontend assets found (embedded/ directory is empty)")
	}

	// Return the embedded subdirectory
	subFS, err := fs.Sub(embeddedFS, embeddedRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedded subdirectory: %w", err)
	}

	return subFS, nil
}
