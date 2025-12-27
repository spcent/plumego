package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadEnv(t *testing.T) {
	// Create a temporary file to simulate .env
	content := `
# Comment line
DB_HOST=127.0.0.1
DB_USER=root
DB_PASS="secret"
EMPTY_KEY=
QUOTED_KEY='quoted_value'
`
	tmpFile := "test.env"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Clear environment variables to avoid polluting tests
	os.Clearenv()

	// Set an existing variable to ensure it is not overwritten
	os.Setenv("DB_USER", "existing_user")

	// Call LoadEnv
	err = LoadEnv(tmpFile, false)
	if err != nil {
		t.Fatalf("LoadEnv execution failed: %v", err)
	}

	// Validate results
	tests := []struct {
		key      string
		expected string
	}{
		{"DB_HOST", "127.0.0.1"},
		{"DB_USER", "existing_user"}, // do not overwrite existing values
		{"DB_PASS", "secret"},
		{"EMPTY_KEY", ""},
		{"QUOTED_KEY", "quoted_value"},
	}

	for _, tt := range tests {
		got := os.Getenv(tt.key)
		if got != tt.expected {
			t.Errorf("Environment variable %s = %q, expected %q", tt.key, got, tt.expected)
		}
	}
}

func TestLoadEnvWithOverwrite(t *testing.T) {
	content := `
DB_HOST=127.0.0.1
DB_USER=root
`
	tmpFile := "test_overwrite.env"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Scenario 1: do not overwrite existing values
	os.Clearenv()
	os.Setenv("DB_USER", "existing_user")
	err = LoadEnv(tmpFile, false)
	if err != nil {
		t.Fatalf("LoadEnv execution failed: %v", err)
	}
	if got := os.Getenv("DB_USER"); got != "existing_user" {
		t.Errorf("DB_USER should remain existing_user, got %q", got)
	}

	// Scenario 2: overwrite existing values
	os.Clearenv()
	os.Setenv("DB_USER", "existing_user")
	err = LoadEnv(tmpFile, true)
	if err != nil {
		t.Fatalf("LoadEnv execution failed: %v", err)
	}
	if got := os.Getenv("DB_USER"); got != "root" {
		t.Errorf("DB_USER should be overwritten to root, got %q", got)
	}
}

func TestLoadEnvFileNotFound(t *testing.T) {
	os.Clearenv()
	err := LoadEnv("nonexistent.env", false)
	if err == nil {
		t.Fatal("Expected LoadEnv to return an error, but it did not")
	}
}

func TestGetHelpers(t *testing.T) {
	if got := GetString("MISSING", "default"); got != "default" {
		t.Fatalf("GetString default fallback failed: %q", got)
	}

	t.Setenv("TEST_STRING", "  spaced ")
	if got := GetString("TEST_STRING", "default"); got != "spaced" {
		t.Fatalf("GetString should trim whitespace, got %q", got)
	}

	t.Setenv("TEST_INT", "notanint")
	if got := GetInt("TEST_INT", 5); got != 5 {
		t.Fatalf("GetInt should fallback to default on invalid input, got %d", got)
	}
	t.Setenv("TEST_INT", " 42 ")
	if got := GetInt("TEST_INT", 5); got != 42 {
		t.Fatalf("GetInt should parse trimmed integer, got %d", got)
	}

	t.Setenv("TEST_BOOL_TRUE", "yes")
	if !GetBool("TEST_BOOL_TRUE", false) {
		t.Fatalf("GetBool should parse affirmative values")
	}
	t.Setenv("TEST_BOOL_FALSE", "OFF")
	if GetBool("TEST_BOOL_FALSE", true) {
		t.Fatalf("GetBool should parse negative values")
	}
	t.Setenv("TEST_BOOL_INVALID", "maybe")
	if !GetBool("TEST_BOOL_INVALID", true) {
		t.Fatalf("GetBool should fallback to default on invalid input")
	}

	t.Setenv("TEST_FLOAT", " 1.5 ")
	if got := GetFloat("TEST_FLOAT", 0); got != 1.5 {
		t.Fatalf("GetFloat should parse trimmed float, got %f", got)
	}
	t.Setenv("TEST_DURATION_MS", " 10 ")
	if got := GetDurationMs("TEST_DURATION_MS", 0); got != 10*time.Millisecond {
		t.Fatalf("GetDurationMs should parse milliseconds, got %s", got)
	}
}
