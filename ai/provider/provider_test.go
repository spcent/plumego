package provider

import (
	"context"
	"testing"
)

func TestNewTextMessage(t *testing.T) {
	msg := NewTextMessage(RoleUser, "Hello")

	if msg.Role != RoleUser {
		t.Errorf("Role = %v, want %v", msg.Role, RoleUser)
	}

	text := msg.GetText()
	if text != "Hello" {
		t.Errorf("GetText() = %v, want Hello", text)
	}
}

func TestNewToolResultMessage(t *testing.T) {
	msg := NewToolResultMessage("tool-123", "result", false)

	if msg.Role != RoleTool {
		t.Errorf("Role = %v, want %v", msg.Role, RoleTool)
	}

	if msg.ToolCallID != "tool-123" {
		t.Errorf("ToolCallID = %v, want tool-123", msg.ToolCallID)
	}

	blocks, ok := msg.Content.([]ContentBlock)
	if !ok || len(blocks) == 0 {
		t.Fatal("Content should be []ContentBlock with at least one block")
	}

	if blocks[0].Type != ContentTypeToolResult {
		t.Errorf("ContentBlock type = %v, want %v", blocks[0].Type, ContentTypeToolResult)
	}
}

func TestCompletionResponse_GetText(t *testing.T) {
	resp := &CompletionResponse{
		Content: []ContentBlock{
			{Type: ContentTypeText, Text: "Hello "},
			{Type: ContentTypeText, Text: "World"},
			{Type: ContentTypeToolUse, ToolUse: &ToolUse{Name: "test"}},
		},
	}

	text := resp.GetText()
	want := "Hello World"
	if text != want {
		t.Errorf("GetText() = %v, want %v", text, want)
	}
}

func TestCompletionResponse_GetToolUses(t *testing.T) {
	resp := &CompletionResponse{
		Content: []ContentBlock{
			{Type: ContentTypeText, Text: "Hello"},
			{
				Type: ContentTypeToolUse,
				ToolUse: &ToolUse{
					ID:   "tool-1",
					Name: "test_tool",
					Input: map[string]any{
						"arg": "value",
					},
				},
			},
			{
				Type: ContentTypeToolUse,
				ToolUse: &ToolUse{
					ID:   "tool-2",
					Name: "another_tool",
				},
			},
		},
	}

	toolUses := resp.GetToolUses()
	if len(toolUses) != 2 {
		t.Errorf("GetToolUses() count = %v, want 2", len(toolUses))
	}

	if toolUses[0].ID != "tool-1" {
		t.Errorf("First tool ID = %v, want tool-1", toolUses[0].ID)
	}

	if toolUses[1].ID != "tool-2" {
		t.Errorf("Second tool ID = %v, want tool-2", toolUses[1].ID)
	}
}

