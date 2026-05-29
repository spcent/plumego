package search

import "time"

const (
	IndexStatusPending = "pending"
	IndexStatusIndexed = "indexed"
	IndexStatusFailed  = "failed"
	IndexStatusStale   = "stale"
	IndexStatusDeleted = "deleted"
)

// IndexStatus is the indexing state record for a single document.
type IndexStatus struct {
	DocumentID     string
	ContentHash    string
	IndexedVersion int
	Status         string
	ErrorMessage   string
	IndexedAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IndexStatusSummary holds aggregated indexing statistics.
type IndexStatusSummary struct {
	TotalDocuments int64
	Indexed        int64
	Pending        int64
	Failed         int64
	Stale          int64
	LastIndexedAt  *time.Time
}

// SearchDocument is the data written into the FTS index for one document.
type SearchDocument struct {
	DocumentID   string
	Title        string
	OriginalPath string
	Summary      string
	Headings     string // plain text of all headings
	Content      string // cleaned plain text
	ContentHash  string
	Version      int
}

// SearchHistoryRecord is one entry in the search history table.
type SearchHistoryRecord struct {
	ID          string
	Query       string
	FiltersJSON string
	ResultCount int
	CreatedAt   time.Time
}

// pendingIndexDoc is returned by the repository when scanning for docs to index.
type pendingIndexDoc struct {
	DocumentID  string
	Title       string
	OriginalPath string
	Summary     string
	HeadingText string
	StorageKey  string
	ContentHash string
	Version     int
	SizeBytes   int64
}
