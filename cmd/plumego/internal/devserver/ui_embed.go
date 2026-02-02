package devserver

import (
	"embed"
	"io/fs"
)

//go:embed ui/*
var uiFS embed.FS

// GetUIFS returns the embedded UI filesystem
func GetUIFS() (fs.FS, error) {
	// Return a sub-filesystem starting from "ui" directory
	return fs.Sub(uiFS, "ui")
}

// HasEmbeddedUI returns true if UI files are embedded
func HasEmbeddedUI() bool {
	// Check if we can access the ui directory
	_, err := uiFS.ReadDir("ui")
	return err == nil
}
