package tenant

import (
	"context"
	"testing"
)

func TestConfigPolicyEvaluator_Allow(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Policy: PolicyConfig{
			AllowedModels: []string{"gpt-4", "gpt-3.5-turbo"},
			AllowedTools:  []string{"search", "calculator", "code_interpreter"},
		},
	})

	evaluator := NewConfigPolicyEvaluator(mgr)
	ctx := context.Background()

	// Test allowed model
	result, err := evaluator.Evaluate(ctx, "test-tenant", PolicyRequest{
		Model: "gpt-4",
		Tool:  "search",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Errorf("expected allowed, got denied")
	}
	if result.Reason != "" {
		t.Errorf("expected empty reason for allowed request, got %s", result.Reason)
	}

	// Test another allowed combination
	result2, err2 := evaluator.Evaluate(ctx, "test-tenant", PolicyRequest{
		Model: "gpt-3.5-turbo",
		Tool:  "code_interpreter",
	})

	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if !result2.Allowed {
		t.Errorf("expected allowed for gpt-3.5-turbo + code_interpreter")
	}
}

func TestConfigPolicyEvaluator_DenyModel(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "restricted-tenant",
		Policy: PolicyConfig{
			AllowedModels: []string{"gpt-3.5-turbo"},
			AllowedTools:  []string{"search"},
		},
	})

	evaluator := NewConfigPolicyEvaluator(mgr)
	ctx := context.Background()

	// Test disallowed model
	result, err := evaluator.Evaluate(ctx, "restricted-tenant", PolicyRequest{
		Model: "gpt-4",
		Tool:  "search",
	})

	if err != ErrPolicyDenied {
		t.Errorf("expected ErrPolicyDenied, got %v", err)
	}
	if result.Allowed {
		t.Errorf("expected denied, got allowed")
	}
	if result.Reason != "model not allowed" {
		t.Errorf("expected 'model not allowed', got %s", result.Reason)
	}
}

func TestConfigPolicyEvaluator_DenyTool(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "limited-tools",
		Policy: PolicyConfig{
			AllowedModels: []string{"gpt-4"},
			AllowedTools:  []string{"search"},
		},
	})

	evaluator := NewConfigPolicyEvaluator(mgr)
	ctx := context.Background()

	// Test disallowed tool
	result, err := evaluator.Evaluate(ctx, "limited-tools", PolicyRequest{
		Model: "gpt-4",
		Tool:  "code_interpreter",
	})

	if err != ErrPolicyDenied {
		t.Errorf("expected ErrPolicyDenied, got %v", err)
	}
	if result.Allowed {
		t.Errorf("expected denied, got allowed")
	}
	if result.Reason != "tool not allowed" {
		t.Errorf("expected 'tool not allowed', got %s", result.Reason)
	}
}

func TestConfigPolicyEvaluator_EmptyListAllowsAll(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "open-tenant",
		Policy: PolicyConfig{
			AllowedModels: []string{}, // Empty = allow all
			AllowedTools:  []string{}, // Empty = allow all
		},
	})

	evaluator := NewConfigPolicyEvaluator(mgr)
	ctx := context.Background()

	tests := []struct {
		name  string
		model string
		tool  string
	}{
		{"any model/tool", "gpt-4", "search"},
		{"different model", "claude-3", "calculator"},
		{"another combo", "gemini-pro", "code_interpreter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(ctx, "open-tenant", PolicyRequest{
				Model: tt.model,
				Tool:  tt.tool,
			})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !result.Allowed {
				t.Errorf("expected allowed (empty list allows all)")
			}
		})
	}
}

func TestConfigPolicyEvaluator_EmptyValues(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Policy: PolicyConfig{
			AllowedModels: []string{"gpt-4"},
			AllowedTools:  []string{"search"},
		},
	})

	evaluator := NewConfigPolicyEvaluator(mgr)
	ctx := context.Background()

	// Empty model/tool in request should be allowed
	result, err := evaluator.Evaluate(ctx, "test-tenant", PolicyRequest{
		Model: "",
		Tool:  "",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Errorf("expected allowed for empty model/tool values")
	}
}

func TestConfigPolicyEvaluator_TenantNotFound(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	evaluator := NewConfigPolicyEvaluator(mgr)
	ctx := context.Background()

	result, err := evaluator.Evaluate(ctx, "non-existent", PolicyRequest{
		Model: "gpt-4",
	})

	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
	if result.Allowed {
		t.Errorf("expected denied for non-existent tenant")
	}
	if result.Reason == "" {
		t.Errorf("expected reason to be set")
	}
}

