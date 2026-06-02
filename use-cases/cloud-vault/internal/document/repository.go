package document

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud-vault/internal/database"
)

// Repository defines the persistence contract for documents.
type Repository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id string) (*Document, error)
	Update(ctx context.Context, doc *Document) error
	List(ctx context.Context, q ListQuery) ([]*Document, int64, error)
	SoftDelete(ctx context.Context, id string) error
	ExistsByHash(ctx context.Context, hash string) (bool, error)
	UpdateFavorite(ctx context.Context, id string, favorite bool) error
	UpdateStatus(ctx context.Context, id string, status string) error
	BatchUpdateStatus(ctx context.Context, ids []string, status string) error
	UpdateReviewStatus(ctx context.Context, id string, reviewStatus string) error
	CreateVersion(ctx context.Context, v *DocumentVersion) error
	GetVersions(ctx context.Context, docID string) ([]*DocumentVersion, error)
	GetVersion(ctx context.Context, docID string, version int) (*DocumentVersion, error)
}

// SQLiteRepository implements Repository using SQLite.
type SQLiteRepository struct {
	db *database.DB
}

// NewSQLiteRepository constructs a SQLiteRepository backed by db.
func NewSQLiteRepository(db *database.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// allDocCols is the column list shared by SELECT statements.
const allDocCols = `id, title, slug, original_path, storage_key, current_version, content_hash,
       size_bytes, word_count, line_count, status, sync_status, is_favorite,
       created_at, updated_at, uploaded_at,
       source_type, import_job_id, imported_at, summary, heading_text, review_status`

func (r *SQLiteRepository) Create(ctx context.Context, doc *Document) error {
	var uploadedAt, importedAt *string
	if doc.UploadedAt != nil {
		s := doc.UploadedAt.UTC().Format(time.RFC3339)
		uploadedAt = &s
	}
	if doc.ImportedAt != nil {
		s := doc.ImportedAt.UTC().Format(time.RFC3339)
		importedAt = &s
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO documents
		  (id, title, slug, original_path, storage_key, current_version, content_hash,
		   size_bytes, word_count, line_count, status, sync_status, is_favorite,
		   created_at, updated_at, uploaded_at,
		   source_type, import_job_id, imported_at, summary, heading_text, review_status)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		doc.ID, doc.Title, doc.Slug, doc.OriginalPath, doc.StorageKey,
		doc.CurrentVersion, doc.ContentHash, doc.SizeBytes, doc.WordCount, doc.LineCount,
		doc.Status, doc.SyncStatus, boolToInt(doc.IsFavorite),
		doc.CreatedAt.UTC().Format(time.RFC3339),
		doc.UpdatedAt.UTC().Format(time.RFC3339),
		uploadedAt,
		doc.SourceType, nullableStr(doc.ImportJobID), importedAt,
		nullableStr(doc.Summary), nullableStr(doc.HeadingText),
		doc.ReviewStatus,
	)
	if err != nil {
		return fmt.Errorf("create document: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*Document, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+allDocCols+`
		 FROM documents
		 WHERE id = ? AND status != 'deleted'`, id)
	return scanDocument(row)
}

func (r *SQLiteRepository) Update(ctx context.Context, doc *Document) error {
	var uploadedAt *string
	if doc.UploadedAt != nil {
		s := doc.UploadedAt.UTC().Format(time.RFC3339)
		uploadedAt = &s
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE documents
		SET title = ?, storage_key = ?, current_version = ?, content_hash = ?,
		    size_bytes = ?, word_count = ?, line_count = ?, sync_status = ?,
		    updated_at = ?, uploaded_at = ?,
		    summary = ?, heading_text = ?
		WHERE id = ?`,
		doc.Title, doc.StorageKey, doc.CurrentVersion, doc.ContentHash,
		doc.SizeBytes, doc.WordCount, doc.LineCount, doc.SyncStatus,
		doc.UpdatedAt.UTC().Format(time.RFC3339), uploadedAt,
		nullableStr(doc.Summary), nullableStr(doc.HeadingText),
		doc.ID,
	)
	if err != nil {
		return fmt.Errorf("update document: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) List(ctx context.Context, q ListQuery) ([]*Document, int64, error) {
	wheres, args := buildListWhere(q)
	where := "WHERE " + strings.Join(wheres, " AND ")

	countArgs := make([]any, len(args))
	copy(countArgs, args)

	var total int64
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM documents `+where,
		countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count documents: %w", err)
	}

	orderClause := buildOrderClause(q)
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	args = append(args, limit, q.Offset)

	rows, err := r.db.QueryContext(ctx,
		`SELECT `+allDocCols+`
		 FROM documents `+where+` `+orderClause+`
		 LIMIT ? OFFSET ?`,
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		doc, err := scanDocumentRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate documents: %w", err)
	}
	return docs, total, nil
}

func (r *SQLiteRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx,
		`UPDATE documents SET status = 'deleted', updated_at = ? WHERE id = ? AND status != 'deleted'`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("soft delete document: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("soft delete rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SQLiteRepository) ExistsByHash(ctx context.Context, hash string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM documents WHERE content_hash = ? AND status != 'deleted'`,
		hash,
	).Scan(&count)
	return count > 0, err
}

func (r *SQLiteRepository) UpdateFavorite(ctx context.Context, id string, favorite bool) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx,
		`UPDATE documents SET is_favorite = ?, updated_at = ? WHERE id = ? AND status != 'deleted'`,
		boolToInt(favorite), now, id,
	)
	if err != nil {
		return fmt.Errorf("update favorite: %w", err)
	}
	return checkRowsAffected(result)
}

func (r *SQLiteRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx,
		`UPDATE documents SET status = ?, updated_at = ? WHERE id = ? AND status != 'deleted'`,
		status, now, id,
	)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return checkRowsAffected(result)
}

