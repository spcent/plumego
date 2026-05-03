package file

import (
	"io"
	"time"
)

// File represents file metadata returned by store/file operations. Metadata is
// caller-owned unless a concrete implementation documents that it makes a
// defensive copy. Provider-specific fields may be projected into the common
// fields only when they are portable across backends; tenant-aware file
// metadata belongs in extension packages.
type File struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Path         string         `json:"path"`
	Size         int64          `json:"size"`
	MimeType     string         `json:"mime_type"`
	Extension    string         `json:"extension"`
	Hash         string         `json:"hash"`
	StorageType  string         `json:"storage_type"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	LastAccessAt *time.Time     `json:"last_access_at,omitempty"`
	DeletedAt    *time.Time     `json:"deleted_at,omitempty"`
}

// PutOptions contains options for uploading a file. Reader is consumed by the
// concrete implementation during Put. Metadata is caller-owned unless a
// concrete implementation documents that it makes a defensive copy. Tenant
// identity is not part of the stable store layer.
type PutOptions struct {
	Reader      io.Reader      // File content
	FileName    string         // Original filename
	ContentType string         // MIME type
	Size        int64          // File size in bytes; negative values mean unknown only if the backend documents it
	Metadata    map[string]any // Additional metadata
}

// FileStat contains basic file information.
type FileStat struct {
	Path         string    // File path
	Size         int64     // File size in bytes
	ModifiedTime time.Time // Last modified time
	IsDir        bool      // Whether this is a directory
	ContentType  string    // MIME type
}
