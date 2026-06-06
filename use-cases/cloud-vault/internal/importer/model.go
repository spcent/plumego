package importer

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("import job not found")
var ErrInvalidSource = errors.New("invalid import source")

// Job status constants.
const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusPaused    = "paused"
	JobStatusDone      = "done"
	JobStatusFailed    = "failed"
	JobStatusCancelled = "cancelled"
)

// Item status constants.
const (
	ItemStatusPending = "pending"
	ItemStatusSuccess = "success"
	ItemStatusSkipped = "skipped"
	ItemStatusFailed  = "failed"
)

// ImportJob is the top-level import task for a directory.
type ImportJob struct {
	ID             string
	Name           string
	SourcePath     string
	Status         string
	TotalCount     int
	ProcessedCount int
	SuccessCount   int
	FailedCount    int
	SkippedCount   int
	ErrorMessage   string
	StartedAt      *time.Time
	CompletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ImportJobItem is a single file entry within an import job.
type ImportJobItem struct {
	ID           string
	JobID        string
	FilePath     string
	DocumentID   string
	Status       string
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DocumentMetadata stores rich extracted metadata for an imported document.
type DocumentMetadata struct {
	ID             string
	DocumentID     string
	Headings       string // JSON
	CodeLanguages  string // JSON
	CodeBlockCount int
	LinkCount      int
	ImageCount     int
	ExtractedAt    time.Time
}
