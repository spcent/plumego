package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// NewEchoTool creates a simple echo tool for testing.
func NewEchoTool() Tool {
	return NewFuncTool(
		"echo",
		"Echoes back the input message",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{
					"type":        "string",
					"description": "The message to echo",
				},
			},
			"required": []string{"message"},
		},
		func(ctx context.Context, input map[string]any) (any, error) {
			message, ok := input["message"].(string)
			if !ok {
				return nil, fmt.Errorf("message must be a string")
			}

			return map[string]any{
				"echoed": message,
			}, nil
		},
	)
}

// NewCalculatorTool creates a basic calculator tool.
func NewCalculatorTool() Tool {
	return NewFuncTool(
		"calculator",
		"Performs basic arithmetic operations",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"operation": map[string]any{
					"type":        "string",
					"description": "The operation to perform (add, subtract, multiply, divide)",
					"enum":        []string{"add", "subtract", "multiply", "divide"},
				},
				"a": map[string]any{
					"type":        "number",
					"description": "First number",
				},
				"b": map[string]any{
					"type":        "number",
					"description": "Second number",
				},
			},
			"required": []string{"operation", "a", "b"},
		},
		func(ctx context.Context, input map[string]any) (any, error) {
			operation, ok := input["operation"].(string)
			if !ok {
				return nil, fmt.Errorf("operation must be a string")
			}

			// Convert numbers (JSON may give us float64)
			var a, b float64
			switch v := input["a"].(type) {
			case float64:
				a = v
			case int:
				a = float64(v)
			default:
				return nil, fmt.Errorf("a must be a number")
			}

			switch v := input["b"].(type) {
			case float64:
				b = v
			case int:
				b = float64(v)
			default:
				return nil, fmt.Errorf("b must be a number")
			}

			var result float64
			switch operation {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			case "multiply":
				result = a * b
			case "divide":
				if b == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = a / b
			default:
				return nil, fmt.Errorf("unknown operation: %s", operation)
			}

			return map[string]any{
				"result": result,
			}, nil
		},
	)
}

// NewReadFileTool creates a tool for reading files.
// WARNING: This should be used with caution in production due to security implications.
func NewReadFileTool() Tool {
	return NewFuncTool(
		"read_file",
		"Reads the contents of a file",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to the file to read",
				},
			},
			"required": []string{"path"},
		},
		func(ctx context.Context, input map[string]any) (any, error) {
			path, ok := input["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path must be a string")
			}

			// Basic security check
			if path == "" {
				return nil, fmt.Errorf("path cannot be empty")
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read file: %w", err)
			}

			return map[string]any{
				"content": string(content),
				"size":    len(content),
			}, nil
		},
	)
}

// NewWriteFileTool creates a tool for writing files.
// WARNING: This should be used with caution in production due to security implications.
func NewWriteFileTool() Tool {
	return NewFuncTool(
		"write_file",
		"Writes content to a file",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to the file to write",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Content to write to the file",
				},
			},
			"required": []string{"path", "content"},
		},
		func(ctx context.Context, input map[string]any) (any, error) {
			path, ok := input["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path must be a string")
			}

			content, ok := input["content"].(string)
			if !ok {
				return nil, fmt.Errorf("content must be a string")
			}

			if path == "" {
				return nil, fmt.Errorf("path cannot be empty")
			}

			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return nil, fmt.Errorf("write file: %w", err)
			}

			return map[string]any{
				"success": true,
				"size":    len(content),
			}, nil
		},
	)
}

// NewBashTool creates a tool for executing bash commands.
// WARNING: This is extremely dangerous and should only be used in trusted environments.
func NewBashTool(timeout time.Duration) Tool {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return NewFuncTool(
		"bash",
		"Executes a bash command",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "The bash command to execute",
				},
			},
			"required": []string{"command"},
		},
		func(ctx context.Context, input map[string]any) (any, error) {
			command, ok := input["command"].(string)
			if !ok {
				return nil, fmt.Errorf("command must be a string")
			}

			if command == "" {
				return nil, fmt.Errorf("command cannot be empty")
			}

			// Create command with timeout
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, "bash", "-c", command)
			output, err := cmd.CombinedOutput()

			if err != nil {
				return map[string]any{
					"output":    string(output),
					"exit_code": cmd.ProcessState.ExitCode(),
					"error":     err.Error(),
				}, nil // Don't return error, include it in output
			}

			return map[string]any{
				"output":    string(output),
				"exit_code": 0,
			}, nil
		},
	)
}

// NewTimestampTool creates a tool that returns the current timestamp.
func NewTimestampTool() Tool {
	return NewFuncTool(
		"get_timestamp",
		"Returns the current timestamp",
		map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		func(ctx context.Context, input map[string]any) (any, error) {
			now := time.Now()
			return map[string]any{
				"timestamp": now.Unix(),
				"iso8601":   now.Format(time.RFC3339),
				"human":     now.Format("2006-01-02 15:04:05"),
			}, nil
		},
	)
}
