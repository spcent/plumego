package ai

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

// Repository handles all AI-related database operations.
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func newID() string { return ulid.Make().String() }

// --- AI Tasks ---

func (r *Repository) CreateTask(taskType, inputJSON string) (*AITask, error) {
	t := &AITask{
		ID:        newID(),
		TaskType:  taskType,
		Status:    TaskStatusPending,
		InputJSON: inputJSON,
		CreatedAt: now(),
		UpdatedAt: now(),
	}
	_, err := r.db.Exec(`
		INSERT INTO ai_tasks (id, task_type, status, input_json, retry_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, 0, ?, ?)`,
		t.ID, t.TaskType, t.Status, t.InputJSON, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create ai_task: %w", err)
	}
	return t, nil
}

func (r *Repository) GetTask(id string) (*AITask, error) {
	row := r.db.QueryRow(`SELECT id, task_type, status, input_json,
		output_document_id, output_json, provider, model,
		error_message, retry_count, started_at, finished_at, created_at, updated_at
		FROM ai_tasks WHERE id = ?`, id)
	return scanTask(row)
}

func (r *Repository) ListTasks(status string, limit, offset int) ([]*AITask, int, error) {
	args := []any{}
	where := ""
	if status != "" {
		where = "WHERE status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM ai_tasks "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.db.Query(`SELECT id, task_type, status, input_json,
		output_document_id, output_json, provider, model,
		error_message, retry_count, started_at, finished_at, created_at, updated_at
		FROM ai_tasks `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var tasks []*AITask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *Repository) ClaimPendingTask() (*AITask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	row := tx.QueryRow(`SELECT id, task_type, status, input_json,
		output_document_id, output_json, provider, model,
		error_message, retry_count, started_at, finished_at, created_at, updated_at
		FROM ai_tasks WHERE status = 'pending' ORDER BY created_at ASC LIMIT 1`)
	task, err := scanTask(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	n := now()
	_, err = tx.Exec(`UPDATE ai_tasks SET status='running', started_at=?, updated_at=? WHERE id=?`,
		n, n, task.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	task.Status = TaskStatusRunning
	task.StartedAt = &n
	return task, nil
}

func (r *Repository) CompleteTask(id, outputJSON, providerName, model string, outputDocID *string) error {
	n := now()
	_, err := r.db.Exec(`UPDATE ai_tasks
		SET status='completed', output_json=?, output_document_id=?, provider=?, model=?, finished_at=?, updated_at=?
		WHERE id=?`,
		outputJSON, outputDocID, providerName, model, n, n, id,
	)
	return err
}

func (r *Repository) FailTask(id, errMsg string, retry bool) error {
	n := now()
	if retry {
		_, err := r.db.Exec(`UPDATE ai_tasks
			SET status='pending', error_message=?, retry_count=retry_count+1, finished_at=?, updated_at=?
			WHERE id=?`, errMsg, n, n, id)
		return err
	}
	_, err := r.db.Exec(`UPDATE ai_tasks
		SET status='dead_letter', error_message=?, finished_at=?, updated_at=?
		WHERE id=?`, errMsg, n, n, id)
	return err
}

func (r *Repository) CancelTask(id string) error {
	n := now()
	_, err := r.db.Exec(`UPDATE ai_tasks SET status='cancelled', updated_at=? WHERE id=? AND status='pending'`, n, id)
	return err
}

func scanTask(s interface {
	Scan(...any) error
}) (*AITask, error) {
	var t AITask
	return &t, s.Scan(
		&t.ID, &t.TaskType, &t.Status, &t.InputJSON,
		&t.OutputDocumentID, &t.OutputJSON, &t.Provider, &t.Model,
		&t.ErrorMessage, &t.RetryCount, &t.StartedAt, &t.FinishedAt,
		&t.CreatedAt, &t.UpdatedAt,
	)
}

// --- Document AI Summaries ---

func (r *Repository) UpsertSummary(s *DocumentAISummary) error {
	if s.ID == "" {
		s.ID = newID()
	}
	n := now()
	if s.CreatedAt == "" {
		s.CreatedAt = n
	}
	_, err := r.db.Exec(`
		INSERT INTO document_ai_summaries
		  (id, document_id, summary, key_points_json, actions_json, code_refs_json, provider, model, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(document_id) DO UPDATE SET
		  summary=excluded.summary,
		  key_points_json=excluded.key_points_json,
		  actions_json=excluded.actions_json,
		  code_refs_json=excluded.code_refs_json,
		  provider=excluded.provider,
		  model=excluded.model,
		  created_at=excluded.created_at`,
		s.ID, s.DocumentID, s.Summary,
		s.KeyPointsJSON, s.ActionsJSON, s.CodeRefsJSON,
		s.Provider, s.Model, s.CreatedAt,
	)
	return err
}

func (r *Repository) GetSummary(documentID string) (*DocumentAISummary, error) {
	row := r.db.QueryRow(`SELECT id, document_id, summary, key_points_json, actions_json,
		code_refs_json, provider, model, created_at
		FROM document_ai_summaries WHERE document_id = ?`, documentID)
	var s DocumentAISummary
	err := row.Scan(&s.ID, &s.DocumentID, &s.Summary,
		&s.KeyPointsJSON, &s.ActionsJSON, &s.CodeRefsJSON,
		&s.Provider, &s.Model, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

// --- Prompts ---

func (r *Repository) CreatePrompt(p *Prompt) error {
	if p.ID == "" {
		p.ID = newID()
	}
	n := now()
	if p.CreatedAt == "" {
		p.CreatedAt = n
	}
	if p.UpdatedAt == "" {
		p.UpdatedAt = n
	}
	_, err := r.db.Exec(`
		INSERT INTO prompts (id, title, content, source_document_id, model_hint, scenario, tags_json, quality_score, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Title, p.Content, p.SourceDocumentID, p.ModelHint,
		p.Scenario, p.TagsJSON, p.QualityScore, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *Repository) GetPrompt(id string) (*Prompt, error) {
	row := r.db.QueryRow(`SELECT id, title, content, source_document_id, model_hint, scenario, tags_json, quality_score, created_at, updated_at
		FROM prompts WHERE id = ?`, id)
	return scanPrompt(row)
}

func (r *Repository) ListPrompts(scenario string, limit, offset int) ([]*Prompt, int, error) {
	args := []any{}
	where := ""
	if scenario != "" {
		where = "WHERE scenario = ?"
		args = append(args, scenario)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM prompts "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.db.Query(`SELECT id, title, content, source_document_id, model_hint, scenario, tags_json, quality_score, created_at, updated_at
		FROM prompts `+where+` ORDER BY quality_score DESC, created_at DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var prompts []*Prompt
	for rows.Next() {
		p, err := scanPrompt(rows)
		if err != nil {
			return nil, 0, err
		}
		prompts = append(prompts, p)
	}
	return prompts, total, rows.Err()
}

func (r *Repository) DeletePrompt(id string) error {
	_, err := r.db.Exec("DELETE FROM prompts WHERE id = ?", id)
	return err
}

func scanPrompt(s interface {
	Scan(...any) error
}) (*Prompt, error) {
	var p Prompt
	return &p, s.Scan(
		&p.ID, &p.Title, &p.Content, &p.SourceDocumentID,
		&p.ModelHint, &p.Scenario, &p.TagsJSON, &p.QualityScore,
		&p.CreatedAt, &p.UpdatedAt,
	)
}

// --- Document Chunks ---

func (r *Repository) DeleteChunks(documentID string) error {
	_, err := r.db.Exec("DELETE FROM document_chunks WHERE document_id = ?", documentID)
	return err
}

func (r *Repository) InsertChunks(chunks []*DocumentChunk) error {
	for _, c := range chunks {
		if c.ID == "" {
			c.ID = newID()
		}
		n := now()
		if c.CreatedAt == "" {
			c.CreatedAt = n
		}
		if c.UpdatedAt == "" {
			c.UpdatedAt = n
		}
		_, err := r.db.Exec(`
			INSERT INTO document_chunks (id, document_id, chunk_index, heading_path, content, content_hash, token_count, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			c.ID, c.DocumentID, c.ChunkIndex, c.HeadingPath,
			c.Content, c.ContentHash, c.TokenCount, c.CreatedAt, c.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetChunks(documentID string) ([]*DocumentChunk, error) {
	rows, err := r.db.Query(`SELECT id, document_id, chunk_index, heading_path, content, content_hash, token_count, created_at, updated_at
		FROM document_chunks WHERE document_id = ? ORDER BY chunk_index ASC`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chunks []*DocumentChunk
	for rows.Next() {
		var c DocumentChunk
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.ChunkIndex, &c.HeadingPath,
			&c.Content, &c.ContentHash, &c.TokenCount, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		chunks = append(chunks, &c)
	}
	return chunks, rows.Err()
}

// --- Helpers for document context ---

type DocMeta struct {
	ID          string
	Title       string
	StorageKey  string
	Summary     *string
	HeadingText *string
}

func (r *Repository) GetDocMeta(id string) (*DocMeta, error) {
	var d DocMeta
	err := r.db.QueryRow(`SELECT id, title, storage_key, summary, heading_text FROM documents WHERE id = ? AND status != 'deleted'`, id).
		Scan(&d.ID, &d.Title, &d.StorageKey, &d.Summary, &d.HeadingText)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &d, err
}

func (r *Repository) SaveSourceLink(documentID, sourceDocumentID, sourceType string) error {
	n := now()
	_, err := r.db.Exec(`
		INSERT INTO document_sources (document_id, source_document_id, source_type, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(document_id, source_document_id) DO NOTHING`,
		documentID, sourceDocumentID, sourceType, n,
	)
	return err
}
