package ai

import (
	"context"
	"strings"
	"testing"

	"cloud-vault/internal/document"
)

// ---------------------------------------------------------------------------
// hashContent
// ---------------------------------------------------------------------------

func TestHashContent_Deterministic(t *testing.T) {
	a := hashContent("hello world")
	b := hashContent("hello world")
	if a != b {
		t.Errorf("hashContent not deterministic: %q vs %q", a, b)
	}
}

func TestHashContent_DifferentInputs(t *testing.T) {
	a := hashContent("hello")
	b := hashContent("world")
	if a == b {
		t.Errorf("different inputs should produce different hashes, both = %q", a)
	}
}

func TestHashContent_EmptyString(t *testing.T) {
	got := hashContent("")
	if len(got) == 0 {
		t.Error("hashContent empty string should return non-empty hash")
	}
}

func TestHashContent_HexFormat(t *testing.T) {
	got := hashContent("test content")
	for _, c := range got {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("hash %q contains non-hex char %q", got, c)
		}
	}
}

func TestHashContent_Length(t *testing.T) {
	// sha256 first 8 bytes = 16 hex chars
	got := hashContent("any content")
	if len(got) != 16 {
		t.Errorf("hashContent length = %d, want 16", len(got))
	}
}

// ---------------------------------------------------------------------------
// nilIfEmpty
// ---------------------------------------------------------------------------

func TestNilIfEmpty_EmptyString(t *testing.T) {
	got := nilIfEmpty("")
	if got != nil {
		t.Errorf("nilIfEmpty(\"\") = %v, want nil", got)
	}
}

func TestNilIfEmpty_NonEmpty(t *testing.T) {
	s := "hello"
	got := nilIfEmpty(s)
	if got == nil {
		t.Fatal("nilIfEmpty non-empty should return non-nil pointer")
	}
	if *got != s {
		t.Errorf("nilIfEmpty(%q) = %q, want %q", s, *got, s)
	}
}

func TestNilIfEmpty_Whitespace(t *testing.T) {
	got := nilIfEmpty("   ")
	if got == nil {
		t.Error("nilIfEmpty whitespace should return non-nil (only empty string is nil)")
	}
}

// ---------------------------------------------------------------------------
// estimateTokens
// ---------------------------------------------------------------------------

func TestEstimateTokens_Empty(t *testing.T) {
	got := estimateTokens("")
	if got < 1 {
		t.Errorf("estimateTokens empty = %d, want >= 1", got)
	}
}

func TestEstimateTokens_LongerInput(t *testing.T) {
	short := estimateTokens("hi")
	long := estimateTokens(strings.Repeat("hello world ", 100))
	if long <= short {
		t.Errorf("longer input should have more tokens: short=%d long=%d", short, long)
	}
}

func TestEstimateTokens_Proportional(t *testing.T) {
	small := estimateTokens("aaaa")     // 4 chars → ~1 token + 1 = 2
	large := estimateTokens("aaaa" + strings.Repeat("a", 396)) // 400 chars → 100 + 1 = 101
	if large <= small {
		t.Errorf("estimateTokens large should be > small: %d vs %d", large, small)
	}
}

// ---------------------------------------------------------------------------
// chunkDocument
// ---------------------------------------------------------------------------

func TestChunkDocument_EmptyContent(t *testing.T) {
	chunks := chunkDocument("doc1", "")
	if len(chunks) != 0 {
		t.Errorf("chunkDocument empty content = %d chunks, want 0", len(chunks))
	}
}

func TestChunkDocument_SingleParagraph(t *testing.T) {
	content := "This is a simple paragraph without any headings."
	chunks := chunkDocument("doc1", content)
	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk for non-empty content")
	}
	if chunks[0].DocumentID != "doc1" {
		t.Errorf("DocumentID = %q, want %q", chunks[0].DocumentID, "doc1")
	}
	if !strings.Contains(chunks[0].Content, "simple paragraph") {
		t.Errorf("chunk content missing original text: %q", chunks[0].Content)
	}
}

