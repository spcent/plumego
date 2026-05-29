package organize

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

// SimilarityType classifies how two documents relate.
type SimilarityType string

const (
	SimilarityTypeExact   SimilarityType = "exact_duplicate"
	SimilarityTypeNear    SimilarityType = "near_duplicate"
	SimilarityTypeRelated SimilarityType = "related"
)

const (
	SimilarityStatusPending   = "pending"
	SimilarityStatusIgnored   = "ignored"
	SimilarityStatusConfirmed = "confirmed"
)

const (
	JobStatusPending = "pending"
	JobStatusRunning = "running"
	JobStatusDone    = "done"
	JobStatusFailed  = "failed"
)

const (
	TagSuggestionStatusPending  = "pending"
	TagSuggestionStatusAccepted = "accepted"
	TagSuggestionStatusRejected = "rejected"
)

// DocumentSimilarity is a detected relationship between two documents.
type DocumentSimilarity struct {
	ID              string
	DocumentIDA     string
	DocumentIDB     string
	SimilarityType  SimilarityType
	SimilarityScore float64
	Reason          string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// OrganizeJob tracks a long-running organize operation.
type OrganizeJob struct {
	ID             string
	JobType        string
	Status         string
	TotalItems     int
	ProcessedItems int
	FailedItems    int
	ErrorMessage   string
	StartedAt      *time.Time
	FinishedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TagSuggestion is a proposed tag for a document (user must confirm).
type TagSuggestion struct {
	ID         string
	DocumentID string
	TagID      string
	TagName    string
	Source     string
	Confidence float64
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Topic is a rule-derived cluster of documents.
type Topic struct {
	ID          string
	Name        string
	Description string
	Source      string
	Status      string
	DocCount    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TopicDocument is a document assignment to a topic.
type TopicDocument struct {
	TopicID    string
	DocumentID string
	Score      float64
	Source     string
	CreatedAt  time.Time
}

// DocumentFingerprint stores lightweight text fingerprints.
type DocumentFingerprint struct {
	DocumentID  string
	TitleNorm   string
	Simhash     string
	HeadingHash string
	KeywordHash string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DuplicateGroup groups documents that share the same content hash.
type DuplicateGroup struct {
	ContentHash string
	Documents   []DuplicateDoc
}

// DuplicateDoc is a lightweight document summary inside a duplicate group.
type DuplicateDoc struct {
	ID           string
	Title        string
	Status       string
	IsFavorite   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ImportJobID  string
	OriginalPath string
}

// SimilarDoc is returned in the "similar documents" sidebar.
type SimilarDoc struct {
	SimilarityID   string
	DocumentID     string
	Title          string
	SimilarityType SimilarityType
	Score          float64
	Status         string
}

// ReviewItem is an entry in the organize review queue.
type ReviewItem struct {
	Type       string         `json:"type"` // duplicate | similar | tag_suggestion | prompt_candidate | low_quality
	DocumentID string         `json:"document_id"`
	Title      string         `json:"title"`
	Score      float64        `json:"score,omitempty"`
	Extra      map[string]any `json:"extra,omitempty"`
}
