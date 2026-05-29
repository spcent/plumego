package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud-vault/internal/database"
)

// FTSEngine implements SearchEngine using SQLite FTS5.
type FTSEngine struct {
	db            *database.DB
	snippetTokens int
}

// NewFTSEngine constructs an FTSEngine.
func NewFTSEngine(db *database.DB, snippetTokens int) *FTSEngine {
	if snippetTokens <= 0 {
		snippetTokens = 20
	}
	return &FTSEngine{db: db, snippetTokens: snippetTokens}
}

// IndexDocument upserts a document into the FTS index (delete + insert).
func (e *FTSEngine) IndexDocument(ctx context.Context, doc SearchDocument) error {
	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("fts begin: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM document_fts WHERE document_id = ?`, doc.DocumentID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("fts delete: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO document_fts(document_id, title, original_path, summary, headings, content)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		doc.DocumentID, doc.Title, doc.OriginalPath, doc.Summary, doc.Headings, doc.Content,
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("fts insert: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("fts commit: %w", err)
	}
	return nil
}

// DeleteDocument removes a document from the FTS index.
func (e *FTSEngine) DeleteDocument(ctx context.Context, documentID string) error {
	_, err := e.db.ExecContext(ctx,
		`DELETE FROM document_fts WHERE document_id = ?`, documentID)
	return err
}

// Search executes a full-text query and returns paginated results.
func (e *FTSEngine) Search(ctx context.Context, q SearchQuery) (*SearchResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// When q is empty fall back to a documents list ordered by updated_at.
	if strings.TrimSpace(q.Q) == "" {
		return e.listFallback(ctx, q, limit)
	}

	ftsQuery := sanitizeFTSQuery(q.Q)
	if ftsQuery == "" {
		return &SearchResult{Items: []SearchResultItem{}, Limit: limit, Offset: q.Offset}, nil
	}

	tokens := e.snippetTokens

	// Build the WHERE clause for the joined documents table.
	wheres, args := []string{"document_fts MATCH ?"}, []any{ftsQuery}
	wheres, args = appendDocFilters(wheres, args, q)

	where := strings.Join(wheres, " AND ")

	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int64
	if err := e.db.QueryRowContext(ctx,
		`SELECT COUNT(*)
		 FROM document_fts
		 JOIN documents d ON d.id = document_fts.document_id
		 WHERE `+where,
		countArgs...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("fts count: %w", err)
	}

	orderSQL := ftsOrderClause(q)
	args = append(args, limit, q.Offset)

	rows, err := e.db.QueryContext(ctx,
		fmt.Sprintf(`
		SELECT d.id, d.title, d.original_path, d.summary,
		       d.is_favorite, d.source_type, d.updated_at,
		       snippet(document_fts, 5, '<mark>', '</mark>', '...', %d) AS snippet_text,
		       bm25(document_fts) AS score
		FROM document_fts
		JOIN documents d ON d.id = document_fts.document_id
		WHERE `+where+`
		`+orderSQL+`
		LIMIT ? OFFSET ?`, tokens),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("fts search: %w", err)
	}
	defer rows.Close()

	items, err := scanSearchRows(rows)
	if err != nil {
		return nil, err
	}

	return &SearchResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: q.Offset,
	}, nil
}

// listFallback handles empty-query searches by reading from documents directly.
func (e *FTSEngine) listFallback(ctx context.Context, q SearchQuery, limit int) (*SearchResult, error) {
	wheres := []string{"d.status != 'deleted'"}
	args := []any{}

	if q.Status != "" && q.Status != "all" {
		wheres = append(wheres, "d.status = ?")
		args = append(args, q.Status)
	}
	if q.SourceType != "" {
		wheres = append(wheres, "d.source_type = ?")
		args = append(args, q.SourceType)
	}
	if q.IsFavorite != nil {
		if *q.IsFavorite {
			wheres = append(wheres, "d.is_favorite = 1")
		} else {
			wheres = append(wheres, "d.is_favorite = 0")
		}
	}
	if q.TagID != "" {
		wheres = append(wheres, "d.id IN (SELECT document_id FROM document_tags WHERE tag_id = ?)")
		args = append(args, q.TagID)
	}
	if q.From != "" {
		wheres = append(wheres, "d.updated_at >= ?")
		args = append(args, q.From)
	}
	if q.To != "" {
		wheres = append(wheres, "d.updated_at <= ?")
		args = append(args, q.To)
	}

	where := strings.Join(wheres, " AND ")
	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int64
	if err := e.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM documents d WHERE `+where, countArgs...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("fallback count: %w", err)
	}

	orderSQL := fallbackOrderClause(q)
	args = append(args, limit, q.Offset)

	rows, err := e.db.QueryContext(ctx,
		`SELECT d.id, d.title, d.original_path, d.summary,
		        d.is_favorite, d.source_type, d.updated_at,
		        '' AS snippet_text, 0.0 AS score
		 FROM documents d
		 WHERE `+where+` `+orderSQL+`
		 LIMIT ? OFFSET ?`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("fallback list: %w", err)
	}
	defer rows.Close()

	items, err := scanSearchRows(rows)
	if err != nil {
		return nil, err
	}

	return &SearchResult{Items: items, Total: total, Limit: limit, Offset: q.Offset}, nil
}

// --- helpers ---

// sanitizeFTSQuery converts free text into a safe FTS5 MATCH expression by
// wrapping each whitespace-delimited token in double quotes.
func sanitizeFTSQuery(q string) string {
	tokens := strings.Fields(q)
	if len(tokens) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(tokens))
	for _, t := range tokens {
		t = strings.ReplaceAll(t, `"`, `""`)
		quoted = append(quoted, `"`+t+`"`)
	}
	return strings.Join(quoted, " ")
}

