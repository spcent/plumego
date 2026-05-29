package ai

import (
	"context"
	"fmt"
	"strings"
)

// MockProvider returns deterministic stub responses for offline development.
type MockProvider struct{}

func NewMockProvider() *MockProvider { return &MockProvider{} }

func (p *MockProvider) Name() string { return "local_mock" }

func (p *MockProvider) Chat(_ context.Context, req ChatRequest) (*ChatResponse, error) {
	lastUser := ""
	for _, m := range req.Messages {
		if m.Role == "user" {
			lastUser = m.Content
		}
	}

	var reply string
	switch {
	case strings.Contains(lastUser, "summarize") || strings.Contains(lastUser, "summary"):
		reply = buildMockSummary()
	case strings.Contains(lastUser, "answer") || strings.Contains(lastUser, "question"):
		reply = buildMockAnswer()
	case strings.Contains(lastUser, "prompt") || strings.Contains(lastUser, "extract"):
		reply = buildMockPromptExtract()
	default:
		reply = fmt.Sprintf("[mock] Received %d messages. This is a stub response from the local mock provider.", len(req.Messages))
	}

	return &ChatResponse{
		Content:      reply,
		PromptTokens: estimateTokens(lastUser),
		ReplyTokens:  estimateTokens(reply),
		Model:        "local_mock",
	}, nil
}

func buildMockSummary() string {
	return `## Summary

This document covers the main topic with several key points.

## Key Points

- Key point one describing an important concept
- Key point two with actionable information
- Key point three highlighting a technical detail

## Actions

- Review the main sections for completeness
- Consider cross-referencing related documents`
}

func buildMockAnswer() string {
	return `Based on the provided documents, the answer involves the following:

The relevant information found in the selected documents suggests that the topic requires careful consideration of multiple factors. The primary source document provides the core context, while supporting documents add supplementary details.

**Sources used:** The answer is derived from the explicitly selected documents only.`
}

func buildMockPromptExtract() string {
	return `## Extracted Prompt

You are an expert assistant. Given the following context, help the user accomplish their goal.

**Scenario:** Technical documentation review
**Model hint:** gpt-4o
**Quality score:** 0.75`
}

func estimateTokens(s string) int {
	// rough 4-chars-per-token heuristic
	return len(s)/4 + 1
}
