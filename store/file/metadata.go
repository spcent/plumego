package file

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DBMetadataManager implements MetadataManager using a SQL database.
type DBMetadataManager struct {
	db *sql.DB
}

// NewDBMetadataManager creates a new database-backed metadata manager.
func NewDBMetadataManager(db *sql.DB) MetadataManager {
	return &DBMetadataManager{db: db}
}

// Save stores file metadata in the database.
func (m *DBMetadataManager) Save(ctx context.Context, file *File) error {
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

	_, err = m.db.ExecContext(ctx, query,
		file.ID, file.TenantID, file.Name, file.Path, file.Size,
		file.MimeType, file.Extension, file.Hash, file.Width, file.Height,
		file.ThumbnailPath, file.StorageType, metadataJSON,
		file.UploadedBy, file.CreatedAt, file.UpdatedAt,
	)

	return err
}

// Get retrieves file metadata by ID.
func (m *DBMetadataManager) Get(ctx context.Context, id string) (*File, error) {
	query := `
		SELECT id, tenant_id, name, path, size, mime_type, extension, hash,
		       width, height, thumbnail_path, storage_type, metadata,
		       uploaded_by, created_at, updated_at, last_access_at, deleted_at
		FROM files
		WHERE id = $1 AND deleted_at IS NULL
	`

	var file File
	var metadataJSON []byte

	err := m.db.QueryRowContext(ctx, query, id).Scan(
		&file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
		&file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
		&file.ThumbnailPath, &file.StorageType, &metadataJSON,
		&file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
		&file.LastAccessAt, &file.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &file.Metadata)
	}

	return &file, nil
}

// GetByPath retrieves file metadata by path.
func (m *DBMetadataManager) GetByPath(ctx context.Context, path string) (*File, error) {
	query := `
		SELECT id, tenant_id, name, path, size, mime_type, extension, hash,
		       width, height, thumbnail_path, storage_type, metadata,
		       uploaded_by, created_at, updated_at, last_access_at, deleted_at
		FROM files
		WHERE path = $1 AND deleted_at IS NULL
	`

	var file File
	var metadataJSON []byte

	err := m.db.QueryRowContext(ctx, query, path).Scan(
		&file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
		&file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
		&file.ThumbnailPath, &file.StorageType, &metadataJSON,
		&file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
		&file.LastAccessAt, &file.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &file.Metadata)
	}

	return &file, nil
}

// GetByHash retrieves file metadata by hash for deduplication.
func (m *DBMetadataManager) GetByHash(ctx context.Context, hash string) (*File, error) {
	query := `
		SELECT id, tenant_id, name, path, size, mime_type, extension, hash,
		       width, height, thumbnail_path, storage_type, metadata,
		       uploaded_by, created_at, updated_at, last_access_at, deleted_at
		FROM files
		WHERE hash = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var file File
	var metadataJSON []byte

	err := m.db.QueryRowContext(ctx, query, hash).Scan(
		&file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
		&file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
		&file.ThumbnailPath, &file.StorageType, &metadataJSON,
		&file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
		&file.LastAccessAt, &file.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not an error, file doesn't exist
	}
	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &file.Metadata)
	}

	return &file, nil
}

// List retrieves file metadata matching the query.
func (m *DBMetadataManager) List(ctx context.Context, query Query) ([]*File, int64, error) {
	// Build WHERE conditions
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

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM files %s", whereClause)
	err := m.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Pagination
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	// Sanitize orderBy to prevent SQL injection by allowing only known columns/directions.
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
	// Add more allowed sort options here as needed, keeping the mapping explicit.
	}

	listQuery := fmt.Sprintf(`
		SELECT id, tenant_id, name, path, size, mime_type, extension, hash,
		       width, height, thumbnail_path, storage_type, metadata,
		       uploaded_by, created_at, updated_at, last_access_at, deleted_at
		FROM files
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)

	rows, err := m.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []*File
	for rows.Next() {
		var file File
		var metadataJSON []byte

		err := rows.Scan(
			&file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
			&file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
			&file.ThumbnailPath, &file.StorageType, &metadataJSON,
			&file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
			&file.LastAccessAt, &file.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &file.Metadata)
		}

		results = append(results, &file)
	}

	return results, total, nil
}

// Delete soft-deletes file metadata.
func (m *DBMetadataManager) Delete(ctx context.Context, id string) error {
	query := `UPDATE files SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := m.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdateAccessTime updates the last access timestamp.
func (m *DBMetadataManager) UpdateAccessTime(ctx context.Context, id string) error {
	query := `UPDATE files SET last_access_at = $1 WHERE id = $2`
	_, err := m.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
