package organize

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cloud-vault/internal/database"
)

// Repository handles all organize-related database operations.
type Repository struct {
	db *database.DB
}

// NewRepository constructs a Repository.
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// --- Similarity ---

func (r *Repository) UpsertSimilarity(ctx context.Context, s *DocumentSimilarity) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
INSERT INTO document_similarity
  (id, document_id_a, document_id_b, similarity_type, similarity_score, reason, status, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?)
ON CONFLICT(document_id_a, document_id_b, similarity_type) DO UPDATE SET
  similarity_score = excluded.similarity_score,
  reason           = excluded.reason,
  updated_at       = excluded.updated_at
`,
		s.ID, s.DocumentIDA, s.DocumentIDB, s.SimilarityType, s.SimilarityScore,
		s.Reason, s.Status, now, now,
	)
	return err
}

func (r *Repository) GetSimilarityByID(ctx context.Context, id string) (*DocumentSimilarity, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, document_id_a, document_id_b, similarity_type, similarity_score, reason, status, created_at, updated_at
FROM document_similarity WHERE id = ?`, id)
	return scanSimilarity(row)
}

func (r *Repository) UpdateSimilarityStatus(ctx context.Context, id, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		`UPDATE document_similarity SET status = ?, updated_at = ? WHERE id = ?`, status, now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListSimilarForDocument(ctx context.Context, docID string) ([]SimilarDoc, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT ds.id,
  CASE WHEN ds.document_id_a = ? THEN ds.document_id_b ELSE ds.document_id_a END AS other_id,
  d.title,
  ds.similarity_type, ds.similarity_score, ds.status
FROM document_similarity ds
JOIN documents d ON d.id = CASE WHEN ds.document_id_a = ? THEN ds.document_id_b ELSE ds.document_id_a END
WHERE (ds.document_id_a = ? OR ds.document_id_b = ?)
  AND ds.status != 'ignored'
  AND d.status != 'deleted'
ORDER BY ds.similarity_score DESC
LIMIT 20
`, docID, docID, docID, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SimilarDoc
	for rows.Next() {
		var s SimilarDoc
		if err := rows.Scan(&s.SimilarityID, &s.DocumentID, &s.Title, &s.SimilarityType, &s.Score, &s.Status); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func scanSimilarity(row *sql.Row) (*DocumentSimilarity, error) {
	var s DocumentSimilarity
	var reason sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(&s.ID, &s.DocumentIDA, &s.DocumentIDB, &s.SimilarityType,
		&s.SimilarityScore, &reason, &s.Status, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	s.Reason = reason.String
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &s, nil
}

// --- Duplicate detection helpers ---

// ListDuplicateGroups returns groups of documents that share the same content hash.
func (r *Repository) ListDuplicateGroups(ctx context.Context) ([]DuplicateGroup, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT d.id, d.title, d.status, d.is_favorite, d.created_at, d.updated_at,
       COALESCE(d.import_job_id,''), COALESCE(d.original_path,''), d.content_hash
FROM documents d
WHERE d.status != 'deleted'
  AND d.content_hash IN (
    SELECT content_hash FROM documents WHERE status != 'deleted'
    GROUP BY content_hash HAVING COUNT(*) > 1
  )
ORDER BY d.content_hash, d.created_at
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groupMap := make(map[string]*DuplicateGroup)
	var order []string
	for rows.Next() {
		var doc DuplicateDoc
		var hash, createdAt, updatedAt string
		var isFav int
		err := rows.Scan(&doc.ID, &doc.Title, &doc.Status, &isFav,
			&createdAt, &updatedAt, &doc.ImportJobID, &doc.OriginalPath, &hash)
		if err != nil {
			return nil, err
		}
		doc.IsFavorite = isFav == 1
		doc.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		doc.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		if _, ok := groupMap[hash]; !ok {
			groupMap[hash] = &DuplicateGroup{ContentHash: hash}
			order = append(order, hash)
		}
		groupMap[hash].Documents = append(groupMap[hash].Documents, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]DuplicateGroup, 0, len(order))
	for _, h := range order {
		result = append(result, *groupMap[h])
	}
	return result, nil
}

// ListAllDocsForFingerprint returns lightweight doc data for fingerprinting.
func (r *Repository) ListAllDocsForFingerprint(ctx context.Context) ([]docFingerprintRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT d.id, d.title, COALESCE(d.original_path,''), COALESCE(d.import_job_id,''),
       COALESCE(dm.headings,''), COALESCE(d.summary,'')
FROM documents d
LEFT JOIN document_metadata dm ON dm.document_id = d.id
WHERE d.status != 'deleted'
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []docFingerprintRow
	for rows.Next() {
		var row docFingerprintRow
		if err := rows.Scan(&row.ID, &row.Title, &row.OriginalPath, &row.ImportJobID,
			&row.HeadingsJSON, &row.Summary); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

type docFingerprintRow struct {
	ID           string
	Title        string
	OriginalPath string
	ImportJobID  string
	HeadingsJSON string
	Summary      string
}

// ListDocTagIDs returns a map of docID → []tagID for all non-deleted docs.
func (r *Repository) ListDocTagIDs(ctx context.Context) (map[string][]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT dt.document_id, dt.tag_id
FROM document_tags dt
JOIN documents d ON d.id = dt.document_id
WHERE d.status != 'deleted'
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]string)
	for rows.Next() {
		var docID, tagID string
		if err := rows.Scan(&docID, &tagID); err != nil {
			return nil, err
		}
		result[docID] = append(result[docID], tagID)
	}
	return result, rows.Err()
}

// ListTagNames returns a map of tagID → tagName.
func (r *Repository) ListTagNames(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM tags`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		result[id] = name
	}
	return result, rows.Err()
}

