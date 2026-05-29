package organize

import (
	"context"
	"errors"
	"fmt"

	"cloud-vault/internal/config"
	"cloud-vault/internal/storage"
)

// Service orchestrates V0.4 organize operations.
type Service struct {
	repo  *Repository
	store storage.ObjectStorage
	cfg   config.OrganizeConfig
}

// NewService constructs the organize Service.
func NewService(repo *Repository, store storage.ObjectStorage, cfg config.OrganizeConfig) *Service {
	return &Service{repo: repo, store: store, cfg: cfg}
}

// --- Duplicate operations ---

// DetectDuplicates runs exact-duplicate detection and returns a job record.
func (s *Service) DetectDuplicates(ctx context.Context) (*OrganizeJob, error) {
	job := &OrganizeJob{ID: newID(), JobType: "detect_duplicates", Status: JobStatusRunning}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	n, err := s.detectDuplicates(ctx)
	if err != nil {
		_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusFailed, 0, 0, err.Error())
		return nil, err
	}
	_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusDone, n, 0, "")
	job.ProcessedItems = n
	job.Status = JobStatusDone
	return job, nil
}

// ListDuplicates returns groups of documents sharing the same content hash.
func (s *Service) ListDuplicates(ctx context.Context) ([]DuplicateGroup, error) {
	return s.repo.ListDuplicateGroups(ctx)
}

// ResolveDuplicates applies the chosen action to the non-kept duplicate documents.
func (s *Service) ResolveDuplicates(ctx context.Context, req ResolveDuplicatesRequest) error {
	if req.KeepDocumentID == "" {
		return errors.New("keep_document_id is required")
	}
	if len(req.DuplicateDocumentIDs) == 0 {
		return errors.New("duplicate_document_ids must not be empty")
	}

	switch req.Action {
	case "archive":
		return s.repo.BatchArchiveDocuments(ctx, req.DuplicateDocumentIDs)
	case "mark_duplicate":
		return s.repo.BatchMarkDuplicate(ctx, req.DuplicateDocumentIDs)
	case "ignore":
		return nil
	default:
		return fmt.Errorf("unknown action %q: must be archive, mark_duplicate, or ignore", req.Action)
	}
}

// --- Similarity operations ---

// DetectSimilarity runs near-duplicate/related detection.
func (s *Service) DetectSimilarity(ctx context.Context) (*OrganizeJob, error) {
	job := &OrganizeJob{ID: newID(), JobType: "detect_similarity", Status: JobStatusRunning}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	n, err := s.detectSimilarity(ctx)
	if err != nil {
		_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusFailed, 0, 0, err.Error())
		return nil, err
	}
	_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusDone, n, 0, "")
	job.ProcessedItems = n
	job.Status = JobStatusDone
	return job, nil
}

// GetSimilarDocuments returns similar documents for a given doc ID.
func (s *Service) GetSimilarDocuments(ctx context.Context, docID string) ([]SimilarDoc, error) {
	return s.repo.ListSimilarForDocument(ctx, docID)
}

// IgnoreSimilarity marks a similarity record as ignored.
func (s *Service) IgnoreSimilarity(ctx context.Context, id string) error {
	return s.repo.UpdateSimilarityStatus(ctx, id, SimilarityStatusIgnored)
}

// ConfirmSimilarity marks a similarity record as confirmed.
func (s *Service) ConfirmSimilarity(ctx context.Context, id string) error {
	return s.repo.UpdateSimilarityStatus(ctx, id, SimilarityStatusConfirmed)
}

// --- Tag suggestions ---

// SuggestTags generates tag suggestions for all documents.
func (s *Service) SuggestTags(ctx context.Context) (*OrganizeJob, error) {
	job := &OrganizeJob{ID: newID(), JobType: "suggest_tags", Status: JobStatusRunning}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	n, err := s.suggestTags(ctx)
	if err != nil {
		_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusFailed, 0, 0, err.Error())
		return nil, err
	}
	_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusDone, n, 0, "")
	job.ProcessedItems = n
	job.Status = JobStatusDone
	return job, nil
}

