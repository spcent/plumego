package search

import "context"

// SearchEngine is the persistence-layer contract for FTS operations.
type SearchEngine interface {
	IndexDocument(ctx context.Context, doc SearchDocument) error
	DeleteDocument(ctx context.Context, documentID string) error
	Search(ctx context.Context, q SearchQuery) (*SearchResult, error)
}
