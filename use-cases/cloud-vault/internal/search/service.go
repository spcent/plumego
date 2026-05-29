package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud-vault/internal/config"
	"cloud-vault/internal/storage"
)

// IndexEvent carries the information needed to update the search index after a
// document is created, updated, or deleted. It is defined here so the document
// package can reference it without importing search (app.go wires the hook).
type IndexEvent struct {
	DocID    string
	Content  string
	Version  int
	Hash     string
	Deleted  bool
	IsImport bool
}

// Service orchestrates search queries, index maintenance, and history.
type Service struct {
	engine SearchEngine
	repo   *Repository
	store  storage.ObjectStorage
	cfg    config.SearchConfig
}

// NewService constructs a Service.
func NewService(engine SearchEngine, repo *Repository, store storage.ObjectStorage, cfg config.SearchConfig) *Service {
	return &Service{engine: engine, repo: repo, store: store, cfg: cfg}
}

// HandleIndexEvent is called by the document service hook after create/update/delete.
// It never returns an error to the caller; failures are recorded in document_index_status.
func (s *Service) HandleIndexEvent(ctx context.Context, event IndexEvent) {
	if !s.cfg.Enabled {
		return
	}
	if event.Deleted {
		s.engine.DeleteDocument(ctx, event.DocID)
		s.repo.DeleteIndexStatus(ctx, event.DocID)
		return
	}

	// Decide whether to index inline or defer to background indexer.
	indexInline := s.cfg.IndexOnSave && !event.IsImport
	if event.IsImport {
		indexInline = s.cfg.IndexOnImport == "inline"
	}

	if !indexInline {
		s.repo.UpsertIndexStatus(ctx, event.DocID, event.Hash, event.Version, IndexStatusPending, "")
		return
	}

	sdoc := SearchDocument{
		DocumentID:  event.DocID,
		Content:     Clean(event.Content),
		ContentHash: event.Hash,
		Version:     event.Version,
	}

	if err := s.engine.IndexDocument(ctx, sdoc); err != nil {
		s.repo.UpsertIndexStatus(ctx, event.DocID, event.Hash, event.Version, IndexStatusFailed, err.Error())
		return
	}
	s.repo.UpsertIndexStatus(ctx, event.DocID, event.Hash, event.Version, IndexStatusIndexed, "")
}

// Search executes a query, saves history for non-empty queries, and enriches
// results with tag data.
func (s *Service) Search(ctx context.Context, q SearchQuery) (*SearchResult, error) {
	if !s.cfg.Enabled {
		return &SearchResult{Items: []SearchResultItem{}, Limit: q.Limit, Offset: q.Offset}, nil
	}

	result, err := s.engine.Search(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	// Enrich with tags.
	if len(result.Items) > 0 {
		ids := make([]string, len(result.Items))
		for i, item := range result.Items {
			ids[i] = item.ID
		}
		tagMap, err := s.repo.GetTagsForDocuments(ctx, ids)
		if err == nil {
			for i := range result.Items {
				if tags, ok := tagMap[result.Items[i].ID]; ok {
					result.Items[i].Tags = tags
				}
			}
		}
	}

	// Save history for non-empty queries.
	query := strings.TrimSpace(q.Q)
	if query != "" {
		rec := SearchHistoryRecord{
			ID:          newID(),
			Query:       query,
			ResultCount: int(result.Total),
			CreatedAt:   time.Now().UTC(),
		}
		_ = s.repo.SaveHistory(ctx, rec)
	}

	return result, nil
}

// GetIndexStatus returns aggregated indexing statistics.
func (s *Service) GetIndexStatus(ctx context.Context) (*IndexStatusResponse, error) {
	summary, err := s.repo.GetIndexStatusSummary(ctx)
	if err != nil {
		return nil, err
	}
	resp := &IndexStatusResponse{
		TotalDocuments: summary.TotalDocuments,
		Indexed:        summary.Indexed,
		Pending:        summary.Pending,
		Failed:         summary.Failed,
		Stale:          summary.Stale,
	}
	if summary.LastIndexedAt != nil {
		resp.LastIndexedAt = summary.LastIndexedAt.UTC().Format(time.RFC3339)
	}
	return resp, nil
}

// Reindex marks documents for re-indexing based on scope.
func (s *Service) Reindex(ctx context.Context, req ReindexRequest) error {
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "all"
	}
	if scope == "document" && req.DocumentID == "" {
		return fmt.Errorf("document_id is required for scope=document")
	}
	return s.repo.MarkStatusByScope(ctx, scope, req.DocumentID)
}

// GetHistory returns recent search history entries.
func (s *Service) GetHistory(ctx context.Context) (*HistoryResponse, error) {
	limit := s.cfg.HistoryLimit
	if limit <= 0 {
		limit = 100
	}
	records, err := s.repo.GetHistory(ctx, limit)
	if err != nil {
		return nil, err
	}
	items := make([]HistoryItem, 0, len(records))
	for _, rec := range records {
		items = append(items, HistoryItem{
			ID:          rec.ID,
			Query:       rec.Query,
			ResultCount: rec.ResultCount,
			CreatedAt:   rec.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return &HistoryResponse{Items: items}, nil
}

// ClearHistory removes all search history.
func (s *Service) ClearHistory(ctx context.Context) error {
	return s.repo.ClearHistory(ctx)
}