// GetTagSuggestions returns pending tag suggestions for a document.
func (s *Service) GetTagSuggestions(ctx context.Context, docID string) ([]TagSuggestion, error) {
	return s.repo.ListTagSuggestionsForDoc(ctx, docID)
}

// AcceptTagSuggestion applies a tag suggestion (creates tag if needed, links to document).
func (s *Service) AcceptTagSuggestion(ctx context.Context, id string) error {
	sugg, err := s.repo.GetTagSuggestionByID(ctx, id)
	if err != nil {
		return err
	}
	if sugg.Status != TagSuggestionStatusPending {
		return fmt.Errorf("suggestion is already %s", sugg.Status)
	}

	tagID := sugg.TagID
	if tagID == "" {
		// Check if tag already exists by name.
		existing, err := s.repo.FindTagByName(ctx, sugg.TagName)
		if err != nil {
			return fmt.Errorf("find tag: %w", err)
		}
		if existing != "" {
			tagID = existing
		} else {
			// Create the tag.
			tid := newID()
			if err := s.repo.CreateTag(ctx, tid, sugg.TagName, "organize"); err != nil {
				return fmt.Errorf("create tag: %w", err)
			}
			tagID = tid
		}
	}

	if err := s.repo.AddDocumentTag(ctx, sugg.DocumentID, tagID); err != nil {
		return fmt.Errorf("add document tag: %w", err)
	}
	return s.repo.UpdateTagSuggestionStatus(ctx, id, TagSuggestionStatusAccepted)
}

// RejectTagSuggestion marks a tag suggestion as rejected.
func (s *Service) RejectTagSuggestion(ctx context.Context, id string) error {
	return s.repo.UpdateTagSuggestionStatus(ctx, id, TagSuggestionStatusRejected)
}

// BatchAcceptTagSuggestions accepts multiple tag suggestions.
func (s *Service) BatchAcceptTagSuggestions(ctx context.Context, ids []string) (int, error) {
	accepted := 0
	for _, id := range ids {
		if err := s.AcceptTagSuggestion(ctx, id); err == nil {
			accepted++
		}
	}
	return accepted, nil
}

// --- Topics ---

// BuildTopics builds rule-based topics.
func (s *Service) BuildTopics(ctx context.Context) (*OrganizeJob, error) {
	job := &OrganizeJob{ID: newID(), JobType: "build_topics", Status: JobStatusRunning}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	n, err := s.buildTopics(ctx)
	if err != nil {
		_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusFailed, 0, 0, err.Error())
		return nil, err
	}
	_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusDone, n, 0, "")
	job.ProcessedItems = n
	job.Status = JobStatusDone
	return job, nil
}

// ListTopics returns all active topics.
func (s *Service) ListTopics(ctx context.Context) ([]Topic, error) {
	return s.repo.ListTopics(ctx)
}

// GetTopic returns a single topic by ID.
func (s *Service) GetTopic(ctx context.Context, id string) (*Topic, error) {
	return s.repo.GetTopicByID(ctx, id)
}

// GetTopicDocuments returns the documents in a topic.
func (s *Service) GetTopicDocuments(ctx context.Context, topicID string, limit, offset int) ([]DuplicateDoc, int, error) {
	return s.repo.ListTopicDocuments(ctx, topicID, limit, offset)
}

// --- Quality scoring ---

// ScoreQuality computes quality scores for all documents.
func (s *Service) ScoreQuality(ctx context.Context) (*OrganizeJob, error) {
	job := &OrganizeJob{ID: newID(), JobType: "score_quality", Status: JobStatusRunning}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	n, err := s.scoreQuality(ctx)
	if err != nil {
		_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusFailed, 0, 0, err.Error())
		return nil, err
	}
	_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusDone, n, 0, "")
	job.ProcessedItems = n
	job.Status = JobStatusDone
	return job, nil
}

