package document

import (
	"errors"
	"fmt"
	"time"
)

// Domain errors returned by the service layer.
var (
	ErrNotFound        = errors.New("document not found")
	ErrVersionConflict = errors.New("document version conflict")
)

const (
	StatusActive   = "active"
	StatusArchived = "archived"
	StatusDeleted  = "deleted"

	SyncStatusSynced  = "synced"
	SyncStatusPending = "pending"

	SourceTypeManual   = "manual"
	SourceTypeImported = "imported"

	ReviewStatusPending  = "pending"
	ReviewStatusReviewed = "reviewed"
)

// Document is the core document entity.
type Document struct {
	ID             string
	Title          string
	Slug           string
	OriginalPath   string
	StorageKey     string
	CurrentVersion int
	ContentHash    string
	SizeBytes      int64
	WordCount      int
	LineCount      int
	Status         string
	SyncStatus     string
	IsFavorite     bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	UploadedAt     *time.Time

	// V0.2 fields
	SourceType   string
	ImportJobID  string
	ImportedAt   *time.Time
	Summary      string
	HeadingText  string
	ReviewStatus string
}

// DocumentVersion is a historical snapshot of a document's content.
type DocumentVersion struct {
	ID          string
	DocumentID  string
	Version     int
	StorageKey  string
	ContentHash string
	SizeBytes   int64
	CreatedAt   time.Time
	Note        string
}

// CurrentKey returns the object storage key for the document's latest content.
func CurrentKey(docID string) string {
	return "docs/" + docID + "/current.md"
}

// VersionKey returns the object storage key for a specific version of a document.
func VersionKey(docID string, version int) string {
	return fmt.Sprintf("docs/%s/versions/%06d.md", docID, version)
}
