package document

import "time"

// CreateRequest is the payload for creating a new document via the editor.
type CreateRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// CreateFromImportRequest is used by the importer to create a document
// with additional provenance fields.
type CreateFromImportRequest struct {
	Title        string
	Content      string
	OriginalPath string
	ImportJobID  string
}

// UpdateRequest is the payload for updating an existing document.
type UpdateRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	BaseVersion int    `json:"base_version"`
}

// FavoriteRequest is the payload for toggling a document's favourite state.
type FavoriteRequest struct {
	IsFavorite bool `json:"is_favorite"`
}

// StatusRequest is the payload for archiving / restoring a document.
type StatusRequest struct {
	Status string `json:"status"` // active | archived
}

// ReviewStatusRequest is the payload for updating the review status.
type ReviewStatusRequest struct {
	ReviewStatus string `json:"review_status"` // pending | reviewed
}

// BatchStatusRequest updates the status of multiple documents at once.
type BatchStatusRequest struct {
	IDs    []string `json:"ids"`
	Status string   `json:"status"` // active | archived | deleted
}

// ListQuery holds query parameters for listing documents.
type ListQuery struct {
	Q            string
	TagID        string
	Status       string // active | archived | all; "" defaults to active
	SourceType   string // manual | imported; "" = no filter
	ImportJobID  string
	IsFavorite   *bool
	ReviewStatus string // pending | reviewed; "" = no filter
	SortBy       string // updated_at | created_at | title | size_bytes
	Order        string // asc | desc
	Limit        int
	Offset       int
	AfterID      string // keyset cursor: ID of the last item on the previous page
}

// Summary is the list-view representation of a document (no content).
type Summary struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Version      int       `json:"version"`
	SizeBytes    int64     `json:"size_bytes"`
	WordCount    int       `json:"word_count"`
	LineCount    int       `json:"line_count"`
	IsFavorite   bool      `json:"is_favorite"`
	SourceType   string    `json:"source_type"`
	ReviewStatus string    `json:"review_status"`
	Summary      string    `json:"summary,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Detail is the full document representation including Markdown content.
type Detail struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Version      int        `json:"version"`
	SizeBytes    int64      `json:"size_bytes"`
	WordCount    int        `json:"word_count"`
	LineCount    int        `json:"line_count"`
	IsFavorite   bool       `json:"is_favorite"`
	SourceType   string     `json:"source_type"`
	ImportJobID  string     `json:"import_job_id,omitempty"`
	OriginalPath string     `json:"original_path,omitempty"`
	Summary      string     `json:"summary,omitempty"`
	HeadingText  string     `json:"heading_text,omitempty"`
	ReviewStatus string     `json:"review_status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ImportedAt   *time.Time `json:"imported_at,omitempty"`
}

// SaveResult is returned after a create or update operation.
type SaveResult struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Version   int       `json:"version"`
	Changed   bool      `json:"changed"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListResult is the paginated list response.
type ListResult struct {
	Items      []Summary `json:"items"`
	Total      int64     `json:"total"` // -1 in keyset-cursor mode (total not computed)
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`                // only meaningful in offset mode
	NextCursor string    `json:"next_cursor,omitempty"` // pass as after_id for the next page
	HasMore    bool      `json:"has_more"`
}

// VersionSummary is a list-view item for a document version.
type VersionSummary struct {
	Version   int       `json:"version"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
	Kind      string    `json:"kind"`
	Pinned    bool      `json:"pinned"`
	Note      string    `json:"note,omitempty"`
}

// VersionListResult is the list of versions for a document.
type VersionListResult struct {
	Items []VersionSummary `json:"items"`
}

// VersionDetail is the full content for a specific document version.
type VersionDetail struct {
	ID        string    `json:"id"`
	Version   int       `json:"version"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Kind      string    `json:"kind"`
	Pinned    bool      `json:"pinned"`
	Note      string    `json:"note,omitempty"`
}

// CreateSnapshotRequest creates or pins a key version for the current content.
type CreateSnapshotRequest struct {
	Note string `json:"note"`
}

func toSummary(d *Document) Summary {
	return Summary{
		ID:           d.ID,
		Title:        d.Title,
		Version:      d.CurrentVersion,
		SizeBytes:    d.SizeBytes,
		WordCount:    d.WordCount,
		LineCount:    d.LineCount,
		IsFavorite:   d.IsFavorite,
		SourceType:   d.SourceType,
		ReviewStatus: d.ReviewStatus,
		Summary:      d.Summary,
		UpdatedAt:    d.UpdatedAt,
	}
}

func toDetail(d *Document, content string) Detail {
	return Detail{
		ID:           d.ID,
		Title:        d.Title,
		Content:      content,
		Version:      d.CurrentVersion,
		SizeBytes:    d.SizeBytes,
		WordCount:    d.WordCount,
		LineCount:    d.LineCount,
		IsFavorite:   d.IsFavorite,
		SourceType:   d.SourceType,
		ImportJobID:  d.ImportJobID,
		OriginalPath: d.OriginalPath,
		Summary:      d.Summary,
		HeadingText:  d.HeadingText,
		ReviewStatus: d.ReviewStatus,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		ImportedAt:   d.ImportedAt,
	}
}
