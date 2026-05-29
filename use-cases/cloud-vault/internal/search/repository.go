package search

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud-vault/internal/database"
)

// Repository handles all non-FTS database operations for the search subsystem.
type Repository struct {
	db *database.DB
}

// NewRepository constructs a Repository.
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// UpsertIndexStatus inserts or updates the indexing state for a document.
func (r *Repository) UpsertIndexStatus(ctx context.Context, docID, hash string, version int, status, errMsg string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	var indexedAt *string
	if status == IndexStatusIndexed {
		s := now
		indexedAt = &s
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO document_index_status
		  (document_id, content_hash, indexed_version, status, error_message, indexed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(document_id) DO UPDATE SET
		  content_hash    = excluded.content_hash,
		  indexed_version = excluded.indexed_version,
		  status          = excluded.status,
		  error_message   = excluded.error_message,
		  indexed_at      = CASE WHEN excluded.status = 'indexed' THEN excluded.indexed_at ELSE indexed_at END,
		  updated_at      = excluded.updated_at`,
		docID, hash, version, status, nullableStr(errMsg), indexedAt, now, now,
	)
	return err
}

// DeleteIndexStatus removes the indexing state for a document.
func (r *Repository) DeleteIndexStatus(ctx context.Context, docID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM document_index_status WHERE document_id = ?`, docID)
	return err
}

// MarkStatusByScope marks documents for re-indexing based on a scope string.
func (r *Repository) MarkStatusByScope(ctx context.Context, scope, docID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	switch scope {
	case "all":
		_, err := r.db.ExecContext(ctx,
			`UPDATE document_index_status SET status = 'pending', updated_at = ? WHERE status != 'deleted'`, now)
		return err
	case "failed":
		_, err := r.db.ExecContext(ctx,
			`UPDATE document_index_status SET status = 'pending', updated_at = ? WHERE status = 'failed'`, now)
		return err
	case "stale":
		_, err := r.db.ExecContext(ctx,
			`UPDATE document_index_status SET status = 'pending', updated_at = ? WHERE status = 'stale'`, now)
		return err
	case "document":
		_, err := r.db.ExecContext(ctx,
			`UPDATE document_index_status SET status = 'pending', updated_at = ? WHERE document_id = ?`, now, docID)
		return err
	default:
		return fmt.Errorf("unknown reindex scope: %q", scope)
	}
}

// GetIndexStatusSummary returns aggregated indexing statistics.
func (r *Repository) GetIndexStatusSummary(ctx context.Context) (*IndexStatusSummary, error) {
	// Total non-deleted documents.
	var totalDocs int64
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM documents WHERE status != 'deleted'`,
	).Scan(&totalDocs); err != nil {
		return nil, fmt.Errorf("count documents: %w", err)
	}

	// Count by status in document_index_status.
	rows, err := r.db.QueryContext(ctx,
		`SELECT status, COUNT(*) FROM document_index_status GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("count index status: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var st string
		var n int64
		if err := rows.Scan(&st, &n); err != nil {
			return nil, err
		}
		counts[st] = n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Stale: indexed docs whose content_hash differs from documents.content_hash.
	var stale int64
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM document_index_status dis
		 JOIN documents d ON d.id = dis.document_id
		 WHERE dis.status = 'indexed' AND dis.content_hash != d.content_hash`,
	).Scan(&stale); err != nil {
		return nil, fmt.Errorf("count stale: %w", err)
	}

	// Most recent indexed_at.
	var lastIndexedAt *time.Time
	var lastStr *string
	if err := r.db.QueryRowContext(ctx,
		`SELECT MAX(indexed_at) FROM document_index_status WHERE indexed_at IS NOT NULL`,
	).Scan(&lastStr); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("last indexed_at: %w", err)
	}
	if lastStr != nil {
		if t, err := time.Parse(time.RFC3339, *lastStr); err == nil {
			lastIndexedAt = &t
		}
	}

	// Pending = untracked docs + explicitly pending.
	tracked := counts[IndexStatusIndexed] + counts[IndexStatusFailed] + counts[IndexStatusPending]
	untracked := totalDocs - tracked
	if untracked < 0 {
		untracked = 0
	}
	pending := counts[IndexStatusPending] + untracked

	return &IndexStatusSummary{
		TotalDocuments: totalDocs,
		Indexed:        counts[IndexStatusIndexed],
		Pending:        pending,
		Failed:         counts[IndexStatusFailed],
		Stale:          stale,
		LastIndexedAt:  lastIndexedAt,
	}, nil
}

// GetPendingDocs returns documents that need (re-)indexing.
// It includes: never indexed, pending, failed, and stale (hash changed).
func (r *Repository) GetPendingDocs(ctx context.Context, limit int) ([]pendingIndexDoc, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.title, d.original_path, d.summary, d.heading_text,
		       d.storage_key, d.content_hash, d.current_version, d.size_bytes
		FROM documents d
		LEFT JOIN document_index_status dis ON dis.document_id = d.id
		WHERE d.status != 'deleted'
		  AND (
		        dis.document_id IS NULL
		     OR dis.status = 'pending'
		     OR dis.status = 'failed'
		     OR (dis.status = 'indexed' AND dis.content_hash != d.content_hash)
		  )
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending docs: %w", err)
	}
	defer rows.Close()

	var docs []pendingIndexDoc
	for rows.Next() {
		var d pendingIndexDoc
		if err := rows.Scan(
			&d.DocumentID, &d.Title, &d.OriginalPath, &d.Summary, &d.HeadingText,
			&d.StorageKey, &d.ContentHash, &d.Version, &d.SizeBytes,
		); err != nil {
			return nil, fmt.Errorf("scan pending doc: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// GetTagsForDocuments returns tag names keyed by document ID for a set of IDs.
func (r *Repository) GetTagsForDocuments(ctx context.Context, docIDs []string) (map[string][]string, error) {
	if len(docIDs) == 0 {
		return map[string][]string{}, nil
	}
	ph := strings.Repeat("?,", len(docIDs))
	ph = ph[:len(ph)-1]
	args := make([]any, len(docIDs))
	for i, id := range docIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT dt.document_id, t.name
		 FROM document_tags dt
		 JOIN tags t ON t.id = dt.tag_id
		 WHERE dt.document_id IN (`+ph+`)
		 ORDER BY t.name`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("get tags for docs: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var docID, tagName string
		if err := rows.Scan(&docID, &tagName); err != nil {
			return nil, err
		}
		result[docID] = append(result[docID], tagName)
	}
	return result, rows.Err()
}

// SaveHistory appends a search history entry.
func (r *Repository) SaveHistory(ctx context.Context, rec SearchHistoryRecord) error {
	now := rec.CreatedAt.UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO search_history (id, query, filters_json, result_count, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		rec.ID, rec.Query, nullableStr(rec.FiltersJSON), rec.ResultCount, now,
	)
	return err
}

// GetHistory returns the most recent history entries.
func (r *Repository) GetHistory(ctx context.Context, limit int) ([]SearchHistoryRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, query, filters_json, result_count, created_at
		 FROM search_history
		 ORDER BY created_at DESC
		 LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	defer rows.Close()

	var records []SearchHistoryRecord
	for rows.Next() {
		var rec SearchHistoryRecord
		var createdAt string
		var filtersJSON *string
		if err := rows.Scan(&rec.ID, &rec.Query, &filtersJSON, &rec.ResultCount, &createdAt); err != nil {
			return nil, err
		}
		if filtersJSON != nil {
			rec.FiltersJSON = *filtersJSON
		}
		rec.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		records = append(records, rec)
	}
	return records, rows.Err()
}

// ClearHistory deletes all search history entries.
func (r *Repository) ClearHistory(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM search_history`)
	return err
}

func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
