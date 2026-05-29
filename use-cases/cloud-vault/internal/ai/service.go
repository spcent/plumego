package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud-vault/internal/config"
	"cloud-vault/internal/storage"
)

// DocumentSaver is a minimal interface to persist AI-generated Markdown back as a document.
type DocumentSaver interface {
	SaveAIDocument(ctx context.Context, title, content, storageKey string, sourceIDs []string) (string, error)
}

// Service orchestrates AI operations.
type Service struct {
	repo            *Repository
	store           storage.ObjectStorage
	cfg             config.AIConfig
	provider        LLMProvider
	summarizer      *Summarizer
	qaEngine        *QAEngine
	promptExtractor *PromptExtractor
	ctxBuilder      *ContextBuilder
	docSaver        DocumentSaver
}

func NewService(
	repo *Repository,
	store storage.ObjectStorage,
	cfg config.AIConfig,
	provider LLMProvider,
	docSaver DocumentSaver,
) *Service {
	cb := NewContextBuilder(repo, store, cfg.MaxContextTokens)
	return &Service{
		repo:            repo,
		store:           store,
		cfg:             cfg,
		provider:        provider,
		summarizer:      NewSummarizer(provider),
		qaEngine:        NewQAEngine(provider, cb),
		promptExtractor: NewPromptExtractor(provider),
		ctxBuilder:      cb,
		docSaver:        docSaver,
	}
}

// EnqueueSummary creates a pending document_summary task.
func (s *Service) EnqueueSummary(ctx context.Context, documentID string) (*AITask, error) {
	inp, _ := json.Marshal(SummaryInput{DocumentID: documentID})
	return s.repo.CreateTask(TaskTypeDocumentSummary, string(inp))
}

// EnqueueQA creates a pending search_answer task.
func (s *Service) EnqueueQA(ctx context.Context, question string, documentIDs []string) (*AITask, error) {
	inp, _ := json.Marshal(QAInput{Question: question, DocumentIDs: documentIDs})
	return s.repo.CreateTask(TaskTypeSearchAnswer, string(inp))
}

// EnqueuePromptExtract creates a pending prompt_extract task.
func (s *Service) EnqueuePromptExtract(ctx context.Context, documentID string) (*AITask, error) {
	inp, _ := json.Marshal(PromptExtractInput{DocumentID: documentID})
	return s.repo.CreateTask(TaskTypePromptExtract, string(inp))
}

// GetTask returns a task by ID.
func (s *Service) GetTask(ctx context.Context, id string) (*AITask, error) {
	return s.repo.GetTask(id)
}

// ListTasks returns tasks filtered by status.
func (s *Service) ListTasks(ctx context.Context, status string, limit, offset int) ([]*AITask, int, error) {
	return s.repo.ListTasks(status, limit, offset)
}

// CancelTask cancels a pending task.
func (s *Service) CancelTask(ctx context.Context, id string) error {
	return s.repo.CancelTask(id)
}

// GetSummary returns the AI summary for a document.
func (s *Service) GetSummary(ctx context.Context, documentID string) (*DocumentAISummary, error) {
	return s.repo.GetSummary(documentID)
}

// ListPrompts returns the prompt library.
func (s *Service) ListPrompts(ctx context.Context, scenario string, limit, offset int) ([]*Prompt, int, error) {
	return s.repo.ListPrompts(scenario, limit, offset)
}

// GetPrompt returns a prompt by ID.
func (s *Service) GetPrompt(ctx context.Context, id string) (*Prompt, error) {
	return s.repo.GetPrompt(id)
}

// DeletePrompt removes a prompt from the library.
func (s *Service) DeletePrompt(ctx context.Context, id string) error {
	return s.repo.DeletePrompt(id)
}

// ProcessTask runs a single task synchronously. Called by the worker.
func (s *Service) ProcessTask(ctx context.Context, task *AITask) error {
	switch task.TaskType {
	case TaskTypeDocumentSummary:
		return s.processSummaryTask(ctx, task)
	case TaskTypeSearchAnswer:
		return s.processQATask(ctx, task)
	case TaskTypePromptExtract:
		return s.processPromptExtractTask(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.TaskType)
	}
}