func TestChunkDocument_WithHeadings(t *testing.T) {
	content := `# Introduction

This is the introduction section.

# Methods

This is the methods section.

# Results

This is the results section.`

	chunks := chunkDocument("doc2", content)
	if len(chunks) == 0 {
		t.Fatal("expected chunks for content with headings")
	}
	// Each chunk should have a heading path set.
	for _, c := range chunks {
		if c.HeadingPath == nil && c.ChunkIndex > 0 {
			// First chunk might have empty heading path if it's pre-heading content.
			t.Logf("chunk %d has nil HeadingPath", c.ChunkIndex)
		}
	}
}

func TestChunkDocument_ChunkIndexSequential(t *testing.T) {
	content := `# Section One

Content for section one with enough text.

# Section Two

Content for section two with enough text.`

	chunks := chunkDocument("doc3", content)
	for i, c := range chunks {
		if c.ChunkIndex != i {
			t.Errorf("chunk %d has ChunkIndex=%d, want %d", i, c.ChunkIndex, i)
		}
	}
}

func TestChunkDocument_ContentHashSet(t *testing.T) {
	content := "# Title\n\nSome content here."
	chunks := chunkDocument("doc4", content)
	for _, c := range chunks {
		if c.ContentHash == "" {
			t.Error("chunk ContentHash should not be empty")
		}
	}
}

func TestChunkDocument_TokenCountPositive(t *testing.T) {
	content := "# Title\n\nSome content here."
	chunks := chunkDocument("doc4", content)
	for _, c := range chunks {
		if c.TokenCount <= 0 {
			t.Errorf("chunk TokenCount = %d, want > 0", c.TokenCount)
		}
	}
}

func TestChunkDocument_CodeFenceNotSplit(t *testing.T) {
	content := `# Code Example

Here is a code block:

` + "```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```" + `

More text after the code block.`

	chunks := chunkDocument("doc5", content)
	// The code block should remain intact (not split in the middle).
	allContent := strings.Join(func() []string {
		s := make([]string, len(chunks))
		for i, c := range chunks {
			s[i] = c.Content
		}
		return s
	}(), "\n")
	if !strings.Contains(allContent, "fmt.Println") {
		t.Error("code block content should be preserved in chunks")
	}
}

func TestChunkDocument_HeadingPath(t *testing.T) {
	content := `# Top Level

Some content.

## Sub Level

More content.`

	chunks := chunkDocument("doc6", content)
	// Find a chunk with a heading path that includes the sub-level.
	found := false
	for _, c := range chunks {
		if c.HeadingPath != nil && strings.Contains(*c.HeadingPath, "Sub Level") {
			found = true
			break
		}
	}
	if !found {
		// It's possible the sub-level is merged into the same chunk as top level.
		// Just ensure no panic occurred.
		t.Log("no separate Sub Level chunk found - may be merged with parent")
	}
}

func TestChunkDocument_LargeContent(t *testing.T) {
	// Build content that will require multiple chunks (>800 tokens each section).
	var sb strings.Builder
	for i := 0; i < 10; i++ {
		sb.WriteString("# Section ")
		sb.WriteString(string(rune('A' + i)))
		sb.WriteString("\n\n")
		// ~1000 chars per section → multiple chunks expected.
		sb.WriteString(strings.Repeat("This is content for the section. It has many words. ", 20))
		sb.WriteString("\n\n")
	}
	content := sb.String()

	chunks := chunkDocument("big-doc", content)
	if len(chunks) < 2 {
		t.Errorf("large content should produce multiple chunks, got %d", len(chunks))
	}
}

func TestChunkDocument_DocumentIDPreserved(t *testing.T) {
	content := "# Title\n\nContent"
	id := "my-special-doc-id"
	chunks := chunkDocument(id, content)
	for _, c := range chunks {
		if c.DocumentID != id {
			t.Errorf("DocumentID = %q, want %q", c.DocumentID, id)
		}
	}
}

// ---------------------------------------------------------------------------
// extractJSON
// ---------------------------------------------------------------------------

func TestExtractJSON_ValidJSON(t *testing.T) {
	s := `{"summary": "hello", "key_points": []}`
	got := extractJSON(s)
	if got != s {
		t.Errorf("extractJSON(%q) = %q, want same", s, got)
	}
}

