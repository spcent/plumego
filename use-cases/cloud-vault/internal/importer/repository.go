package importer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cloud-vault/internal/database"
)

// Repository manages import_jobs, import_job_items, and document_metadata.
type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// --- import_jobs ---

func (r *Repository) CreateJob(ctx context.Context, j *ImportJob) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO import_jobs
		  (id, name, source_path, status, total_count, processed_count,
		   success_count, failed_count, skipped_count, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		j.ID, j.Name, j.SourcePath, j.Status,
		j.TotalCount, j.ProcessedCount, j.SuccessCount, j.FailedCount, j.SkippedCount,
		j.CreatedAt.UTC().Format(time.RFC3339),
		j.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("create import job: %w", err)
	}
	return nil
}

func (r *Repository) GetJob(ctx context.Context, id string) (*ImportJob, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, source_path, status,
		       total_count, processed_count, success_count, failed_count, skipped_count,
		       error_message, started_at, completed_at, created_at, updated_at
		FROM import_jobs WHERE id = ?`, id)
	return scanJob(row)
}

func (r *Repository) ListJobs(ctx context.Context, limit, offset int) ([]*ImportJob, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM import_jobs`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, source_path, status,
		       total_count, processed_count, success_count, failed_count, skipped_count,
		       error_message, started_at, completed_at, created_at, updated_at
		FROM import_jobs ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []*ImportJob
	for rows.Next() {
		j, err := scanJobRow(rows)
		if err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, j)
	}
	return jobs, total, rows.Err()
}

func (r *Repository) UpdateJobStatus(ctx context.Context, id, status string, extra map[string]any) error {
	now := time.Now().UTC().Format(time.RFC3339)

	switch status {
	case JobStatusRunning:
		_, err := r.db.ExecContext(ctx,
			`UPDATE import_jobs SET status = ?, started_at = ?, updated_at = ? WHERE id = ?`,
			status, now, now, id)
		return err
	case JobStatusDone, JobStatusFailed, JobStatusCancelled:
		msg := ""
		if v, ok := extra["error_message"]; ok {
			if s, ok := v.(string); ok {
				msg = s
			}
		}
		_, err := r.db.ExecContext(ctx,
			`UPDATE import_jobs SET status = ?, completed_at = ?, error_message = ?, updated_at = ? WHERE id = ?`,
			status, now, nullableStr(msg), now, id)
		return err
	default:
		_, err := r.db.ExecContext(ctx,
			`UPDATE import_jobs SET status = ?, updated_at = ? WHERE id = ?`,
			status, now, id)
		return err
	}
}

