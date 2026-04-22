package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- ClaudeProvider ---

func TestNewClaudeProvider_Defaults(t *testing.T) {
	p := NewClaudeProvider("key-123")
	if p.Name() != "claude" {
		t.Errorf("Name() = %q, want claude", p.Name())
	}
	if p.apiKey != "key-123" {
		t.Errorf("apiKey = %q, want key-123", p.apiKey)
	}
	if p.baseURL != claudeAPIBaseURL {
		t.Errorf("baseURL = %q, want %q", p.baseURL, claudeAPIBaseURL)
	}
	if p.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestWithClaudeBaseURL(t *testing.T) {
	p := NewClaudeProvider("k", WithClaudeBaseURL("http://custom:1234"))
	if p.baseURL != "http://custom:1234" {
		t.Errorf("baseURL = %q, want http://custom:1234", p.baseURL)
	}
}

func TestWithClaudeHTTPClient(t *testing.T) {
	custom := &http.Client{}
	p := NewClaudeProvider("k", WithClaudeHTTPClient(custom))
	if p.httpClient != custom {
		t.Error("httpClient was not replaced by WithClaudeHTTPClient")
	}
}

func TestClaudeProvider_Complete_OK(t *testing.T) {
	resp := map[string]any{
		"id":          "msg-1",
		"type":        "message",
		"role":        "assistant",
		"model":       "claude-3-opus",
		"stop_reason": "end_turn",
		"content": []map[string]any{
			{"type": "text", "text": "Hello from Claude"},
		},
		"usage": map[string]any{"input_tokens": 10, "output_tokens": 5},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			t.Error("missing x-api-key header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := NewClaudeProvider("test-key", WithClaudeBaseURL(srv.URL))
	result, err := p.Complete(t.Context(), &CompletionRequest{
		Model:     "claude-3-opus",
		Messages:  []Message{NewTextMessage(RoleUser, "hi")},
		MaxTokens: 100,
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if result.GetText() != "Hello from Claude" {
		t.Errorf("text = %q, want 'Hello from Claude'", result.GetText())
	}
}

func TestClaudeProvider_Complete_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid key"}`))
	}))
	defer srv.Close()

	p := NewClaudeProvider("bad-key", WithClaudeBaseURL(srv.URL))
	_, err := p.Complete(t.Context(), &CompletionRequest{
		Model:    "claude-3",
		Messages: []Message{NewTextMessage(RoleUser, "hi")},
	})
	if err == nil {
		t.Error("expected error for 401 response")
	}
}

func TestClaudeProvider_ImplementsProvider(t *testing.T) {
	var _ Provider = (*ClaudeProvider)(nil)
}

// --- OpenAIProvider ---

func TestNewOpenAIProvider_Defaults(t *testing.T) {
	p := NewOpenAIProvider("sk-123")
	if p.Name() != "openai" {
		t.Errorf("Name() = %q, want openai", p.Name())
	}
	if p.apiKey != "sk-123" {
		t.Errorf("apiKey = %q, want sk-123", p.apiKey)
	}
	if p.baseURL != openaiAPIBaseURL {
		t.Errorf("baseURL = %q, want %q", p.baseURL, openaiAPIBaseURL)
	}
	if p.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestWithOpenAIBaseURL(t *testing.T) {
	p := NewOpenAIProvider("k", WithOpenAIBaseURL("http://custom:5678"))
	if p.baseURL != "http://custom:5678" {
		t.Errorf("baseURL = %q, want http://custom:5678", p.baseURL)
	}
}

func TestWithOpenAIHTTPClient(t *testing.T) {
	custom := &http.Client{}
	p := NewOpenAIProvider("k", WithOpenAIHTTPClient(custom))
	if p.httpClient != custom {
		t.Error("httpClient was not replaced by WithOpenAIHTTPClient")
	}
}

func TestOpenAIProvider_Complete_OK(t *testing.T) {
	resp := map[string]any{
		"id":    "chatcmpl-1",
		"model": "gpt-4o",
		"choices": []map[string]any{
			{
				"index":         0,
				"finish_reason": "stop",
				"message": map[string]any{
					"role":    "assistant",
					"content": "Hello from OpenAI",
				},
			},
		},
		"usage": map[string]any{
			"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15,
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := NewOpenAIProvider("sk-test", WithOpenAIBaseURL(srv.URL))
	result, err := p.Complete(t.Context(), &CompletionRequest{
		Model:     "gpt-4o",
		Messages:  []Message{NewTextMessage(RoleUser, "hi")},
		MaxTokens: 100,
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if result.GetText() != "Hello from OpenAI" {
		t.Errorf("text = %q, want 'Hello from OpenAI'", result.GetText())
	}
}

func TestOpenAIProvider_Complete_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer srv.Close()

	p := NewOpenAIProvider("sk-bad", WithOpenAIBaseURL(srv.URL))
	_, err := p.Complete(t.Context(), &CompletionRequest{
		Model:    "gpt-4o",
		Messages: []Message{NewTextMessage(RoleUser, "hi")},
	})
	if err == nil {
		t.Error("expected error for 429 response")
	}
}

func TestOpenAIProvider_ImplementsProvider(t *testing.T) {
	var _ Provider = (*OpenAIProvider)(nil)
}