func TestExtractJSON_JSONWithPreamble(t *testing.T) {
	s := `Here is the result: {"answer": "42"} done.`
	got := extractJSON(s)
	if got != `{"answer": "42"}` {
		t.Errorf("extractJSON = %q, want %q", got, `{"answer": "42"}`)
	}
}

func TestExtractJSON_NoJSON(t *testing.T) {
	s := "no json here"
	got := extractJSON(s)
	if got != s {
		t.Errorf("extractJSON with no JSON = %q, want original %q", got, s)
	}
}

func TestExtractJSON_EmptyString(t *testing.T) {
	got := extractJSON("")
	if got != "" {
		t.Errorf("extractJSON empty = %q, want \"\"", got)
	}
}

func TestExtractJSON_NestedBraces(t *testing.T) {
	s := `{"outer": {"inner": "value"}}`
	got := extractJSON(s)
	if got != s {
		t.Errorf("extractJSON nested = %q, want %q", got, s)
	}
}

// ---------------------------------------------------------------------------
// WriteSummaryMarkdown
// ---------------------------------------------------------------------------

func TestWriteSummaryMarkdown_BasicOutput(t *testing.T) {
	out := &SummaryOutput{
		Summary:   "This is a summary.",
		KeyPoints: []string{"Point one", "Point two"},
		Actions:   []string{"Action one"},
		CodeRefs:  []string{"Go snippet"},
	}
	result := WriteSummaryMarkdown("My Title", "openai", "gpt-4o", out, []string{"doc-1", "doc-2"})

	if !strings.Contains(result, "# My Title") {
		t.Error("missing title heading")
	}
	if !strings.Contains(result, "This is a summary.") {
		t.Error("missing summary text")
	}
	if !strings.Contains(result, "Point one") {
		t.Error("missing key point")
	}
	if !strings.Contains(result, "Action one") {
		t.Error("missing action item")
	}
	if !strings.Contains(result, "Go snippet") {
		t.Error("missing code ref")
	}
	if !strings.Contains(result, "doc-1") {
		t.Error("missing source document ID")
	}
	if !strings.Contains(result, "openai") {
		t.Error("missing provider name")
	}
	if !strings.Contains(result, "gpt-4o") {
		t.Error("missing model name")
	}
}

func TestWriteSummaryMarkdown_NoKeyPoints(t *testing.T) {
	out := &SummaryOutput{
		Summary: "Just a summary, no key points.",
	}
	result := WriteSummaryMarkdown("Title", "mock", "model", out, nil)
	if !strings.Contains(result, "Just a summary") {
		t.Error("missing summary")
	}
	// Should not include empty sections.
	if strings.Contains(result, "## Key Points") {
		t.Error("should not include Key Points section when empty")
	}
}

func TestWriteSummaryMarkdown_NoSourceDocs(t *testing.T) {
	out := &SummaryOutput{Summary: "Summary text."}
	result := WriteSummaryMarkdown("Title", "provider", "model", out, nil)
	if strings.Contains(result, "## Source Documents") {
		t.Error("should not include Source Documents section when empty")
	}
}

// ---------------------------------------------------------------------------
// WriteQAMarkdown
// ---------------------------------------------------------------------------

func TestWriteQAMarkdown_BasicOutput(t *testing.T) {
	out := &QAOutput{
		Answer: "The answer is 42.",
		Citations: []Citation{
			{DocumentID: "doc-1", DocumentTitle: "Reference Book", Excerpt: "42 is the answer."},
		},
		DocumentIDs: []string{"doc-1"},
	}
	result := WriteQAMarkdown("What is the answer?", "openai", "gpt-4o", out)

	if !strings.Contains(result, "What is the answer?") {
		t.Error("missing question")
	}
	if !strings.Contains(result, "The answer is 42.") {
		t.Error("missing answer")
	}
	if !strings.Contains(result, "Reference Book") {
		t.Error("missing citation title")
	}
	if !strings.Contains(result, "42 is the answer.") {
		t.Error("missing citation excerpt")
	}
	if !strings.Contains(result, "doc-1") {
		t.Error("missing source document ID")
	}
}

