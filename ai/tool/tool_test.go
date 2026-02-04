package tool

import (
	"context"
	"fmt"
	"testing"

	"github.com/spcent/plumego/ai/provider"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	tool := NewEchoTool()
	if err := registry.Register(tool); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Try to register same tool again
	if err := registry.Register(tool); err == nil {
		t.Error("Register() should error when registering duplicate tool")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	echoTool := NewEchoTool()

	registry.Register(echoTool)

	tool, err := registry.Get("echo")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if tool.Name() != "echo" {
		t.Errorf("Tool name = %v, want echo", tool.Name())
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Error("Get() should error for nonexistent tool")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	registry.Register(NewEchoTool())
	registry.Register(NewCalculatorTool())

	tools := registry.List()
	if len(tools) != 2 {
		t.Errorf("List() count = %v, want 2", len(tools))
	}
}

func TestRegistry_Execute(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewEchoTool())

	input := map[string]any{
		"message": "Hello, World!",
	}

	result, err := registry.Execute(context.Background(), "echo", input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.IsError {
		t.Error("Result should not be an error")
	}

	output, ok := result.Output.(map[string]any)
	if !ok {
		t.Fatal("Output should be a map")
	}

	if echoed := output["echoed"]; echoed != "Hello, World!" {
		t.Errorf("Echoed message = %v, want 'Hello, World!'", echoed)
	}
}

func TestRegistry_ExecuteToolUse(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewCalculatorTool())

	toolUse := &provider.ToolUse{
		ID:   "tool-1",
		Name: "calculator",
		Input: map[string]any{
			"operation": "add",
			"a":         float64(5),
			"b":         float64(3),
		},
	}

	result, err := registry.ExecuteToolUse(context.Background(), toolUse)
	if err != nil {
		t.Fatalf("ExecuteToolUse() error = %v", err)
	}

	output, ok := result.Output.(map[string]any)
	if !ok {
		t.Fatal("Output should be a map")
	}

	if result := output["result"]; result != float64(8) {
		t.Errorf("Result = %v, want 8", result)
	}
}

func TestRegistry_ToProviderTools(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewEchoTool())
	registry.Register(NewCalculatorTool())

	tools := registry.ToProviderTools(context.Background())

	if len(tools) != 2 {
		t.Errorf("Tool count = %v, want 2", len(tools))
	}

	for _, tool := range tools {
		if tool.Type != "function" {
			t.Errorf("Tool type = %v, want function", tool.Type)
		}

		if tool.Function.Name == "" {
			t.Error("Tool name should not be empty")
		}

		if tool.Function.Description == "" {
			t.Error("Tool description should not be empty")
		}
	}
}

func TestAllowListPolicy(t *testing.T) {
	policy := NewAllowListPolicy([]string{"echo", "calculator"})

	if !policy.CanExecute(context.Background(), "echo") {
		t.Error("CanExecute() should allow 'echo'")
	}

	if policy.CanExecute(context.Background(), "bash") {
		t.Error("CanExecute() should not allow 'bash'")
	}

	// Add tool
	policy.Add("bash")
	if !policy.CanExecute(context.Background(), "bash") {
		t.Error("CanExecute() should allow 'bash' after Add()")
	}

	// Remove tool
	policy.Remove("bash")
	if policy.CanExecute(context.Background(), "bash") {
		t.Error("CanExecute() should not allow 'bash' after Remove()")
	}
}

func TestRegistry_WithPolicy(t *testing.T) {
	policy := NewAllowListPolicy([]string{"echo"})
	registry := NewRegistry(WithPolicy(policy))

	registry.Register(NewEchoTool())
	registry.Register(NewCalculatorTool())

	// Echo should work
	_, err := registry.Execute(context.Background(), "echo", map[string]any{
		"message": "test",
	})
	if err != nil {
		t.Errorf("Execute(echo) error = %v", err)
	}

	// Calculator should be blocked
	_, err = registry.Execute(context.Background(), "calculator", map[string]any{
		"operation": "add",
		"a":         1.0,
		"b":         2.0,
	})
	if err == nil {
		t.Error("Execute(calculator) should be blocked by policy")
	}
}

func TestEchoTool(t *testing.T) {
	tool := NewEchoTool()

	if tool.Name() != "echo" {
		t.Errorf("Name() = %v, want echo", tool.Name())
	}

	result, err := tool.Execute(context.Background(), map[string]any{
		"message": "test",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.IsError {
		t.Error("Result should not be an error")
	}
}

func TestCalculatorTool(t *testing.T) {
	tool := NewCalculatorTool()

	tests := []struct {
		name      string
		input     map[string]any
		expected  float64
		wantError bool
	}{
		{
			name: "add",
			input: map[string]any{
				"operation": "add",
				"a":         float64(5),
				"b":         float64(3),
			},
			expected: 8,
		},
		{
			name: "subtract",
			input: map[string]any{
				"operation": "subtract",
				"a":         float64(10),
				"b":         float64(4),
			},
			expected: 6,
		},
		{
			name: "multiply",
			input: map[string]any{
				"operation": "multiply",
				"a":         float64(4),
				"b":         float64(5),
			},
			expected: 20,
		},
		{
			name: "divide",
			input: map[string]any{
				"operation": "divide",
				"a":         float64(10),
				"b":         float64(2),
			},
			expected: 5,
		},
		{
			name: "divide by zero",
			input: map[string]any{
				"operation": "divide",
				"a":         float64(10),
				"b":         float64(0),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(context.Background(), tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("Execute() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			output, ok := result.Output.(map[string]any)
			if !ok {
				t.Fatal("Output should be a map")
			}

			if result := output["result"]; result != tt.expected {
				t.Errorf("Result = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimestampTool(t *testing.T) {
	tool := NewTimestampTool()

	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output, ok := result.Output.(map[string]any)
	if !ok {
		t.Fatal("Output should be a map")
	}

	if _, ok := output["timestamp"]; !ok {
		t.Error("Output should contain timestamp")
	}

	if _, ok := output["iso8601"]; !ok {
		t.Error("Output should contain iso8601")
	}

	if _, ok := output["human"]; !ok {
		t.Error("Output should contain human")
	}
}

func TestFormatResultAsString(t *testing.T) {
	tests := []struct {
		name     string
		result   *Result
		contains string
	}{
		{
			name: "success",
			result: &Result{
				Output: map[string]any{"result": 42},
			},
			contains: "result",
		},
		{
			name: "error",
			result: &Result{
				Error:   fmt.Errorf("test error"),
				IsError: true,
			},
			contains: "Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := FormatResultAsString(tt.result)
			if !contains(str, tt.contains) {
				t.Errorf("FormatResultAsString() = %v, should contain %v", str, tt.contains)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		 findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