func appendDocFilters(wheres []string, args []any, q SearchQuery) ([]string, []any) {
	// Status: default to excluding deleted.
	if q.Status == "all" {
		wheres = append(wheres, "d.status != 'deleted'")
	} else if q.Status != "" {
		wheres = append(wheres, "d.status = ?")
		args = append(args, q.Status)
	} else {
		wheres = append(wheres, "d.status = 'active'")
	}
	if q.SourceType != "" {
		wheres = append(wheres, "d.source_type = ?")
		args = append(args, q.SourceType)
	}
	if q.ReviewStatus != "" {
		wheres = append(wheres, "d.review_status = ?")
		args = append(args, q.ReviewStatus)
	}
	if q.ImportJobID != "" {
		wheres = append(wheres, "d.import_job_id = ?")
		args = append(args, q.ImportJobID)
	}
	if q.IsFavorite != nil {
		if *q.IsFavorite {
			wheres = append(wheres, "d.is_favorite = 1")
		} else {
			wheres = append(wheres, "d.is_favorite = 0")
		}
	}
	if q.TagID != "" {
		wheres = append(wheres, "d.id IN (SELECT document_id FROM document_tags WHERE tag_id = ?)")
		args = append(args, q.TagID)
	}
	if q.From != "" {
		wheres = append(wheres, "d.updated_at >= ?")
		args = append(args, q.From)
	}
	if q.To != "" {
		wheres = append(wheres, "d.updated_at <= ?")
		args = append(args, q.To)
	}
	return wheres, args
}

func ftsOrderClause(q SearchQuery) string {
	switch q.Sort {
	case "updated_at":
		if strings.ToUpper(q.Order) == "ASC" {
			return "ORDER BY d.updated_at ASC"
		}
		return "ORDER BY d.updated_at DESC"
	case "title":
		if strings.ToUpper(q.Order) == "ASC" {
			return "ORDER BY d.title ASC"
		}
		return "ORDER BY d.title DESC"
	default:
		return "ORDER BY bm25(document_fts)"
	}
}

func fallbackOrderClause(q SearchQuery) string {
	col := "d.updated_at"
	switch q.Sort {
	case "title":
		col = "d.title"
	case "created_at":
		col = "d.created_at"
	}
	dir := "DESC"
	if strings.ToUpper(q.Order) == "ASC" {
		dir = "ASC"
	}
	return fmt.Sprintf("ORDER BY %s %s", col, dir)
}

func scanSearchRows(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]SearchResultItem, error) {
	var items []SearchResultItem
	for rows.Next() {
		var item SearchResultItem
		var updatedAt string
		var snippet string
		var isFavorite int
		if err := rows.Scan(
			&item.ID, &item.Title, &item.OriginalPath, &item.Summary,
			&isFavorite, &item.SourceType, &updatedAt,
			&snippet, &item.Score,
		); err != nil {
			return nil, fmt.Errorf("scan search row: %w", err)
		}
		item.IsFavorite = isFavorite != 0
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			item.UpdatedAt = t.UTC().Format(time.RFC3339)
		} else {
			item.UpdatedAt = updatedAt
		}
		if snippet != "" {
			item.Highlights = []string{snippet}
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search rows: %w", err)
	}
	if items == nil {
		items = []SearchResultItem{}
	}
	return items, nil
}