func (r *Repository) IncrementJobCounts(ctx context.Context, id string, success, failed, skipped int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_jobs
		SET processed_count = processed_count + ?,
		    success_count   = success_count   + ?,
		    failed_count    = failed_count    + ?,
		    skipped_count   = skipped_count   + ?,
		    updated_at = ?
		WHERE id = ?`,
		success+failed+skipped, success, failed, skipped, now, id,
	)
	return err
}

// --- import_job_items ---

func (r *Repository) CreateItem(ctx context.Context, item *ImportJobItem) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO import_job_items
		  (id, job_id, file_path, status, created_at, updated_at)
		VALUES (?,?,?,?,?,?)`,
		item.ID, item.JobID, item.FilePath, item.Status,
		item.CreatedAt.UTC().Format(time.RFC3339),
		item.UpdatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *Repository) BulkCreateItems(ctx context.Context, items []*ImportJobItem) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO import_job_items (id, job_id, file_path, status, created_at, updated_at)
		VALUES (?,?,?,?,?,?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.ExecContext(ctx,
			item.ID, item.JobID, item.FilePath, item.Status, now, now); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) GetPendingItems(ctx context.Context, jobID string) ([]*ImportJobItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, job_id, file_path, document_id, status, error_message, created_at, updated_at
		FROM import_job_items
		WHERE job_id = ? AND status = 'pending'
		ORDER BY file_path ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*ImportJobItem
	for rows.Next() {
		item, err := scanItemRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateItemStatus(ctx context.Context, id, status, docID, errMsg string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_job_items
		SET status = ?, document_id = ?, error_message = ?, updated_at = ?
		WHERE id = ?`,
		status, nullableStr(docID), nullableStr(errMsg), now, id,
	)
	return err
}

func (r *Repository) ResetFailedItems(ctx context.Context, jobID string) (int, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx, `
		UPDATE import_job_items
		SET status = 'pending', error_message = NULL, updated_at = ?
		WHERE job_id = ? AND status = 'failed'`, now, jobID)
	if err != nil {
		return 0, err
	}
	n, err := result.RowsAffected()
	return int(n), err
}

func (r *Repository) ListItems(ctx context.Context, jobID, statusFilter string, limit, offset int) ([]*ImportJobItem, int, error) {
	where := "WHERE job_id = ?"
	args := []any{jobID}
	if statusFilter != "" {
		where += " AND status = ?"
		args = append(args, statusFilter)
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM import_job_items `+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, job_id, file_path, document_id, status, error_message, created_at, updated_at
		FROM import_job_items `+where+`
		ORDER BY file_path ASC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*ImportJobItem
	for rows.Next() {
		item, err := scanItemRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// --- document_metadata ---

func (r *Repository) UpsertMetadata(ctx context.Context, m *DocumentMetadata) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO document_metadata
		  (id, document_id, headings, code_languages, code_block_count, link_count, image_count, extracted_at)
		VALUES (?,?,?,?,?,?,?,?)
		ON CONFLICT(document_id) DO UPDATE SET
		  headings = excluded.headings,
		  code_languages = excluded.code_languages,
		  code_block_count = excluded.code_block_count,
		  link_count = excluded.link_count,
		  image_count = excluded.image_count,
		  extracted_at = excluded.extracted_at`,
		m.ID, m.DocumentID,
		nullableStr(m.Headings), nullableStr(m.CodeLanguages),
		m.CodeBlockCount, m.LinkCount, m.ImageCount,
		m.ExtractedAt.UTC().Format(time.RFC3339),
	)
	return err
}

// --- scan helpers ---

func scanJob(row *sql.Row) (*ImportJob, error) {
	var j ImportJob
	var createdAt, updatedAt string
	var startedAt, completedAt, errorMsg *string

	err := row.Scan(
		&j.ID, &j.Name, &j.SourcePath, &j.Status,
		&j.TotalCount, &j.ProcessedCount, &j.SuccessCount, &j.FailedCount, &j.SkippedCount,
		&errorMsg, &startedAt, &completedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if errorMsg != nil {
		j.ErrorMessage = *errorMsg
	}
	j.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	j.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	if startedAt != nil {
		t, _ := time.Parse(time.RFC3339, *startedAt)
		j.StartedAt = &t
	}
	if completedAt != nil {
		t, _ := time.Parse(time.RFC3339, *completedAt)
		j.CompletedAt = &t
	}
	return &j, nil
}

func scanJobRow(rows *sql.Rows) (*ImportJob, error) {
	var j ImportJob
	var createdAt, updatedAt string
	var startedAt, completedAt, errorMsg *string

	err := rows.Scan(
		&j.ID, &j.Name, &j.SourcePath, &j.Status,
		&j.TotalCount, &j.ProcessedCount, &j.SuccessCount, &j.FailedCount, &j.SkippedCount,
		&errorMsg, &startedAt, &completedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if errorMsg != nil {
		j.ErrorMessage = *errorMsg
	}
	j.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	j.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	if startedAt != nil {
		t, _ := time.Parse(time.RFC3339, *startedAt)
		j.StartedAt = &t
	}
	if completedAt != nil {
		t, _ := time.Parse(time.RFC3339, *completedAt)
		j.CompletedAt = &t
	}
	return &j, nil
}

func scanItemRow(rows *sql.Rows) (*ImportJobItem, error) {
	var item ImportJobItem
	var createdAt, updatedAt string
	var docID, errMsg *string

	err := rows.Scan(
		&item.ID, &item.JobID, &item.FilePath, &docID, &item.Status, &errMsg,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	if docID != nil {
		item.DocumentID = *docID
	}
	if errMsg != nil {
		item.ErrorMessage = *errMsg
	}
	item.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &item, nil
}

func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
