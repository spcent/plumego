package vectorstore

import (
	"context"
	"time"
)

// Backend defines the interface for vector storage backends.
type Backend interface {
	// Add adds a vector entry to the store.
	Add(ctx context.Context, entry *Entry) error

	// AddBatch adds multiple vector entries to the store.
	AddBatch(ctx context.Context, entries []*Entry) error

	// Search searches for similar vectors.
	Search(ctx context.Context, query []float64, topK int, threshold float64) ([]*SearchResult, error)

	// Get retrieves an entry by ID.
	Get(ctx context.Context, id string) (*Entry, error)

	// Delete removes an entry by ID.
	Delete(ctx context.Context, id string) error

	// DeleteBatch removes multiple entries by ID.
	DeleteBatch(ctx context.Context, ids []string) error

	// Update updates an entry.
	Update(ctx context.Context, entry *Entry) error

	// Count returns the number of entries in the store.
	Count(ctx context.Context) (int64, error)

	// Clear removes all entries from the store.
	Clear(ctx context.Context) error

	// Close closes the backend and releases resources.
	Close() error

	// Name returns the backend name.
	Name() string

	// Stats returns backend statistics.
	Stats(ctx context.Context) (*BackendStats, error)
}

// Entry represents a vector entry in the store.
type Entry struct {
	ID          string         `json:"id"`
	Vector      []float64      `json:"vector"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Text        string         `json:"text"`
	Response    string         `json:"response,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	AccessedAt  time.Time      `json:"accessed_at"`
	AccessCount int64          `json:"access_count"`
	TTL         time.Duration  `json:"ttl,omitempty"`
}

// SearchResult represents a search result with similarity score.
type SearchResult struct {
	Entry      *Entry  `json:"entry"`
	Similarity float64 `json:"similarity"`
	Distance   float64 `json:"distance"`
}

// BackendStats holds backend statistics.
type BackendStats struct {
	Name           string        `json:"name"`
	TotalEntries   int64         `json:"total_entries"`
	TotalVectors   int64         `json:"total_vectors"`
	Dimensions     int           `json:"dimensions"`
	MemoryUsage    int64         `json:"memory_usage_bytes"`
	AverageLatency time.Duration `json:"average_latency"`
	HitRate        float64       `json:"hit_rate"`
	LastCompaction time.Time     `json:"last_compaction,omitempty"`
}

// BackendType represents the type of vector store backend.
type BackendType string

const (
	BackendMemory   BackendType = "memory"
	BackendPinecone BackendType = "pinecone"
	BackendWeaviate BackendType = "weaviate"
	BackendQdrant   BackendType = "qdrant"
	BackendMilvus   BackendType = "milvus"
	BackendChroma   BackendType = "chroma"
	BackendRedis    BackendType = "redis"
)

// Config holds backend configuration.
type Config struct {
	Type       BackendType
	Endpoint   string
	APIKey     string
	Index      string
	Namespace  string
	Dimensions int
	Timeout    time.Duration
	MaxRetries int
	Options    map[string]any
}
