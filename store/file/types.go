package file

import (
	"io"
	"time"
)

// File represents file metadata. This is the pure, tenant-agnostic record
// returned by store/file operations. Callers that need tenant isolation should
// use the tenant-aware types in x/data/file. Metadata is caller-owned unless a
// concrete implementation documents that it makes a defensive copy.
type File struct {
	ID           string         `json:"id" db:"id"`
	Name         string         `json:"name" db:"name"`
	Path         string         `json:"path" db:"path"`
	Size         int64          `json:"size" db:"size"`
	MimeType     string         `json:"mime_type" db:"mime_type"`
	Extension    string         `json:"extension" db:"extension"`
	Hash         string         `json:"hash" db:"hash"`
	StorageType  string         `json:"storage_type" db:"storage_type"`
	Metadata     map[string]any `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	LastAccessAt *time.Time     `json:"last_access_at,omitempty" db:"last_access_at"`
	DeletedAt    *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PutOptions contains options for uploading a file.
// Tenant identity is not part of the stable store layer; use x/data/file.PutOptions
// when tenant-scoped uploads are required. Metadata is caller-owned unless a
// concrete implementation documents that it makes a defensive copy.
type PutOptions struct {
	Reader      io.Reader      // File content
	FileName    string         // Original filename
	ContentType string         // MIME type
	Size        int64          // File size in bytes; -1 may mean unknown
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

// Query contains parameters for querying file metadata.
// Tenant filtering is not part of the stable store layer; use x/data/file.Query
// when tenant-scoped queries are required. Zero values mean no stable-layer
// filter or ordering preference; concrete implementations own backend-specific
// validation.
type Query struct {
	MimeType  string
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int
	OrderBy   string
}
