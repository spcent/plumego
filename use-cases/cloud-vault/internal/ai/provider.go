package ai

import "context"

// ChatMessage is a single turn in a conversation.
type ChatMessage struct {
	Role    string // "system" | "user" | "assistant"
	Content string
}

// ChatRequest is the input to LLMProvider.Chat.
type ChatRequest struct {
	Messages    []ChatMessage
	MaxTokens   int
	Temperature float64
}

// ChatResponse is the output from LLMProvider.Chat.
type ChatResponse struct {
	Content      string
	PromptTokens int
	ReplyTokens  int
	Model        string
}

// LLMProvider is the single abstraction over AI backends.
type LLMProvider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