func TestConfigPolicyEvaluator_NilProvider(t *testing.T) {
	evaluator := NewConfigPolicyEvaluator(nil)
	ctx := context.Background()

	// Should allow all when provider is nil
	result, err := evaluator.Evaluate(ctx, "any-tenant", PolicyRequest{
		Model: "gpt-4",
		Tool:  "search",
	})

	if err != nil {
		t.Errorf("expected no error with nil provider, got %v", err)
	}
	if !result.Allowed {
		t.Errorf("expected allowed with nil provider")
	}
}

func TestConfigPolicyEvaluator_NilEvaluator(t *testing.T) {
	var evaluator *ConfigPolicyEvaluator
	ctx := context.Background()

	// Should allow all when evaluator is nil
	result, err := evaluator.Evaluate(ctx, "any-tenant", PolicyRequest{
		Model: "gpt-4",
	})

	if err != nil {
		t.Errorf("expected no error with nil evaluator, got %v", err)
	}
	if !result.Allowed {
		t.Errorf("expected allowed with nil evaluator")
	}
}

func TestConfigPolicyEvaluator_OnlyModelRestricted(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "model-only",
		Policy: PolicyConfig{
			AllowedModels: []string{"gpt-3.5-turbo"},
			AllowedTools:  []string{}, // All tools allowed
		},
	})

	evaluator := NewConfigPolicyEvaluator(mgr)
	ctx := context.Background()

	// Allowed model with any tool
	result1, err1 := evaluator.Evaluate(ctx, "model-only", PolicyRequest{
		Model: "gpt-3.5-turbo",
		Tool:  "any-tool",
	})
	if err1 != nil || !result1.Allowed {
		t.Errorf("expected allowed for gpt-3.5-turbo with any tool")
	}

	// Disallowed model
	result2, err2 := evaluator.Evaluate(ctx, "model-only", PolicyRequest{
		Model: "gpt-4",
		Tool:  "any-tool",
	})
	if err2 != ErrPolicyDenied {
		t.Errorf("expected ErrPolicyDenied for gpt-4")
	}
	if result2.Allowed {
		t.Errorf("expected denied for gpt-4")
	}
}

func TestConfigPolicyEvaluator_OnlyToolRestricted(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "tool-only",
		Policy: PolicyConfig{
			AllowedModels: []string{}, // All models allowed
			AllowedTools:  []string{"search"},
		},
	})

	evaluator := NewConfigPolicyEvaluator(mgr)
	ctx := context.Background()

	// Any model with allowed tool
	result1, err1 := evaluator.Evaluate(ctx, "tool-only", PolicyRequest{
		Model: "any-model",
		Tool:  "search",
	})
	if err1 != nil || !result1.Allowed {
		t.Errorf("expected allowed for any model with search tool")
	}

	// Any model with disallowed tool
	result2, err2 := evaluator.Evaluate(ctx, "tool-only", PolicyRequest{
		Model: "any-model",
		Tool:  "code_interpreter",
	})
	if err2 != ErrPolicyDenied {
		t.Errorf("expected ErrPolicyDenied for code_interpreter")
	}
	if result2.Allowed {
		t.Errorf("expected denied for code_interpreter")
	}
}

func TestConfigPolicyEvaluator_WithMethodAndPath(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	mgr.SetTenantConfig(Config{
		TenantID: "test-tenant",
		Policy: PolicyConfig{
			AllowedModels: []string{"gpt-4"},
			AllowedTools:  []string{"search"},
		},
	})

	evaluator := NewConfigPolicyEvaluator(mgr)
	ctx := context.Background()

	// Method and Path are currently not evaluated, but should not cause errors
	result, err := evaluator.Evaluate(ctx, "test-tenant", PolicyRequest{
		Model:  "gpt-4",
		Tool:   "search",
		Method: "POST",
		Path:   "/api/chat",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Errorf("expected allowed")
	}
}

func TestAllowedInList(t *testing.T) {
	tests := []struct {
		name     string
		list     []string
		value    string
		expected bool
	}{
		{"empty list allows all", []string{}, "anything", true},
		{"empty value allowed", []string{"a", "b"}, "", true},
		{"value in list", []string{"a", "b", "c"}, "b", true},
		{"value not in list", []string{"a", "b", "c"}, "d", false},
		{"case sensitive", []string{"ABC"}, "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := allowedInList(tt.list, tt.value)
			if result != tt.expected {
				t.Errorf("allowedInList(%v, %s) = %v, want %v",
					tt.list, tt.value, result, tt.expected)
			}
		})
	}
}
