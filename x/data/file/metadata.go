package file

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	storefile "github.com/spcent/plumego/store/file"
)

// ErrNilMetadataDB is returned when DBMetadataManager has no database handle.
var ErrNilMetadataDB = errors.New("file metadata: database cannot be nil")

const metadataSelectColumns = `
		id, tenant_id, name, path, size, mime_type, extension, hash,
		width, height, thumbnail_path, storage_type, metadata,
		uploaded_by, created_at, updated_at, last_access_at, deleted_at
`

// DBMetadataManager implements MetadataManager using a PostgreSQL database.
type DBMetadataManager struct {
	db  *sql.DB
	now func() time.Time
}

// DBMetadataOption configures DBMetadataManager.
type DBMetadataOption func(*DBMetadataManager)

// WithMetadataClock configures the clock used for mutation timestamps.
func WithMetadataClock(now func() time.Time) DBMetadataOption {
	return func(m *DBMetadataManager) {
		if now != nil {
			m.now = now
		}
	}
}

// NewDBMetadataManager creates a new database-backed metadata manager.
func NewDBMetadataManager(db *sql.DB, opts ...DBMetadataOption) MetadataManager {
	m := newDBMetadataManager(db, opts...)
	return m
}

// NewDBMetadataManagerE creates a database-backed metadata manager and reports
// invalid dependencies without relying on later method calls.
func NewDBMetadataManagerE(db *sql.DB, opts ...DBMetadataOption) (*DBMetadataManager, error) {
	if db == nil {
		return nil, ErrNilMetadataDB
	}
	return newDBMetadataManager(db, opts...), nil
}

func newDBMetadataManager(db *sql.DB, opts ...DBMetadataOption) *DBMetadataManager {
	m := &DBMetadataManager{db: db, now: time.Now}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *DBMetadataManager) requireDB() (*sql.DB, error) {
	if m == nil || m.db == nil {
		return nil, ErrNilMetadataDB
	}
	return m.db, nil
}

// Save stores file metadata in the database.
func (m *DBMetadataManager) Save(ctx context.Context, file *File) error {
	db, err := m.requireDB()
	if err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(file.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO files
		(id, tenant_id, name, path, size, mime_type, extension, hash, width, height,
		 thumbnail_path, storage_type, metadata, uploaded_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			size = EXCLUDED.size,
			updated_at = EXCLUDED.updated_at
	`

	_, err = db.ExecContext(ctx, query,
		file.ID, file.TenantID, file.Name, file.Path, file.Size,
		file.MimeType, file.Extension, file.Hash, file.Width, file.Height,
		file.ThumbnailPath, file.StorageType, metadataJSON,
		file.UploadedBy, file.CreatedAt, file.UpdatedAt,
	)

	return err
}

// Get retrieves file metadata by ID.
func (m *DBMetadataManager) Get(ctx context.Context, id string) (*File, error) {
	db, err := m.requireDB()
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + metadataSelectColumns + `
		FROM files
		WHERE id = $1 AND deleted_at IS NULL
	`

	file, err := scanMetadataFile(db.QueryRowContext(ctx, query, id).Scan, "id "+id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storefile.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return file, nil
}

// GetByPath retrieves file metadata by path.
func (m *DBMetadataManager) GetByPath(ctx context.Context, p string) (*File, error) {
	db, err := m.requireDB()
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + metadataSelectColumns + `
		FROM files
		WHERE path = $1 AND deleted_at IS NULL
	`

	file, err := scanMetadataFile(db.QueryRowContext(ctx, query, p).Scan, "path "+p)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storefile.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return file, nil
}

// GetByHash retrieves file metadata by hash for deduplication.
func (m *DBMetadataManager) GetByHash(ctx context.Context, hash string) (*File, error) {
	db, err := m.requireDB()
	if err != nil {
		return nil, err
	}

	query := `SELECT ` + metadataSelectColumns + `
		FROM files
		WHERE hash = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	file, err := scanMetadataFile(db.QueryRowContext(ctx, query, hash).Scan, "hash "+hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // not found, not an error
	}
	if err != nil {
		return nil, err
	}

	return file, nil
}

// List retrieves file metadata matching the query.
func (m *DBMetadataManager) List(ctx context.Context, query Query) ([]*File, int64, error) {
	db, err := m.requireDB()
	if err != nil {
		return nil, 0, err
	}

	conditions := []string{"deleted_at IS NULL"}
	args := []any{}
	argIndex := 1

	if query.TenantID != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, query.TenantID)
		argIndex++
	}

	if query.UploadedBy != "" {
		conditions = append(conditions, fmt.Sprintf("uploaded_by = $%d", argIndex))
		args = append(args, query.UploadedBy)
		argIndex++
	}

	if query.MimeType != "" {
		conditions = append(conditions, fmt.Sprintf("mime_type = $%d", argIndex))
		args = append(args, query.MimeType)
		argIndex++
	}

	if !query.StartTime.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, query.StartTime)
		argIndex++
	}

	if !query.EndTime.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, query.EndTime)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	var total int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM files "+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	orderBy := "created_at DESC"
	switch query.OrderBy {
	case "created_at":
		orderBy = "created_at ASC"
	case "created_at_desc":
		orderBy = "created_at DESC"
	case "name":
		orderBy = "name ASC"
	case "name_desc":
		orderBy = "name DESC"
	case "size":
		orderBy = "size ASC"
	case "size_desc":
		orderBy = "size DESC"
	}

	listQuery := fmt.Sprintf(`
		SELECT %s
		FROM files
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, metadataSelectColumns, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)

	rows, err := db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []*File
	for rows.Next() {
		file, err := scanMetadataFile(rows.Scan, "list row")
		if err != nil {
			return nil, 0, err
		}
		results = append(results, file)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// Delete soft-deletes file metadata.
func (m *DBMetadataManager) Delete(ctx context.Context, id string) error {
	db, err := m.requireDB()
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx,
		`UPDATE files SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`,
		m.now(), id,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return storefile.ErrNotFound
	}

	return nil
}

// UpdateAccessTime updates the last access timestamp.
func (m *DBMetadataManager) UpdateAccessTime(ctx context.Context, id string) error {
	db, err := m.requireDB()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx,
		`UPDATE files SET last_access_at = $1 WHERE id = $2`,
		m.now(), id,
	)
	return err
}

func scanMetadataFile(scan func(dest ...any) error, contextLabel string) (*File, error) {
	var file File
	var metadataJSON []byte

	if err := scan(
		&file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
		&file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
		&file.ThumbnailPath, &file.StorageType, &metadataJSON,
		&file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
		&file.LastAccessAt, &file.DeletedAt,
	); err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &file.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata for %s: %w", contextLabel, err)
		}
	}

	return &file, nil
}
