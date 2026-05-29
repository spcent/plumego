package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const systemSummarize = `You are a technical document analyst. Given a Markdown document, produce a structured analysis. Respond in JSON:
{
  "summary": "2-4 sentence plain text summary",
  "key_points": ["...", "..."],
  "actions": ["action item 1", "action item 2"],
  "code_refs": ["language or snippet description"]
}
Be concise. Do not invent information not in the document.`

type Summarizer struct {
	provider LLMProvider
}

func NewSummarizer(provider LLMProvider) *Summarizer {
	return &Summarizer{provider: provider}
}

func (s *Summarizer) Summarize(ctx context.Context, title, content string) (*SummaryOutput, error) {
	userMsg := fmt.Sprintf("Document title: %s\n\n---\n\n%s\n\n---\n\nPlease summarize this document.", title, content)

	resp, err := s.provider.Chat(ctx, ChatRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: systemSummarize},
			{Role: "user", Content: userMsg},
		},
		MaxTokens:   800,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("summarize call: %w", err)
	}

	var out SummaryOutput
	cleaned := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(cleaned), &out); err != nil {
		// Graceful fallback: put raw content as summary.
		out.Summary = resp.Content
	}
	return &out, nil
}

const systemQA = `You are a knowledge assistant. Answer the user's question strictly based on the provided documents.
If the answer is not in the documents, say "I could not find an answer in the selected documents."
Always cite which document(s) your answer comes from.
Respond in JSON:
{
  "answer": "your answer here",
  "citations": [
    {"document_id": "...", "document_title": "...", "excerpt": "short quote from document"}
  ]
}`

type QAEngine struct {
	provider       LLMProvider
	contextBuilder *ContextBuilder
}

func NewQAEngine(provider LLMProvider, cb *ContextBuilder) *QAEngine {
	return &QAEngine{provider: provider, contextBuilder: cb}
}

func (e *QAEngine) Answer(ctx context.Context, question string, documentIDs []string) (*QAOutput, []DocMeta, error) {
	docContext, metas, err := e.contextBuilder.BuildContext(ctx, documentIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("build context: %w", err)
	}

	userMsg := fmt.Sprintf("Documents:\n\n%s\n\nQuestion: %s", docContext, question)

	resp, err := e.provider.Chat(ctx, ChatRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: systemQA},
			{Role: "user", Content: userMsg},
		},
		MaxTokens:   1000,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, metas, fmt.Errorf("qa call: %w", err)
	}

	var out QAOutput
	cleaned := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(cleaned), &out); err != nil {
		out.Answer = resp.Content
	}

	ids := make([]string, len(metas))
	for i, m := range metas {
		ids[i] = m.ID
	}
	out.DocumentIDs = ids
	return &out, metas, nil
}

const systemPromptExtract = `You are a prompt engineering expert. Given a document, extract or compose a reusable LLM prompt from its content.
Respond in JSON:
{
  "title": "short title for the prompt",
  "content": "the full prompt text",
  "model_hint": "recommended model (e.g. gpt-4o, claude-3)",
  "scenario": "use case category (e.g. summarization, code_review, qa)",
  "quality_score": 0.0
}
quality_score is 0.0-1.0 reflecting how useful and reusable the prompt is.`

type PromptExtractor struct {
	provider LLMProvider
}

func NewPromptExtractor(provider LLMProvider) *PromptExtractor {
	return &PromptExtractor{provider: provider}
}

func (e *PromptExtractor) Extract(ctx context.Context, title, content string) (*PromptExtractOutput, error) {
	userMsg := fmt.Sprintf("Document: %s\n\n%s\n\nExtract a reusable prompt from this document.", title, content)

	resp, err := e.provider.Chat(ctx, ChatRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: systemPromptExtract},
			{Role: "user", Content: userMsg},
		},
		MaxTokens:   600,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("prompt extract call: %w", err)
	}

	var out PromptExtractOutput
	cleaned := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(cleaned), &out); err != nil {
		out.Title = title + " (prompt)"
		out.Content = resp.Content
	}
	return &out, nil
}

// extractJSON tries to find the first {...} block in a string.
func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end < start {
		return s
	}
	return s[start : end+1]
}