// --- Prompt candidate detection ---

// DetectPromptCandidates scans documents for prompt-like content.
func (s *Service) DetectPromptCandidates(ctx context.Context) (*OrganizeJob, error) {
	job := &OrganizeJob{ID: newID(), JobType: "detect_prompt_candidates", Status: JobStatusRunning}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	n, err := s.detectPromptCandidatesFTSBased(ctx)
	if err != nil {
		_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusFailed, 0, 0, err.Error())
		return nil, err
	}
	_ = s.repo.UpdateJobStatus(ctx, job.ID, JobStatusDone, n, 0, "")
	job.ProcessedItems = n
	job.Status = JobStatusDone
	return job, nil
}

// --- Review queue ---

// GetReviewQueue returns the organize review queue filtered by type.
func (s *Service) GetReviewQueue(ctx context.Context, q ReviewQueueQuery) (*ReviewQueueResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	var items []ReviewItem
	var total int
	var err error

	switch q.Type {
	case "duplicates":
		items, total, err = s.repo.ListDuplicateReviewItems(ctx, limit, q.Offset)
	case "similar":
		items, total, err = s.repo.ListPendingSimilarities(ctx, limit, q.Offset)
	case "tag_suggestions":
		items, total, err = s.repo.ListPendingTagSuggestions(ctx, limit, q.Offset)
	case "prompt_candidates":
		items, total, err = s.repo.ListPromptCandidates(ctx, limit, q.Offset)
	case "low_quality":
		items, total, err = s.repo.ListLowQualityDocs(ctx, limit, q.Offset)
	default:
		items, total, err = s.mixedQueue(ctx, limit, q.Offset)
	}

	if err != nil {
		return nil, err
	}
	return &ReviewQueueResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: q.Offset,
	}, nil
}

func (s *Service) mixedQueue(ctx context.Context, limit, offset int) ([]ReviewItem, int, error) {
	var all []ReviewItem

	if dups, _, err := s.repo.ListDuplicateReviewItems(ctx, 20, 0); err == nil {
		all = append(all, dups...)
	}
	if sims, _, err := s.repo.ListPendingSimilarities(ctx, 20, 0); err == nil {
		all = append(all, sims...)
	}
	if tags, _, err := s.repo.ListPendingTagSuggestions(ctx, 20, 0); err == nil {
		all = append(all, tags...)
	}
	if prompts, _, err := s.repo.ListPromptCandidates(ctx, 20, 0); err == nil {
		all = append(all, prompts...)
	}
	if low, _, err := s.repo.ListLowQualityDocs(ctx, 20, 0); err == nil {
		all = append(all, low...)
	}

	total := len(all)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

// --- Jobs ---

// ListJobs returns recent organize jobs.
func (s *Service) ListJobs(ctx context.Context) ([]OrganizeJob, error) {
	return s.repo.ListJobs(ctx)
}

// GetJob returns a single organize job by ID.
func (s *Service) GetJob(ctx context.Context, id string) (*OrganizeJob, error) {
	return s.repo.GetJobByID(ctx, id)
}

// RunAll runs all organize operations in sequence.
func (s *Service) RunAll(ctx context.Context) error {
	if _, err := s.DetectDuplicates(ctx); err != nil {
		return fmt.Errorf("detect duplicates: %w", err)
	}
	if _, err := s.DetectSimilarity(ctx); err != nil {
		return fmt.Errorf("detect similarity: %w", err)
	}
	if _, err := s.SuggestTags(ctx); err != nil {
		return fmt.Errorf("suggest tags: %w", err)
	}
	if _, err := s.BuildTopics(ctx); err != nil {
		return fmt.Errorf("build topics: %w", err)
	}
	if _, err := s.ScoreQuality(ctx); err != nil {
		return fmt.Errorf("score quality: %w", err)
	}
	if _, err := s.DetectPromptCandidates(ctx); err != nil {
		return fmt.Errorf("detect prompt candidates: %w", err)
	}
	return nil
}
