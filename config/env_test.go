package config

import (
	"context"
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
	// 创建自定义配置实例
	cfg := New()

	// 添加环境变量源
	envSource := NewEnvSource("")
	cfg.AddSource(envSource)

	// 加载配置
	ctx := context.Background()
	err := cfg.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 测试缺省值
	if got := cfg.GetString("MISSING", "default"); got != "default" {
		t.Fatalf("GetString default fallback failed: %q", got)
	}

	// 设置环境变量
	t.Setenv("TEST_STRING", "  spaced ")

	// 重新加载配置以获取新的环境变量
	err = cfg.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	// 测试字符串
	if got := cfg.GetString("TEST_STRING", "default"); got != "spaced" {
		t.Fatalf("GetString should trim whitespace, got %q", got)
	}

	// 测试整数
	t.Setenv("TEST_INT", "notanint")
	cfg.Load(ctx)
	if got := cfg.GetInt("TEST_INT", 5); got != 5 {
		t.Fatalf("GetInt should fallback to default on invalid input, got %d", got)
	}
	t.Setenv("TEST_INT", " 42 ")
	cfg.Load(ctx)
	if got := cfg.GetInt("TEST_INT", 5); got != 42 {
		t.Fatalf("GetInt should parse trimmed integer, got %d", got)
	}

	// 测试布尔值
	t.Setenv("TEST_BOOL_TRUE", "yes")
	cfg.Load(ctx)
	if !cfg.GetBool("TEST_BOOL_TRUE", false) {
		t.Fatalf("GetBool should parse affirmative values")
	}
	t.Setenv("TEST_BOOL_FALSE", "OFF")
	cfg.Load(ctx)
	if cfg.GetBool("TEST_BOOL_FALSE", true) {
		t.Fatalf("GetBool should parse negative values")
	}
	t.Setenv("TEST_BOOL_INVALID", "maybe")
	cfg.Load(ctx)
	if !cfg.GetBool("TEST_BOOL_INVALID", true) {
		t.Fatalf("GetBool should fallback to default on invalid input")
	}

	// 测试浮点数
	t.Setenv("TEST_FLOAT", " 1.5 ")
	cfg.Load(ctx)
	if got := cfg.GetFloat("TEST_FLOAT", 0); got != 1.5 {
		t.Fatalf("GetFloat should parse trimmed float, got %f", got)
	}

	// 测试持续时间
	t.Setenv("TEST_DURATION_MS", " 10 ")
	cfg.Load(ctx)
	if got := cfg.GetDurationMs("TEST_DURATION_MS", 0); got != 10*time.Millisecond {
		t.Fatalf("GetDurationMs should parse milliseconds, got %s", got)
	}
}