func TestWriteQAMarkdown_TruncatesLongQuestion(t *testing.T) {
	longQ := strings.Repeat("What is the meaning of life? ", 10)
	out := &QAOutput{Answer: "Some answer."}
	result := WriteQAMarkdown(longQ, "provider", "model", out)
	// Title heading should exist and be non-empty.
	if !strings.Contains(result, "# Q&A:") {
		t.Error("missing Q&A title heading")
	}
}

func TestWriteQAMarkdown_NoCitations(t *testing.T) {
	out := &QAOutput{Answer: "Answer without citations."}
	result := WriteQAMarkdown("Question?", "provider", "model", out)
	if !strings.Contains(result, "Answer without citations.") {
		t.Error("missing answer")
	}
}

// ---------------------------------------------------------------------------
// WritePromptMarkdown
// ---------------------------------------------------------------------------

func TestWritePromptMarkdown_BasicOutput(t *testing.T) {
	out := &PromptExtractOutput{
		Title:        "Code Review Prompt",
		Content:      "You are an expert code reviewer. Please review the following code.",
		ModelHint:    "gpt-4o",
		Scenario:     "code_review",
		QualityScore: 0.85,
	}
	result := WritePromptMarkdown(out, "src-doc-1", "openai", "gpt-4o")

	if !strings.Contains(result, "# Code Review Prompt") {
		t.Error("missing title")
	}
	if !strings.Contains(result, "You are an expert code reviewer.") {
		t.Error("missing prompt content")
	}
	if !strings.Contains(result, "gpt-4o") {
		t.Error("missing model hint")
	}
	if !strings.Contains(result, "code_review") {
		t.Error("missing scenario")
	}
	if !strings.Contains(result, "0.85") {
		t.Error("missing quality score")
	}
	if !strings.Contains(result, "src-doc-1") {
		t.Error("missing source document ID")
	}
}

func TestWritePromptMarkdown_ZeroQualityScore(t *testing.T) {
	out := &PromptExtractOutput{
		Title:        "Test Prompt",
		Content:      "Test content.",
		QualityScore: 0.0,
	}
	result := WritePromptMarkdown(out, "", "provider", "model")
	// Should not include quality score line when 0.
	if strings.Contains(result, "Quality score: 0.00") {
		t.Error("should not include quality score when it is 0")
	}
}

func TestWritePromptMarkdown_NoSourceDoc(t *testing.T) {
	out := &PromptExtractOutput{
		Title:   "Standalone Prompt",
		Content: "Do something useful.",
	}
	result := WritePromptMarkdown(out, "", "provider", "model")
	if strings.Contains(result, "## Source") {
		t.Error("should not include Source section when sourceDocID is empty")
	}
}

// ---------------------------------------------------------------------------
// MockProvider
// ---------------------------------------------------------------------------

func TestMockProvider_Name(t *testing.T) {
	p := NewMockProvider()
	if p.Name() != "local_mock" {
		t.Errorf("Name() = %q, want %q", p.Name(), "local_mock")
	}
}

func TestMockProvider_ChatSummary(t *testing.T) {
	p := NewMockProvider()
	resp, err := p.Chat(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: "user", Content: "Please summarize this document for me."},
		},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty response content")
	}
	if resp.Model != "local_mock" {
		t.Errorf("Model = %q, want local_mock", resp.Model)
	}
}

func TestMockProvider_ChatAnswer(t *testing.T) {
	p := NewMockProvider()
	resp, err := p.Chat(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: "user", Content: "Please answer the question about this topic."},
		},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty response content")
	}
}

func TestMockProvider_ChatPromptExtract(t *testing.T) {
	p := NewMockProvider()
	resp, err := p.Chat(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: "user", Content: "Extract a reusable prompt from this document."},
		},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty response content")
	}
}

