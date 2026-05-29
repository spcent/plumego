package tag

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud-vault/internal/database"
)

// Repository is the persistence contract for tags.
type Repository interface {
	Create(ctx context.Context, t *Tag) error
	GetByID(ctx context.Context, id string) (*Tag, error)
	List(ctx context.Context) ([]*Tag, error)
	Update(ctx context.Context, t *Tag) error
	Delete(ctx context.Context, id string) error
	GetDocumentTags(ctx context.Context, docID string) ([]*Tag, error)
	SetDocumentTags(ctx context.Context, docID string, tagIDs []string) error
	RemoveDocumentTag(ctx context.Context, docID, tagID string) error
	EnsureTag(ctx context.Context, name, color, source string) (*Tag, error)
}

// SQLiteRepository implements Repository using SQLite.
type SQLiteRepository struct {
	db *database.DB
}

func NewSQLiteRepository(db *database.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, t *Tag) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tags (id, name, color, source, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		t.ID,
		t.Name,
		nullableStr(t.Color),
		nullableStr(t.Source),
		t.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicate
		}
		return fmt.Errorf("create tag: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id string) (*Tag, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, color, source, created_at FROM tags WHERE id = ?`, id)
	return scanTag(row)
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*Tag, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, color, source, created_at FROM tags ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		t, err := scanTagRow(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *SQLiteRepository) Update(ctx context.Context, t *Tag) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tags SET name = ?, color = ? WHERE id = ?`,
		t.Name, nullableStr(t.Color), t.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicate
		}
		return fmt.Errorf("update tag: %w", err)
	}
	return checkRowsAffected(result)
}

func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	// Delete associations first (foreign key ON DELETE CASCADE not set in schema).
	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM document_tags WHERE tag_id = ?`, id); err != nil {
		return fmt.Errorf("remove document_tags for tag: %w", err)
	}
	result, err := r.db.ExecContext(ctx, `DELETE FROM tags WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	return checkRowsAffected(result)
}

func (r *SQLiteRepository) GetDocumentTags(ctx context.Context, docID string) ([]*Tag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.name, t.color, t.source, t.created_at
		FROM tags t
		INNER JOIN document_tags dt ON dt.tag_id = t.id
		WHERE dt.document_id = ?
		ORDER BY t.name ASC`, docID)
	if err != nil {
		return nil, fmt.Errorf("get document tags: %w", err)
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		t, err := scanTagRow(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *SQLiteRepository) SetDocumentTags(ctx context.Context, docID string, tagIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM document_tags WHERE document_id = ?`, docID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("clear document tags: %w", err)
	}

	for _, tagID := range tagIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO document_tags (document_id, tag_id) VALUES (?, ?)`,
			docID, tagID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert document tag: %w", err)
		}
	}

	return tx.Commit()
}

func (r *SQLiteRepository) RemoveDocumentTag(ctx context.Context, docID, tagID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM document_tags WHERE document_id = ? AND tag_id = ?`, docID, tagID)
	return err
}

// EnsureTag returns an existing tag by name or creates a new one.
func (r *SQLiteRepository) EnsureTag(ctx context.Context, name, color, source string) (*Tag, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, color, source, created_at FROM tags WHERE name = ?`, name)
	t, err := scanTag(row)
	if err == nil {
		return t, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	// Create it.
	t = &Tag{
		ID:        newID(),
		Name:      name,
		Color:     color,
		Source:    source,
		CreatedAt: time.Now().UTC(),
	}
	if err := r.Create(ctx, t); err != nil && !errors.Is(err, ErrDuplicate) {
		return nil, err
	}
	// Re-fetch in case of race.
	row = r.db.QueryRowContext(ctx,
		`SELECT id, name, color, source, created_at FROM tags WHERE name = ?`, name)
	return scanTag(row)
}

// --- scan helpers ---

func scanTag(row *sql.Row) (*Tag, error) {
	var t Tag
	var createdAt string
	var color, source *string
	err := row.Scan(&t.ID, &t.Name, &color, &source, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if color != nil {
		t.Color = *color
	}
	if source != nil {
		t.Source = *source
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &t, nil
}

func scanTagRow(rows *sql.Rows) (*Tag, error) {
	var t Tag
	var createdAt string
	var color, source *string
	err := rows.Scan(&t.ID, &t.Name, &color, &source, &createdAt)
	if err != nil {
		return nil, err
	}
	if color != nil {
		t.Color = *color
	}
	if source != nil {
		t.Source = *source
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &t, nil
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

func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