// --- Fingerprints ---

func (r *Repository) UpsertFingerprint(ctx context.Context, f *DocumentFingerprint) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
INSERT INTO document_fingerprints
  (document_id, title_norm, simhash, heading_hash, keyword_hash, created_at, updated_at)
VALUES (?,?,?,?,?,?,?)
ON CONFLICT(document_id) DO UPDATE SET
  title_norm   = excluded.title_norm,
  simhash      = excluded.simhash,
  heading_hash = excluded.heading_hash,
  keyword_hash = excluded.keyword_hash,
  updated_at   = excluded.updated_at
`,
		f.DocumentID, f.TitleNorm, f.Simhash, f.HeadingHash, f.KeywordHash, now, now,
	)
	return err
}

// --- Tag suggestions ---

func (r *Repository) UpsertTagSuggestion(ctx context.Context, s *TagSuggestion) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `
INSERT INTO tag_suggestions
  (id, document_id, tag_id, tag_name, source, confidence, status, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO NOTHING
`,
		s.ID, s.DocumentID, nilStr(s.TagID), s.TagName, s.Source, s.Confidence, s.Status,
		s.CreatedAt.UTC().Format(time.RFC3339), now,
	)
	return err
}

func nilStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *Repository) GetTagSuggestionByID(ctx context.Context, id string) (*TagSuggestion, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, document_id, COALESCE(tag_id,''), tag_name, source, confidence, status, created_at, updated_at
FROM tag_suggestions WHERE id = ?`, id)
	return scanTagSuggestion(row)
}

func (r *Repository) UpdateTagSuggestionStatus(ctx context.Context, id, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		`UPDATE tag_suggestions SET status = ?, updated_at = ? WHERE id = ?`, status, now, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListTagSuggestionsForDoc(ctx context.Context, docID string) ([]TagSuggestion, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, document_id, COALESCE(tag_id,''), tag_name, source, confidence, status, created_at, updated_at
FROM tag_suggestions WHERE document_id = ? ORDER BY confidence DESC`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TagSuggestion
	for rows.Next() {
		var s TagSuggestion
		var createdAt, updatedAt string
		if err := rows.Scan(&s.ID, &s.DocumentID, &s.TagID, &s.TagName, &s.Source,
			&s.Confidence, &s.Status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		result = append(result, s)
	}
	return result, rows.Err()
}

func scanTagSuggestion(row *sql.Row) (*TagSuggestion, error) {
	var s TagSuggestion
	var createdAt, updatedAt string
	err := row.Scan(&s.ID, &s.DocumentID, &s.TagID, &s.TagName, &s.Source,
		&s.Confidence, &s.Status, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &s, nil
}

// FindTagByName returns the tag ID for a given name, or "" if not found.
func (r *Repository) FindTagByName(ctx context.Context, name string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ?`, name).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

// CreateTag inserts a new tag and returns its ID.
func (r *Repository) CreateTag(ctx context.Context, id, name, source string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tags (id, name, source, created_at) VALUES (?,?,?,?)`, id, name, source, now)
	return err
}

// AddDocumentTag adds a tag to a document (idempotent).
func (r *Repository) AddDocumentTag(ctx context.Context, docID, tagID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO document_tags (document_id, tag_id) VALUES (?,?) ON CONFLICT DO NOTHING`, docID, tagID)
	return err
}

// --- Topics ---

