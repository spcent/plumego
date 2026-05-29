package ai

// Task type constants.
const (
	TaskTypeDocumentSummary = "document_summary"
	TaskTypeTopicSummary    = "topic_summary"
	TaskTypeSearchAnswer    = "search_answer"
	TaskTypePromptExtract   = "prompt_extract"
)

// Task status constants.
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"
)

// AITask represents a queued or completed AI job.
type AITask struct {
	ID               string
	TaskType         string
	Status           string
	InputJSON        string
	OutputDocumentID *string
	OutputJSON       *string
	Provider         *string
	Model            *string
	ErrorMessage     *string
	RetryCount       int
	StartedAt        *string
	FinishedAt       *string
	CreatedAt        string
	UpdatedAt        string
}

// DocumentAISummary is a structured AI-generated summary for one document.
type DocumentAISummary struct {
	ID            string
	DocumentID    string
	Summary       string
	KeyPointsJSON *string
	ActionsJSON   *string
	CodeRefsJSON  *string
	Provider      *string
	Model         *string
	CreatedAt     string
}

// Prompt is an entry in the prompt library.
type Prompt struct {
	ID               string
	Title            string
	Content          string
	SourceDocumentID *string
	ModelHint        *string
	Scenario         *string
	TagsJSON         *string
	QualityScore     float64
	CreatedAt        string
	UpdatedAt        string
}

// DocumentChunk is a heading-split slice of a document's content.
type DocumentChunk struct {
	ID          string
	DocumentID  string
	ChunkIndex  int
	HeadingPath *string
	Content     string
	ContentHash string
	TokenCount  int
	CreatedAt   string
	UpdatedAt   string
}

// SummaryInput is the JSON payload stored in ai_tasks.input_json for document_summary tasks.
type SummaryInput struct {
	DocumentID string `json:"document_id"`
}

// SummaryOutput is stored in ai_tasks.output_json after completion.
type SummaryOutput struct {
	Summary   string   `json:"summary"`
	KeyPoints []string `json:"key_points"`
	Actions   []string `json:"actions"`
	CodeRefs  []string `json:"code_refs"`
}

// QAInput is the JSON payload for search_answer tasks.
type QAInput struct {
	Question    string   `json:"question"`
	DocumentIDs []string `json:"document_ids"`
}

// QAOutput is stored after a Q&A task completes.
type QAOutput struct {
	Answer      string     `json:"answer"`
	Citations   []Citation `json:"citations"`
	DocumentIDs []string   `json:"document_ids"`
}

// Citation points to a specific document used as a source.
type Citation struct {
	DocumentID    string `json:"document_id"`
	DocumentTitle string `json:"document_title"`
	Excerpt       string `json:"excerpt"`
}

// PromptExtractInput is the JSON payload for prompt_extract tasks.
type PromptExtractInput struct {
	DocumentID string `json:"document_id"`
}

// PromptExtractOutput holds extracted prompt information.
type PromptExtractOutput struct {
	Title        string  `json:"title"`
	Content      string  `json:"content"`
	ModelHint    string  `json:"model_hint"`
	Scenario     string  `json:"scenario"`
	QualityScore float64 `json:"quality_score"`
}