func (r *SQLiteRepository) BatchUpdateStatus(ctx context.Context, ids []string, status string) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]

	args := []any{status, now}
	for _, id := range ids {
		args = append(args, id)
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE documents SET status = ?, updated_at = ?
		 WHERE id IN (`+placeholders+`) AND status != 'deleted'`,
		args...,
	)
	return err
}

func (r *SQLiteRepository) UpdateReviewStatus(ctx context.Context, id string, reviewStatus string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx,
		`UPDATE documents SET review_status = ?, updated_at = ? WHERE id = ? AND status != 'deleted'`,
		reviewStatus, now, id,
	)
	if err != nil {
		return fmt.Errorf("update review status: %w", err)
	}
	return checkRowsAffected(result)
}

func (r *SQLiteRepository) CreateVersion(ctx context.Context, v *DocumentVersion) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO document_versions
		  (id, document_id, version, storage_key, content_hash, size_bytes, created_at, note)
		VALUES (?,?,?,?,?,?,?,?)`,
		v.ID, v.DocumentID, v.Version, v.StorageKey, v.ContentHash, v.SizeBytes,
		v.CreatedAt.UTC().Format(time.RFC3339), v.Note,
	)
	if err != nil {
		return fmt.Errorf("create document version: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetVersions(ctx context.Context, docID string) ([]*DocumentVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, document_id, version, storage_key, content_hash, size_bytes, created_at, note
		FROM document_versions
		WHERE document_id = ?
		ORDER BY version DESC`, docID)
	if err != nil {
		return nil, fmt.Errorf("get versions: %w", err)
	}
	defer rows.Close()

	var versions []*DocumentVersion
	for rows.Next() {
		v, err := scanVersionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *SQLiteRepository) GetVersion(ctx context.Context, docID string, version int) (*DocumentVersion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, document_id, version, storage_key, content_hash, size_bytes, created_at, note
		FROM document_versions
		WHERE document_id = ? AND version = ?`, docID, version)

	v, err := scanVersionRowSingle(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get version: %w", err)
	}
	return v, nil
}

// --- query helpers ---

func buildListWhere(q ListQuery) ([]string, []any) {
	var wheres []string
	var args []any

	// Status filter.
	if q.Status == "all" {
		wheres = append(wheres, "status != 'deleted'")
	} else {
		status := q.Status
		if status == "" {
			status = StatusActive
		}
		wheres = append(wheres, "status = ?")
		args = append(args, status)
	}

	// Title search.
	if q.Q != "" {
		wheres = append(wheres, "title LIKE ?")
		args = append(args, "%"+q.Q+"%")
	}

	// Source type filter.
	if q.SourceType != "" {
		wheres = append(wheres, "source_type = ?")
		args = append(args, q.SourceType)
	}

	// Import job filter.
	if q.ImportJobID != "" {
		wheres = append(wheres, "import_job_id = ?")
		args = append(args, q.ImportJobID)
	}

	// Favorite filter.
	if q.IsFavorite != nil {
		if *q.IsFavorite {
			wheres = append(wheres, "is_favorite = 1")
		} else {
			wheres = append(wheres, "is_favorite = 0")
		}
	}

	// Review status filter.
	if q.ReviewStatus != "" {
		wheres = append(wheres, "review_status = ?")
		args = append(args, q.ReviewStatus)
	}

	// Tag filter — subquery avoids duplicate rows from JOIN.
	if q.TagID != "" {
		wheres = append(wheres, "id IN (SELECT document_id FROM document_tags WHERE tag_id = ?)")
		args = append(args, q.TagID)
	}

	return wheres, args
}