func (r *Repository) UpsertTopic(ctx context.Context, t *Topic) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
INSERT INTO topics (id, name, description, source, status, created_at, updated_at)
VALUES (?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  name        = excluded.name,
  description = excluded.description,
  updated_at  = excluded.updated_at
`,
		t.ID, t.Name, nilStr(t.Description), t.Source, t.Status, now, now,
	)
	return err
}

func (r *Repository) AddTopicDocument(ctx context.Context, td *TopicDocument) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
INSERT INTO topic_documents (topic_id, document_id, score, source, created_at)
VALUES (?,?,?,?,?)
ON CONFLICT(topic_id, document_id) DO UPDATE SET score = excluded.score
`,
		td.TopicID, td.DocumentID, td.Score, td.Source, now,
	)
	return err
}

func (r *Repository) ListTopics(ctx context.Context) ([]Topic, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT t.id, t.name, COALESCE(t.description,''), t.source, t.status,
       COUNT(td.document_id) AS doc_count,
       t.created_at, t.updated_at
FROM topics t
LEFT JOIN topic_documents td ON td.topic_id = t.id
WHERE t.status = 'active'
GROUP BY t.id
ORDER BY doc_count DESC, t.name
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Topic
	for rows.Next() {
		var topic Topic
		var createdAt, updatedAt string
		if err := rows.Scan(&topic.ID, &topic.Name, &topic.Description, &topic.Source,
			&topic.Status, &topic.DocCount, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		topic.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		topic.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		result = append(result, topic)
	}
	return result, rows.Err()
}

func (r *Repository) GetTopicByID(ctx context.Context, id string) (*Topic, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT t.id, t.name, COALESCE(t.description,''), t.source, t.status,
       COUNT(td.document_id) AS doc_count,
       t.created_at, t.updated_at
FROM topics t
LEFT JOIN topic_documents td ON td.topic_id = t.id
WHERE t.id = ?
GROUP BY t.id
`, id)
	var topic Topic
	var createdAt, updatedAt string
	err := row.Scan(&topic.ID, &topic.Name, &topic.Description, &topic.Source,
		&topic.Status, &topic.DocCount, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	topic.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	topic.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &topic, nil
}

func (r *Repository) ListTopicDocuments(ctx context.Context, topicID string, limit, offset int) ([]DuplicateDoc, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM topic_documents WHERE topic_id = ?`, topicID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT d.id, d.title, d.status, d.is_favorite, d.created_at, d.updated_at,
       COALESCE(d.import_job_id,''), COALESCE(d.original_path,'')
FROM topic_documents td
JOIN documents d ON d.id = td.document_id
WHERE td.topic_id = ? AND d.status != 'deleted'
ORDER BY td.score DESC
LIMIT ? OFFSET ?
`, topicID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var docs []DuplicateDoc
	for rows.Next() {
		var doc DuplicateDoc
		var isFav int
		var createdAt, updatedAt string
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.Status, &isFav,
			&createdAt, &updatedAt, &doc.ImportJobID, &doc.OriginalPath); err != nil {
			return nil, 0, err
		}
		doc.IsFavorite = isFav == 1
		doc.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		doc.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		docs = append(docs, doc)
	}
	return docs, total, rows.Err()
}

// --- Organize jobs ---

func (r *Repository) CreateJob(ctx context.Context, job *OrganizeJob) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
INSERT INTO organize_jobs
  (id, job_type, status, total_items, processed_items, failed_items, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?)
`,
		job.ID, job.JobType, job.Status, job.TotalItems, job.ProcessedItems, job.FailedItems,
		now, now,
	)
	return err
}

