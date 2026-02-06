package file

import (
	"io"
	"time"
)

// File represents file metadata stored in the database.
type File struct {
	ID            string                 `json:"id" db:"id"`
	TenantID      string                 `json:"tenant_id" db:"tenant_id"`
	Name          string                 `json:"name" db:"name"`
	Path          string                 `json:"path" db:"path"`
	Size          int64                  `json:"size" db:"size"`
	MimeType      string                 `json:"mime_type" db:"mime_type"`
	Extension     string                 `json:"extension" db:"extension"`
	Hash          string                 `json:"hash" db:"hash"`
	Width         int                    `json:"width,omitempty" db:"width"`
	Height        int                    `json:"height,omitempty" db:"height"`
	ThumbnailPath string                 `json:"thumbnail_path,omitempty" db:"thumbnail_path"`
	StorageType   string                 `json:"storage_type" db:"storage_type"`
	Metadata      map[string]any `json:"metadata,omitempty" db:"metadata"`
	UploadedBy    string                 `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
	LastAccessAt  *time.Time             `json:"last_access_at,omitempty" db:"last_access_at"`
	DeletedAt     *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PutOptions contains options for uploading a file.
type PutOptions struct {
	TenantID      string                 // Tenant ID for multi-tenancy
	Reader        io.Reader              // File content
	FileName      string                 // Original filename
	ContentType   string                 // MIME type
	Size          int64                  // File size (-1 if unknown)
	UploadedBy    string                 // User ID of uploader
	GenerateThumb bool                   // Whether to generate thumbnail
	ThumbWidth    int                    // Thumbnail width (default 200)
	ThumbHeight   int                    // Thumbnail height (default 200)
	Metadata      map[string]any // Additional metadata
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
type Query struct {
	TenantID   string
	UploadedBy string
	MimeType   string
	StartTime  time.Time
	EndTime    time.Time
	Page       int
	PageSize   int
	OrderBy    string
}

// ImageInfo contains image metadata.
type ImageInfo struct {
	Width  int    // Image width in pixels
	Height int    // Image height in pixels
	Format string // Image format (jpeg, png, gif)
}

// StorageConfig contains configuration for storage backends.
type StorageConfig struct {
	Type string // "local" or "s3"

	// Local storage config
	LocalBasePath string
	LocalBaseURL  string

	// S3 storage config
	S3Endpoint  string
	S3Region    string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3UseSSL    bool
	S3PathStyle bool // Path-style (true) or virtual-hosted-style (false)
}
