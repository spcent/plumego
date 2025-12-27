package frontend

import (
	"embed"
	"errors"
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

func embeddedSub() (fs.FS, error) {
	entries, err := fs.ReadDir(embeddedFS, embeddedRoot)
	if err != nil {
		return nil, errors.New("no embedded frontend assets")
	}

	for _, entry := range entries {
		if entry.Name() != ".keep" {
			return fs.Sub(embeddedFS, embeddedRoot)
		}
	}

	return nil, errors.New("no embedded frontend assets")
}