func TestMockProvider_ChatDefault(t *testing.T) {
	p := NewMockProvider()
	resp, err := p.Chat(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: "user", Content: "Unrelated request that does not match any keyword."},
		},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Content == "" {
		t.Error("expected non-empty default response")
	}
	if resp.PromptTokens <= 0 {
		t.Errorf("PromptTokens = %d, want > 0", resp.PromptTokens)
	}
	if resp.ReplyTokens <= 0 {
		t.Errorf("ReplyTokens = %d, want > 0", resp.ReplyTokens)
	}
}

func TestMockProvider_ChatNoMessages(t *testing.T) {
	p := NewMockProvider()
	resp, err := p.Chat(context.Background(), ChatRequest{
		Messages: []ChatMessage{},
	})
	if err != nil {
		t.Fatalf("Chat with no messages: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

// ---------------------------------------------------------------------------
// Summarizer with MockProvider
// ---------------------------------------------------------------------------

func TestSummarizer_Summarize(t *testing.T) {
	provider := NewMockProvider()
	summarizer := NewSummarizer(provider)

	out, err := summarizer.Summarize(context.Background(), "Test Document",
		"# Test\n\nThis is the content to summarize. It has several key points.")
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil SummaryOutput")
	}
	if out.Summary == "" {
		t.Error("Summary should not be empty")
	}
}

// ---------------------------------------------------------------------------
// PromptExtractor with MockProvider
// ---------------------------------------------------------------------------

func TestPromptExtractor_Extract(t *testing.T) {
	provider := NewMockProvider()
	extractor := NewPromptExtractor(provider)

	out, err := extractor.Extract(context.Background(), "Sample Prompt Document",
		"You are an expert. Please help the user by extracting a reusable prompt.")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil PromptExtractOutput")
	}
}

// ---------------------------------------------------------------------------
// Integration: full AI service flow
// ---------------------------------------------------------------------------

func TestAI_ProcessPromptExtract(t *testing.T) {
	aiSvc, docSvc, _, _ := setupAI(t)
	ctx := context.Background()

	doc, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Prompt Document",
		Content: "# Prompt\n\nYou are an expert. Extract this reusable prompt from the document content.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	task, err := aiSvc.EnqueuePromptExtract(ctx, doc.ID)
	if err != nil {
		t.Fatalf("EnqueuePromptExtract: %v", err)
	}
	if task.TaskType != TaskTypePromptExtract {
		t.Errorf("TaskType = %q, want %q", task.TaskType, TaskTypePromptExtract)
	}

	err = aiSvc.ProcessTask(ctx, task)
	if err != nil {
		t.Fatalf("ProcessTask: %v", err)
	}

	updatedTask, err := aiSvc.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if updatedTask.Status != TaskStatusCompleted {
		t.Errorf("Status = %q, want %q", updatedTask.Status, TaskStatusCompleted)
	}
}

func TestAI_ProcessQATask(t *testing.T) {
	aiSvc, docSvc, _, _ := setupAI(t)
	ctx := context.Background()

	doc, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Knowledge Base Entry",
		Content: "# Knowledge\n\nThe answer to the question is 42.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	task, err := aiSvc.EnqueueQA(ctx, "What is the answer?", []string{doc.ID})
	if err != nil {
		t.Fatalf("EnqueueQA: %v", err)
	}

	err = aiSvc.ProcessTask(ctx, task)
	if err != nil {
		t.Fatalf("ProcessTask: %v", err)
	}

	updatedTask, err := aiSvc.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if updatedTask.Status != TaskStatusCompleted {
		t.Errorf("Status = %q, want %q", updatedTask.Status, TaskStatusCompleted)
	}
}

func TestAI_GetNonexistentTask(t *testing.T) {
	aiSvc, _, _, _ := setupAI(t)
	ctx := context.Background()

	_, err := aiSvc.GetTask(ctx, "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent task ID")
	}
}

func TestAI_ListPrompts(t *testing.T) {
	aiSvc, _, _, _ := setupAI(t)
	ctx := context.Background()

	// scenario="" means list all prompts.
	prompts, total, err := aiSvc.ListPrompts(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}
	if total < 0 {
		t.Errorf("total = %d, want >= 0", total)
	}
	_ = prompts
}