func TestCompletionResponse_HasToolUse(t *testing.T) {
	tests := []struct {
		name     string
		resp     *CompletionResponse
		expected bool
	}{
		{
			name: "has tool use",
			resp: &CompletionResponse{
				StopReason: StopReasonToolUse,
				Content: []ContentBlock{
					{
						Type:    ContentTypeToolUse,
						ToolUse: &ToolUse{ID: "tool-1"},
					},
				},
			},
			expected: true,
		},
		{
			name: "no tool use - wrong stop reason",
			resp: &CompletionResponse{
				StopReason: StopReasonEndTurn,
				Content: []ContentBlock{
					{
						Type:    ContentTypeToolUse,
						ToolUse: &ToolUse{ID: "tool-1"},
					},
				},
			},
			expected: false,
		},
		{
			name: "no tool use - no tool blocks",
			resp: &CompletionResponse{
				StopReason: StopReasonToolUse,
				Content: []ContentBlock{
					{Type: ContentTypeText, Text: "hello"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.HasToolUse(); got != tt.expected {
				t.Errorf("HasToolUse() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToolChoice(t *testing.T) {
	if ToolChoiceAuto.Type != "auto" {
		t.Errorf("ToolChoiceAuto.Type = %v, want auto", ToolChoiceAuto.Type)
	}

	if ToolChoiceNone.Type != "none" {
		t.Errorf("ToolChoiceNone.Type = %v, want none", ToolChoiceNone.Type)
	}

	if ToolChoiceAny.Type != "any" {
		t.Errorf("ToolChoiceAny.Type = %v, want any", ToolChoiceAny.Type)
	}

	toolChoice := ToolChoiceTool("my_tool")
	if toolChoice.Type != "tool" || toolChoice.Name != "my_tool" {
		t.Errorf("ToolChoiceTool() = %+v, want {Type:tool Name:my_tool}", toolChoice)
	}
}

func TestManager_Register(t *testing.T) {
	manager := NewManager()

	// Mock provider
	mockProvider := &mockProvider{name: "mock"}

	manager.Register(mockProvider)

	provider, err := manager.Get("mock")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if provider.Name() != "mock" {
		t.Errorf("Provider name = %v, want mock", provider.Name())
	}
}

func TestManager_Get_NotFound(t *testing.T) {
	manager := NewManager()

	_, err := manager.Get("nonexistent")
	if err == nil {
		t.Error("Get() should return error for nonexistent provider")
	}
}

func TestDefaultRouter_Route(t *testing.T) {
	router := &DefaultRouter{}

	mockClaude := &mockProvider{name: "claude"}
	mockOpenAI := &mockProvider{name: "openai"}

	providers := []Provider{mockClaude, mockOpenAI}

	tests := []struct {
		model    string
		expected string
	}{
		{"claude-3-opus", "claude"},
		{"gpt-4", "openai"},
		{"unknown-model", "claude"}, // fallback to first
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			req := &CompletionRequest{Model: tt.model}
			provider, err := router.Route(context.Background(), req, providers)
			if err != nil {
				t.Errorf("Route() error = %v", err)
			}

			if provider.Name() != tt.expected {
				t.Errorf("Route() provider = %v, want %v", provider.Name(), tt.expected)
			}
		})
	}
}

func TestLoadBalancerRouter_Route(t *testing.T) {
	router := &LoadBalancerRouter{}

	mock1 := &mockProvider{name: "provider1"}
	mock2 := &mockProvider{name: "provider2"}
	providers := []Provider{mock1, mock2}

	req := &CompletionRequest{Model: "test"}

	// First call should go to provider1
	provider1, err := router.Route(context.Background(), req, providers)
	if err != nil {
		t.Errorf("Route() error = %v", err)
	}

	// Second call should go to provider2
	provider2, err := router.Route(context.Background(), req, providers)
	if err != nil {
		t.Errorf("Route() error = %v", err)
	}

	if provider1.Name() == provider2.Name() {
		t.Error("LoadBalancerRouter should alternate between providers")
	}
}

func TestMatchesProvider(t *testing.T) {
	tests := []struct {
		model    string
		provider string
		expected bool
	}{
		{"claude-3-opus", "claude", true},
		{"claude-2.1", "claude", true},
		{"gpt-4", "openai", true},
		{"gpt-3.5-turbo", "openai", true},
		{"claude-3-opus", "openai", false},
		{"gpt-4", "claude", false},
	}

	for _, tt := range tests {
		t.Run(tt.model+"_"+tt.provider, func(t *testing.T) {
			if got := matchesProvider(tt.model, tt.provider); got != tt.expected {
				t.Errorf("matchesProvider(%v, %v) = %v, want %v", tt.model, tt.provider, got, tt.expected)
			}
		})
	}
}

// Mock provider for testing
type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{
		ID:    "test-id",
		Model: req.Model,
		Content: []ContentBlock{
			{Type: ContentTypeText, Text: "mock response"},
		},
	}, nil
}

func (m *mockProvider) CompleteStream(ctx context.Context, req *CompletionRequest) (*StreamReader, error) {
	return nil, nil
}

func (m *mockProvider) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{
		{ID: "mock-model", Name: "Mock Model", Provider: m.name},
	}, nil
}

func (m *mockProvider) GetModel(ctx context.Context, modelID string) (*Model, error) {
	return &Model{ID: modelID, Name: "Mock Model", Provider: m.name}, nil
}

func (m *mockProvider) CountTokens(text string) (int, error) {
	return len(text) / 4, nil
}
