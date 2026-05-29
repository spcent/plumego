package collection

import (
	"context"
	"database/sql"
	"time"

	"cloud-vault/internal/database"
)

// Repository handles collection database operations.
type Repository struct {
	db *database.DB
}

// NewRepository constructs a Repository.
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c *Collection) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
INSERT INTO collections (id, name, description, type, status, created_at, updated_at)
VALUES (?,?,?,?,?,?,?)
`,
		c.ID, c.Name, nilStr(c.Description), c.Type, c.Status, now, now,
	)
	return err
}

func (r *Repository) Update(ctx context.Context, id, name, description string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		`UPDATE collections SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		name, nilStr(description), now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM collection_documents WHERE collection_id = ?`, id)
	if err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM collections WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Collection, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT c.id, c.name, COALESCE(c.description,''), c.type, c.status,
       COUNT(cd.document_id) AS doc_count, c.created_at, c.updated_at
FROM collections c
LEFT JOIN collection_documents cd ON cd.collection_id = c.id
WHERE c.id = ?
GROUP BY c.id
`, id)
	return scanCollection(row)
}

func (r *Repository) List(ctx context.Context) ([]Collection, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT c.id, c.name, COALESCE(c.description,''), c.type, c.status,
       COUNT(cd.document_id) AS doc_count, c.created_at, c.updated_at
FROM collections c
LEFT JOIN collection_documents cd ON cd.collection_id = c.id
WHERE c.status = 'active'
GROUP BY c.id
ORDER BY c.updated_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Collection
	for rows.Next() {
		var c Collection
		var createdAt, updatedAt string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Type, &c.Status,
			&c.DocCount, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		c.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		result = append(result, c)
	}
	return result, rows.Err()
}

func scanCollection(row *sql.Row) (*Collection, error) {
	var c Collection
	var createdAt, updatedAt string
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.Type, &c.Status,
		&c.DocCount, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &c, nil
}

// --- Collection documents ---

func (r *Repository) AddDocument(ctx context.Context, collectionID, docID, note string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	// Get current max sort_order.
	var maxOrder int
	_ = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sort_order),0) FROM collection_documents WHERE collection_id = ?`,
		collectionID).Scan(&maxOrder)

	_, err := r.db.ExecContext(ctx, `
INSERT INTO collection_documents (collection_id, document_id, sort_order, note, created_at)
VALUES (?,?,?,?,?)
`,
		collectionID, docID, maxOrder+1, nilStr(note), now,
	)
	if err != nil {
		// Duplicate is OK: treat as no-op.
		return nil
	}
	return r.touchCollection(ctx, collectionID)
}

func (r *Repository) RemoveDocument(ctx context.Context, collectionID, docID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM collection_documents WHERE collection_id = ? AND document_id = ?`,
		collectionID, docID)
	if err != nil {
		return err
	}
	return r.touchCollection(ctx, collectionID)
}

func (r *Repository) ListDocuments(ctx context.Context, collectionID string) ([]CollectionDocument, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT cd.collection_id, cd.document_id, d.title, d.status,
       cd.sort_order, COALESCE(cd.note,''), cd.created_at
FROM collection_documents cd
JOIN documents d ON d.id = cd.document_id
WHERE cd.collection_id = ? AND d.status != 'deleted'
ORDER BY cd.sort_order, cd.created_at
`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []CollectionDocument
	for rows.Next() {
		var cd CollectionDocument
		var createdAt string
		if err := rows.Scan(&cd.CollectionID, &cd.DocumentID, &cd.Title, &cd.Status,
			&cd.SortOrder, &cd.Note, &createdAt); err != nil {
			return nil, err
		}
		cd.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		result = append(result, cd)
	}
	return result, rows.Err()
}

func (r *Repository) Reorder(ctx context.Context, collectionID string, orderedDocIDs []string) error {
	for i, docID := range orderedDocIDs {
		_, err := r.db.ExecContext(ctx,
			`UPDATE collection_documents SET sort_order = ? WHERE collection_id = ? AND document_id = ?`,
			i+1, collectionID, docID)
		if err != nil {
			return err
		}
	}
	return r.touchCollection(ctx, collectionID)
}

func (r *Repository) touchCollection(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `UPDATE collections SET updated_at = ? WHERE id = ?`, now, id)
	return err
}

func nilStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