func (r *Repository) UpdateJobStatus(ctx context.Context, id, status string, processedItems, failedItems int, errMsg string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	var finishedAt any
	var startedAt any
	if status == JobStatusRunning {
		startedAt = now
	}
	if status == JobStatusDone || status == JobStatusFailed {
		finishedAt = now
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE organize_jobs SET
  status = ?, processed_items = ?, failed_items = ?,
  error_message = ?,
  started_at  = COALESCE(started_at,  ?),
  finished_at = COALESCE(finished_at, ?),
  updated_at  = ?
WHERE id = ?
`,
		status, processedItems, failedItems, nilStr(errMsg),
		startedAt, finishedAt, now, id,
	)
	return err
}

func (r *Repository) ListJobs(ctx context.Context) ([]OrganizeJob, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, job_type, status, total_items, processed_items, failed_items,
       COALESCE(error_message,''), started_at, finished_at, created_at, updated_at
FROM organize_jobs ORDER BY created_at DESC LIMIT 50
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []OrganizeJob
	for rows.Next() {
		var job OrganizeJob
		var startedAt, finishedAt sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&job.ID, &job.JobType, &job.Status, &job.TotalItems,
			&job.ProcessedItems, &job.FailedItems, &job.ErrorMessage,
			&startedAt, &finishedAt, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			t, _ := time.Parse(time.RFC3339, startedAt.String)
			job.StartedAt = &t
		}
		if finishedAt.Valid {
			t, _ := time.Parse(time.RFC3339, finishedAt.String)
			job.FinishedAt = &t
		}
		job.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		job.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		result = append(result, job)
	}
	return result, rows.Err()
}

func (r *Repository) GetJobByID(ctx context.Context, id string) (*OrganizeJob, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, job_type, status, total_items, processed_items, failed_items,
       COALESCE(error_message,''), started_at, finished_at, created_at, updated_at
FROM organize_jobs WHERE id = ?`, id)
	var job OrganizeJob
	var startedAt, finishedAt sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(&job.ID, &job.JobType, &job.Status, &job.TotalItems,
		&job.ProcessedItems, &job.FailedItems, &job.ErrorMessage,
		&startedAt, &finishedAt, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if startedAt.Valid {
		t, _ := time.Parse(time.RFC3339, startedAt.String)
		job.StartedAt = &t
	}
	if finishedAt.Valid {
		t, _ := time.Parse(time.RFC3339, finishedAt.String)
		job.FinishedAt = &t
	}
	job.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	job.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &job, nil
}

// --- Quality score ---

func (r *Repository) UpdateQualityScore(ctx context.Context, docID string, score float64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`UPDATE documents SET quality_score = ?, updated_at = ? WHERE id = ?`, score, now, docID)
	return err
}

// --- Prompt candidate ---

func (r *Repository) UpdatePromptCandidate(ctx context.Context, docID string, isCandidate bool, score float64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	var candidate int
	if isCandidate {
		candidate = 1
	}
	_, err := r.db.ExecContext(ctx, `
INSERT INTO document_metadata (id, document_id, is_prompt_candidate, prompt_score, extracted_at)
VALUES (?,?,?,?,?)
ON CONFLICT(document_id) DO UPDATE SET
  is_prompt_candidate = excluded.is_prompt_candidate,
  prompt_score        = excluded.prompt_score
`,
		newID(), docID, candidate, score, now,
	)
	return err
}

// ListDocumentIDsForQuality returns all non-deleted document IDs + metadata needed for scoring.
func (r *Repository) ListDocumentsForScoring(ctx context.Context) ([]docScoringRow, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT d.id, d.title, d.word_count, d.is_favorite, d.content_hash,
       COALESCE(dm.code_block_count, 0), COALESCE(dm.headings,''),
       d.review_status
FROM documents d
LEFT JOIN document_metadata dm ON dm.document_id = d.id
WHERE d.status != 'deleted'
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []docScoringRow
	for rows.Next() {
		var row docScoringRow
		var isFav int
		if err := rows.Scan(&row.ID, &row.Title, &row.WordCount, &isFav,
			&row.ContentHash, &row.CodeBlockCount, &row.HeadingsJSON, &row.ReviewStatus); err != nil {
			return nil, err
		}
		row.IsFavorite = isFav == 1
		result = append(result, row)
	}
	return result, rows.Err()
}

type docScoringRow struct {
	ID             string
	Title          string
	WordCount      int
	IsFavorite     bool
	ContentHash    string
	CodeBlockCount int
	HeadingsJSON   string
	ReviewStatus   string
}

