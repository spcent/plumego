package file

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	storefile "github.com/spcent/plumego/store/file"
)

// DBMetadataManager implements MetadataManager using a PostgreSQL database.
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
		return nil, storefile.ErrNotFound
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
func (m *DBMetadataManager) GetByPath(ctx context.Context, p string) (*File, error) {
	query := `
		SELECT id, tenant_id, name, path, size, mime_type, extension, hash,
		       width, height, thumbnail_path, storage_type, metadata,
		       uploaded_by, created_at, updated_at, last_access_at, deleted_at
		FROM files
		WHERE path = $1 AND deleted_at IS NULL
	`

	var file File
	var metadataJSON []byte

	err := m.db.QueryRowContext(ctx, query, p).Scan(
		&file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
		&file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
		&file.ThumbnailPath, &file.StorageType, &metadataJSON,
		&file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
		&file.LastAccessAt, &file.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, storefile.ErrNotFound
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
		return nil, nil // not found, not an error
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
	if err := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM files "+whereClause, args...).Scan(&total); err != nil {
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

		if err := rows.Scan(
			&file.ID, &file.TenantID, &file.Name, &file.Path, &file.Size,
			&file.MimeType, &file.Extension, &file.Hash, &file.Width, &file.Height,
			&file.ThumbnailPath, &file.StorageType, &metadataJSON,
			&file.UploadedBy, &file.CreatedAt, &file.UpdatedAt,
			&file.LastAccessAt, &file.DeletedAt,
		); err != nil {
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
	result, err := m.db.ExecContext(ctx,
		`UPDATE files SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`,
		time.Now(), id,
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
	_, err := m.db.ExecContext(ctx,
		`UPDATE files SET last_access_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	return err
}