func (s *Service) processSummaryTask(ctx context.Context, task *AITask) error {
	var inp SummaryInput
	if err := json.Unmarshal([]byte(task.InputJSON), &inp); err != nil {
		return fmt.Errorf("parse input: %w", err)
	}

	meta, err := s.repo.GetDocMeta(inp.DocumentID)
	if err != nil || meta == nil {
		return fmt.Errorf("document not found: %s", inp.DocumentID)
	}

	content, err := s.loadDocContent(ctx, *meta)
	if err != nil {
		return fmt.Errorf("load content: %w", err)
	}

	out, err := s.summarizer.Summarize(ctx, meta.Title, content)
	if err != nil {
		return err
	}

	// Store structured summary.
	kp, _ := json.Marshal(out.KeyPoints)
	ac, _ := json.Marshal(out.Actions)
	cr, _ := json.Marshal(out.CodeRefs)
	provName := s.provider.Name()
	model := s.cfg.Model
	sum := &DocumentAISummary{
		DocumentID:    inp.DocumentID,
		Summary:       out.Summary,
		KeyPointsJSON: strPtr(string(kp)),
		ActionsJSON:   strPtr(string(ac)),
		CodeRefsJSON:  strPtr(string(cr)),
		Provider:      &provName,
		Model:         &model,
	}
	if err := s.repo.UpsertSummary(sum); err != nil {
		return fmt.Errorf("upsert summary: %w", err)
	}

	// Save as Markdown document and record output.
	md := WriteSummaryMarkdown(meta.Title+" — AI Summary", provName, model, out, []string{inp.DocumentID})
	outDocID, err := s.docSaver.SaveAIDocument(ctx,
		meta.Title+" — AI Summary", md,
		"docs/ai/summaries/"+meta.ID+"-"+fmt.Sprintf("%d", time.Now().Unix())+".md",
		[]string{inp.DocumentID},
	)
	if err != nil {
		// Non-fatal: summary is already stored in document_ai_summaries.
		outDocID = ""
	}

	outJSON, _ := json.Marshal(out)
	var outDocPtr *string
	if outDocID != "" {
		outDocPtr = &outDocID
	}
	return s.repo.CompleteTask(task.ID, string(outJSON), provName, model, outDocPtr)
}

func (s *Service) processQATask(ctx context.Context, task *AITask) error {
	var inp QAInput
	if err := json.Unmarshal([]byte(task.InputJSON), &inp); err != nil {
		return fmt.Errorf("parse input: %w", err)
	}

	out, metas, err := s.qaEngine.Answer(ctx, inp.Question, inp.DocumentIDs)
	if err != nil {
		return err
	}

	provName := s.provider.Name()
	model := s.cfg.Model
	md := WriteQAMarkdown(inp.Question, provName, model, out)

	sourceIDs := make([]string, len(metas))
	for i, m := range metas {
		sourceIDs[i] = m.ID
	}

	outDocID, err := s.docSaver.SaveAIDocument(ctx,
		"Q&A: "+truncateTitle(inp.Question, 60), md,
		"docs/ai/qa/"+task.ID+".md",
		sourceIDs,
	)
	if err != nil {
		outDocID = ""
	}

	outJSON, _ := json.Marshal(out)
	var outDocPtr *string
	if outDocID != "" {
		outDocPtr = &outDocID
	}
	return s.repo.CompleteTask(task.ID, string(outJSON), provName, model, outDocPtr)
}

func (s *Service) processPromptExtractTask(ctx context.Context, task *AITask) error {
	var inp PromptExtractInput
	if err := json.Unmarshal([]byte(task.InputJSON), &inp); err != nil {
		return fmt.Errorf("parse input: %w", err)
	}

	meta, err := s.repo.GetDocMeta(inp.DocumentID)
	if err != nil || meta == nil {
		return fmt.Errorf("document not found: %s", inp.DocumentID)
	}

	content, err := s.loadDocContent(ctx, *meta)
	if err != nil {
		return fmt.Errorf("load content: %w", err)
	}

	out, err := s.promptExtractor.Extract(ctx, meta.Title, content)
	if err != nil {
		return err
	}

	provName := s.provider.Name()
	model := s.cfg.Model

	// Save to prompt library.
	p := &Prompt{
		Title:            out.Title,
		Content:          out.Content,
		SourceDocumentID: &inp.DocumentID,
		QualityScore:     out.QualityScore,
	}
	if out.ModelHint != "" {
		p.ModelHint = &out.ModelHint
	}
	if out.Scenario != "" {
		p.Scenario = &out.Scenario
	}
	if err := s.repo.CreatePrompt(p); err != nil {
		return fmt.Errorf("save prompt: %w", err)
	}

	// Also save as Markdown document.
	md := WritePromptMarkdown(out, inp.DocumentID, provName, model)
	outDocID, err := s.docSaver.SaveAIDocument(ctx,
		out.Title, md,
		"docs/ai/prompts/"+task.ID+".md",
		[]string{inp.DocumentID},
	)
	if err != nil {
		outDocID = ""
	}

	outJSON, _ := json.Marshal(out)
	var outDocPtr *string
	if outDocID != "" {
		outDocPtr = &outDocID
	}
	return s.repo.CompleteTask(task.ID, string(outJSON), provName, model, outDocPtr)
}

func (s *Service) loadDocContent(ctx context.Context, meta DocMeta) (string, error) {
	rc, err := s.store.Get(ctx, meta.StorageKey)
	if err != nil {
		return "", err
	}
	defer rc.Close()
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(rc); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ChunkDocument rebuilds document chunks and persists them.
func (s *Service) ChunkDocument(ctx context.Context, documentID, content string) error {
	chunks := chunkDocument(documentID, content)
	if err := s.repo.DeleteChunks(documentID); err != nil {
		return err
	}
	return s.repo.InsertChunks(chunks)
}

func strPtr(s string) *string { return &s }
