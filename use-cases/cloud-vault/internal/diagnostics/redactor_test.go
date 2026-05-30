package diagnostics

import (
	"testing"

	"cloud-vault/internal/config"
)

func TestRedactConfig(t *testing.T) {
	cfg := config.Config{
		App: config.AppConfig{
			Version: "1.0.0",
		},
		Qiniu: config.QiniuConfig{
			AccessKey: "secret-access-key-123",
			SecretKey: "super-secret-key-456",
			Bucket:    "my-bucket",
		},
		AI: config.AIConfig{
			APIKey: "sk-ant-api-key-789",
			Model:  "claude-3-5-sonnet",
		},
	}

	redacted := RedactConfig(cfg)

	// Version should be visible
	if redacted.App["version"] != "1.0.0" {
		t.Errorf("version should not be redacted, got: %v", redacted.App["version"])
	}

	// Bucket should be visible
	if redacted.Storage["qiniu_bucket"] != "my-bucket" {
		t.Errorf("bucket should not be redacted, got: %v", redacted.Storage["qiniu_bucket"])
	}

	// Model should be visible
	if redacted.AI["model"] != "claude-3-5-sonnet" {
		t.Errorf("model should not be redacted, got: %v", redacted.AI["model"])
	}

	// Secrets should be redacted
	if redacted.Storage["qiniu_access_key"] != "<redacted>" {
		t.Errorf("access key should be redacted, got: %v", redacted.Storage["qiniu_access_key"])
	}

	if redacted.Storage["qiniu_secret_key"] != "<redacted>" {
		t.Errorf("secret key should be redacted, got: %v", redacted.Storage["qiniu_secret_key"])
	}

	if redacted.AI["api_key"] != "<redacted>" {
		t.Errorf("API key should be redacted, got: %v", redacted.AI["api_key"])
	}
}

func TestRedactLogLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Bearer token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "Authorization: Bearer <redacted>",
		},
		{
			name:     "API key in URL",
			input:    "GET /api?api_key=sk-ant-123456 HTTP/1.1",
			expected: "GET /api?api_key=<redacted> HTTP/1.1",
		},
		{
			name:     "Password in JSON",
			input:    `{"username": "admin", "password": "secret123"}`,
			expected: `{"username": "admin", "password": "<redacted>"}`,
		},
		{
			name:     "Token parameter",
			input:    "Processing token=abc123xyz for user",
			expected: "Processing token=<redacted> for user",
		},
		{
			name:     "No sensitive data",
			input:    "User logged in successfully",
			expected: "User logged in successfully",
		},
		{
			name:     "Multiple sensitive fields",
			input:    "api_key=key123 and secret_key=secret456",
			expected: "api_key=<redacted> and secret_key=<redacted>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactLogLine(tt.input)
			if result != tt.expected {
				t.Errorf("RedactLogLine(%q)\ngot:      %q\nexpected: %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRedactLogLinePreservesStructure(t *testing.T) {
	input := "2024-01-15 10:30:45 INFO Request completed in 150ms"
	result := RedactLogLine(input)

	if result != input {
		t.Errorf("non-sensitive log should be unchanged\ngot:      %q\nexpected: %q", result, input)
	}
}
