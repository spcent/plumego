package organize

// --- Duplicate detection ---

type DetectDuplicatesRequest struct{}

type ResolveDuplicatesRequest struct {
	KeepDocumentID      string   `json:"keep_document_id"`
	DuplicateDocumentIDs []string `json:"duplicate_document_ids"`
	Action              string   `json:"action"` // archive | mark_duplicate | ignore
}

// --- Similarity ---

type DetectSimilarityRequest struct{}

type IgnoreSimilarityRequest struct{}

type ConfirmSimilarityRequest struct{}

// --- Tag suggestions ---

type SuggestTagsRequest struct{}

type AcceptTagSuggestionRequest struct{}

type RejectTagSuggestionRequest struct{}

type BatchAcceptTagSuggestionsRequest struct {
	IDs []string `json:"ids"`
}

// --- Topics ---

type BuildTopicsRequest struct{}

type TopicToCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// --- Review queue ---

type ReviewQueueQuery struct {
	Type   string // duplicates | similar | tag_suggestions | prompt_candidates | low_quality
	Limit  int
	Offset int
}

type ReviewQueueResult struct {
	Items  []ReviewItem `json:"items"`
	Total  int          `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

// --- Jobs ---

type JobListResult struct {
	Items []OrganizeJob `json:"items"`
}