func buildOrderClause(q ListQuery) string {
	const (
		colUpdatedAt = "updated_at"
		colCreatedAt = "created_at"
		colTitle     = "title"
		colSizeBytes = "size_bytes"

		dirAsc  = "ASC"
		dirDesc = "DESC"
	)

	col := colUpdatedAt
	switch strings.ToLower(q.SortBy) {
	case colCreatedAt:
		col = colCreatedAt
	case colTitle:
		col = colTitle
	case colSizeBytes:
		col = colSizeBytes
	}

	dir := dirDesc
	if strings.EqualFold(q.Order, dirAsc) {
		dir = dirAsc
	}

	return "ORDER BY " + col + " " + dir
}

// --- scan helpers ---

func scanDocument(row *sql.Row) (*Document, error) {
	var doc Document
	var createdAt, updatedAt string
	var uploadedAt, importedAt *string
	var slug, originalPath, importJobID, summary, headingText *string
	var isFavorite int

	err := row.Scan(
		&doc.ID, &doc.Title, &slug, &originalPath, &doc.StorageKey,
		&doc.CurrentVersion, &doc.ContentHash, &doc.SizeBytes, &doc.WordCount, &doc.LineCount,
		&doc.Status, &doc.SyncStatus, &isFavorite,
		&createdAt, &updatedAt, &uploadedAt,
		&doc.SourceType, &importJobID, &importedAt, &summary, &headingText, &doc.ReviewStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	doc.IsFavorite = isFavorite != 0
	doc.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	doc.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	if uploadedAt != nil {
		t, _ := time.Parse(time.RFC3339, *uploadedAt)
		doc.UploadedAt = &t
	}
	if importedAt != nil {
		t, _ := time.Parse(time.RFC3339, *importedAt)
		doc.ImportedAt = &t
	}
	if slug != nil {
		doc.Slug = *slug
	}
	if originalPath != nil {
		doc.OriginalPath = *originalPath
	}
	if importJobID != nil {
		doc.ImportJobID = *importJobID
	}
	if summary != nil {
		doc.Summary = *summary
	}
	if headingText != nil {
		doc.HeadingText = *headingText
	}
	return &doc, nil
}

func scanDocumentRow(rows *sql.Rows) (*Document, error) {
	var doc Document
	var createdAt, updatedAt string
	var uploadedAt, importedAt *string
	var slug, originalPath, importJobID, summary, headingText *string
	var isFavorite int

	err := rows.Scan(
		&doc.ID, &doc.Title, &slug, &originalPath, &doc.StorageKey,
		&doc.CurrentVersion, &doc.ContentHash, &doc.SizeBytes, &doc.WordCount, &doc.LineCount,
		&doc.Status, &doc.SyncStatus, &isFavorite,
		&createdAt, &updatedAt, &uploadedAt,
		&doc.SourceType, &importJobID, &importedAt, &summary, &headingText, &doc.ReviewStatus,
	)
	if err != nil {
		return nil, err
	}

	doc.IsFavorite = isFavorite != 0
	if slug != nil {
		doc.Slug = *slug
	}
	if originalPath != nil {
		doc.OriginalPath = *originalPath
	}
	if importJobID != nil {
		doc.ImportJobID = *importJobID
	}
	if summary != nil {
		doc.Summary = *summary
	}
	if headingText != nil {
		doc.HeadingText = *headingText
	}
	doc.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	doc.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	if uploadedAt != nil {
		t, _ := time.Parse(time.RFC3339, *uploadedAt)
		doc.UploadedAt = &t
	}
	if importedAt != nil {
		t, _ := time.Parse(time.RFC3339, *importedAt)
		doc.ImportedAt = &t
	}
	return &doc, nil
}

func scanVersionRow(rows *sql.Rows) (*DocumentVersion, error) {
	var v DocumentVersion
	var createdAt string
	err := rows.Scan(
		&v.ID, &v.DocumentID, &v.Version, &v.StorageKey, &v.ContentHash, &v.SizeBytes,
		&createdAt, &v.Note,
	)
	if err != nil {
		return nil, err
	}
	v.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &v, nil
}

func scanVersionRowSingle(row *sql.Row) (*DocumentVersion, error) {
	var v DocumentVersion
	var createdAt string
	err := row.Scan(
		&v.ID, &v.DocumentID, &v.Version, &v.StorageKey, &v.ContentHash, &v.SizeBytes,
		&createdAt, &v.Note,
	)
	if err != nil {
		return nil, err
	}
	v.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &v, nil
}

// --- misc helpers ---

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func checkRowsAffected(result sql.Result) error {
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
