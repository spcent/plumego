package search

import (
	"context"
	"io"
	"time"

	plumelog "github.com/spcent/plumego/log"

	"cloud-vault/internal/config"
	"cloud-vault/internal/storage"
)

// Indexer is a background worker that periodically processes documents pending indexing.
type Indexer struct {
	engine SearchEngine
	repo   *Repository
	store  storage.ObjectStorage
	cfg    config.SearchConfig
	logger plumelog.StructuredLogger
}

// NewIndexer constructs an Indexer.
func NewIndexer(engine SearchEngine, repo *Repository, store storage.ObjectStorage, cfg config.SearchConfig, logger plumelog.StructuredLogger) *Indexer {
	return &Indexer{engine: engine, repo: repo, store: store, cfg: cfg, logger: logger}
}

// Run starts the periodic indexing loop and blocks until ctx is cancelled.
func (idx *Indexer) Run(ctx context.Context) {
	if !idx.cfg.Enabled {
		return
	}
	interval := time.Duration(idx.cfg.IndexIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			idx.runBatch(ctx)
		}
	}
}

func (idx *Indexer) runBatch(ctx context.Context) {
	batchSize := idx.cfg.IndexBatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	docs, err := idx.repo.GetPendingDocs(ctx, batchSize)
	if err != nil {
		if ctx.Err() == nil {
			idx.logger.Error("indexer: get pending docs", plumelog.Fields{"error": err.Error()})
		}
		return
	}
	if len(docs) == 0 {
		return
	}

	maxBytes := idx.cfg.MaxContentSizeMB * 1024 * 1024
	if maxBytes <= 0 {
		maxBytes = 5 * 1024 * 1024
	}

	for _, d := range docs {
		if ctx.Err() != nil {
			return
		}
		idx.indexOne(ctx, d, maxBytes)
	}
}

func (idx *Indexer) indexOne(ctx context.Context, d pendingIndexDoc, maxBytes int64) {
	// Skip oversized files: still index title/headings/summary but not full content.
	content := ""
	if d.SizeBytes <= maxBytes {
		rc, err := idx.store.Get(ctx, d.StorageKey)
		if err != nil {
			idx.repo.UpsertIndexStatus(ctx, d.DocumentID, d.ContentHash, d.Version,
				IndexStatusFailed, err.Error())
			return
		}
		raw, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			idx.repo.UpsertIndexStatus(ctx, d.DocumentID, d.ContentHash, d.Version,
				IndexStatusFailed, err.Error())
			return
		}
		content = Clean(string(raw))
	}

	sdoc := SearchDocument{
		DocumentID:   d.DocumentID,
		Title:        d.Title,
		OriginalPath: d.OriginalPath,
		Summary:      d.Summary,
		Headings:     d.HeadingText,
		Content:      content,
		ContentHash:  d.ContentHash,
		Version:      d.Version,
	}

	if err := idx.engine.IndexDocument(ctx, sdoc); err != nil {
		idx.repo.UpsertIndexStatus(ctx, d.DocumentID, d.ContentHash, d.Version,
			IndexStatusFailed, err.Error())
		return
	}
	idx.repo.UpsertIndexStatus(ctx, d.DocumentID, d.ContentHash, d.Version,
		IndexStatusIndexed, "")
}