// IsDuplicateHash reports if there is more than one non-deleted document with the given content hash.
func (r *Repository) IsDuplicateHash(ctx context.Context, hash string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM documents WHERE content_hash = ? AND status != 'deleted'`, hash).Scan(&count)
	return count > 1, err
}

// BatchArchiveDocuments sets the status of the given document IDs to 'archived'.
func (r *Repository) BatchArchiveDocuments(ctx context.Context, ids []string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	for _, id := range ids {
		if _, err := r.db.ExecContext(ctx,
			`UPDATE documents SET status = 'archived', updated_at = ? WHERE id = ?`, now, id); err != nil {
			return err
		}
	}
	return nil
}

// BatchMarkDuplicate sets status to 'archived' and review_status to 'duplicate' for the given IDs.
func (r *Repository) BatchMarkDuplicate(ctx context.Context, ids []string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	for _, id := range ids {
		if _, err := r.db.ExecContext(ctx, `
UPDATE documents SET status = 'archived', review_status = 'duplicate', updated_at = ?
WHERE id = ?`, now, id); err != nil {
			return err
		}
	}
	return nil
}

// --- Review queue ---

func (r *Repository) ListPromptCandidates(ctx context.Context, limit, offset int) ([]ReviewItem, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM document_metadata dm
JOIN documents d ON d.id = dm.document_id
WHERE dm.is_prompt_candidate = 1 AND d.status != 'deleted'
`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT d.id, d.title, dm.prompt_score
FROM document_metadata dm
JOIN documents d ON d.id = dm.document_id
WHERE dm.is_prompt_candidate = 1 AND d.status != 'deleted'
ORDER BY dm.prompt_score DESC
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []ReviewItem
	for rows.Next() {
		var item ReviewItem
		if err := rows.Scan(&item.DocumentID, &item.Title, &item.Score); err != nil {
			return nil, 0, err
		}
		item.Type = "prompt_candidate"
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) ListLowQualityDocs(ctx context.Context, limit, offset int) ([]ReviewItem, int, error) {
	if limit <= 0 {
		limit = 50
	}
	const threshold = 30.0
	var total int
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM documents WHERE quality_score < ? AND status != 'deleted'
`, threshold).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, title, quality_score FROM documents
WHERE quality_score < ? AND status != 'deleted'
ORDER BY quality_score ASC
LIMIT ? OFFSET ?
`, threshold, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []ReviewItem
	for rows.Next() {
		var item ReviewItem
		if err := rows.Scan(&item.DocumentID, &item.Title, &item.Score); err != nil {
			return nil, 0, err
		}
		item.Type = "low_quality"
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) ListPendingTagSuggestions(ctx context.Context, limit, offset int) ([]ReviewItem, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tag_suggestions WHERE status = 'pending'`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT ts.id, d.id, d.title, ts.tag_name, ts.confidence
FROM tag_suggestions ts
JOIN documents d ON d.id = ts.document_id
WHERE ts.status = 'pending' AND d.status != 'deleted'
ORDER BY ts.confidence DESC
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []ReviewItem
	for rows.Next() {
		var suggID, docID, title, tagName string
		var confidence float64
		if err := rows.Scan(&suggID, &docID, &title, &tagName, &confidence); err != nil {
			return nil, 0, err
		}
		items = append(items, ReviewItem{
			Type:       "tag_suggestion",
			DocumentID: docID,
			Title:      title,
			Score:      confidence,
			Extra:      map[string]any{"suggestion_id": suggID, "tag_name": tagName},
		})
	}
	return items, total, rows.Err()
}

func (r *Repository) ListPendingSimilarities(ctx context.Context, limit, offset int) ([]ReviewItem, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM document_similarity WHERE status = 'pending'`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT ds.id, da.id, da.title, ds.similarity_type, ds.similarity_score, db.title
FROM document_similarity ds
JOIN documents da ON da.id = ds.document_id_a
JOIN documents db ON db.id = ds.document_id_b
WHERE ds.status = 'pending' AND da.status != 'deleted' AND db.status != 'deleted'
ORDER BY ds.similarity_score DESC
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []ReviewItem
	for rows.Next() {
		var simID, docIDA, titleA, simType, titleB string
		var score float64
		if err := rows.Scan(&simID, &docIDA, &titleA, &simType, &score, &titleB); err != nil {
			return nil, 0, err
		}
		items = append(items, ReviewItem{
			Type:       "similar",
			DocumentID: docIDA,
			Title:      titleA,
			Score:      score,
			Extra: map[string]any{
				"similarity_id":   simID,
				"similarity_type": simType,
				"other_title":     titleB,
			},
		})
	}
	return items, total, rows.Err()
}

func (r *Repository) ListDuplicateReviewItems(ctx context.Context, limit, offset int) ([]ReviewItem, int, error) {
	groups, err := r.ListDuplicateGroups(ctx)
	if err != nil {
		return nil, 0, err
	}

	total := len(groups)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset >= total {
		return nil, total, nil
	}

	var items []ReviewItem
	for _, g := range groups[offset:end] {
		if len(g.Documents) == 0 {
			continue
		}
		items = append(items, ReviewItem{
			Type:       "duplicate",
			DocumentID: g.Documents[0].ID,
			Title:      g.Documents[0].Title,
			Score:      1.0,
			Extra: map[string]any{
				"content_hash": g.ContentHash,
				"count":        fmt.Sprintf("%d", len(g.Documents)),
			},
		})
	}
	return items, total, nil
}
