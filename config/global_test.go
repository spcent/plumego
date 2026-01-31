package config

import (
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/spcent/plumego/log"
)

func TestGetBool(t *testing.T) {
	// Clean up environment first
	os.Unsetenv("test_bool_true")
	os.Unsetenv("test_bool_false")
	os.Unsetenv("test_bool_invalid")
	os.Unsetenv("test_bool_missing")

	// Set up a clean global config for testing
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	testConfig.AddSource(NewEnvSource(""))

	// Manually set the environment variables before loading
	os.Setenv("test_bool_true", "true")
	os.Setenv("test_bool_false", "false")
	os.Setenv("test_bool_invalid", "invalid")

	// Load the config
	_ = testConfig.Load(t.Context())
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	// Test with environment variable set to true
	result := GetBool("test_bool_true", false)
	if !result {
		t.Errorf("expected true, got false")
	}

	// Test with environment variable set to false
	result = GetBool("test_bool_false", true)
	if result {
		t.Errorf("expected false, got true")
	}

	// Test with default value when env var not set
	result = GetBool("test_bool_missing", true)
	if !result {
		t.Errorf("expected true (default), got false")
	}

	// Test with invalid value
	result = GetBool("test_bool_invalid", true)
	if !result {
		t.Errorf("expected true (default on invalid), got false")
	}
}

func TestGetFloat(t *testing.T) {
	// Clean up first
	os.Unsetenv("test_float")
	os.Unsetenv("test_float_missing")
	os.Unsetenv("test_float_invalid")

	// Set up a clean global config for testing
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	testConfig.AddSource(NewEnvSource(""))

	// Manually set environment variables
	os.Setenv("test_float", "3.14")
	os.Setenv("test_float_invalid", "not_a_float")

	// Load the config
	_ = testConfig.Load(t.Context())
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	// Test with valid float
	result := GetFloat("test_float", 1.0)
	if result != 3.14 {
		t.Errorf("expected 3.14, got %f", result)
	}

	// Test with default value
	result = GetFloat("test_float_missing", 2.5)
	if result != 2.5 {
		t.Errorf("expected 2.5 (default), got %f", result)
	}

	// Test with invalid value
	result = GetFloat("test_float_invalid", 1.5)
	if result != 1.5 {
		t.Errorf("expected 1.5 (default on invalid), got %f", result)
	}
}

func TestGetDurationMs(t *testing.T) {
	// Clean up first
	os.Unsetenv("test_duration")
	os.Unsetenv("test_duration_missing")
	os.Unsetenv("test_duration_invalid")

	// Set up a clean global config for testing
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	testConfig.AddSource(NewEnvSource(""))

	// Manually set environment variables
	os.Setenv("test_duration", "5000")
	os.Setenv("test_duration_invalid", "invalid")

	// Load the config
	_ = testConfig.Load(t.Context())
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	// Test with valid milliseconds
	result := GetDurationMs("test_duration", 1000)
	expected := 5 * time.Second
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Test with default value
	result = GetDurationMs("test_duration_missing", 2000)
	expected = 2 * time.Second
	if result != expected {
		t.Errorf("expected %v (default), got %v", expected, result)
	}

	// Test with invalid value
	result = GetDurationMs("test_duration_invalid", 3000)
	expected = 3 * time.Second
	if result != expected {
		t.Errorf("expected %v (default on invalid), got %v", expected, result)
	}
}

func TestSet(t *testing.T) {
	// Set up a clean global config for testing
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	testConfig.AddSource(NewEnvSource(""))
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	// Test setting a value
	err := Set("TEST_SET_KEY", "test_value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the value is set
	if os.Getenv("TEST_SET_KEY") != "test_value" {
		t.Errorf("expected test_value, got %s", os.Getenv("TEST_SET_KEY"))
	}

	// Clean up
	os.Unsetenv("TEST_SET_KEY")
}

func TestGetGlobalConfig(t *testing.T) {
	// Test 1: With existing config
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	config := GetGlobalConfig()
	if config == nil {
		t.Error("expected non-nil config")
	}
	if config != testConfig {
		t.Error("expected same config instance")
	}

	// Test 2: Verify GetGlobalConfig returns the same instance on subsequent calls
	result := GetGlobalConfig()
	if result != config {
		t.Error("expected same config instance on second call")
	}
}

func TestGetStringSafe(t *testing.T) {
	// Clean up first
	os.Unsetenv("test_safe_string")
	os.Unsetenv("test_safe_string_missing")

	// Set up a clean global config for testing
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	testConfig.AddSource(NewEnvSource(""))

	// Manually set environment variables
	os.Setenv("test_safe_string", "valid")

	// Load the config
	_ = testConfig.Load(t.Context())
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	// Test with valid value
	result, err := GetStringSafe("test_safe_string", "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "valid" {
		t.Errorf("expected 'valid', got '%s'", result)
	}

	// Test with default
	result, err = GetStringSafe("test_safe_string_missing", "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "default" {
		t.Errorf("expected 'default', got '%s'", result)
	}
}

func TestGetIntSafe(t *testing.T) {
	// Clean up first
	os.Unsetenv("test_safe_int")
	os.Unsetenv("test_safe_int_missing")

	// Set up a clean global config for testing
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	testConfig.AddSource(NewEnvSource(""))

	// Manually set environment variables
	os.Setenv("test_safe_int", "42")

	// Load the config
	_ = testConfig.Load(t.Context())
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	// Test with valid value
	result, err := GetIntSafe("test_safe_int", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}

	// Test with default
	result, err = GetIntSafe("test_safe_int_missing", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 20 {
		t.Errorf("expected 20, got %d", result)
	}
}

func TestGetBoolSafe(t *testing.T) {
	// Clean up first
	os.Unsetenv("test_safe_bool")
	os.Unsetenv("test_safe_bool_missing")

	// Set up a clean global config for testing
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	testConfig.AddSource(NewEnvSource(""))

	// Manually set environment variables
	os.Setenv("test_safe_bool", "true")

	// Load the config
	_ = testConfig.Load(t.Context())
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	// Test with valid value
	result, err := GetBoolSafe("test_safe_bool", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true")
	}

	// Test with default
	result, err = GetBoolSafe("test_safe_bool_missing", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true (default)")
	}
}

func TestGetFloatSafe(t *testing.T) {
	// Clean up first
	os.Unsetenv("test_safe_float")
	os.Unsetenv("test_safe_float_missing")

	// Set up a clean global config for testing
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	testConfig.AddSource(NewEnvSource(""))

	// Manually set environment variables
	os.Setenv("test_safe_float", "3.14")

	// Load the config
	_ = testConfig.Load(t.Context())
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	// Test with valid value
	result, err := GetFloatSafe("test_safe_float", 1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 3.14 {
		t.Errorf("expected 3.14, got %f", result)
	}

	// Test with default
	result, err = GetFloatSafe("test_safe_float_missing", 2.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 2.5 {
		t.Errorf("expected 2.5, got %f", result)
	}
}

func TestGetDurationMsSafe(t *testing.T) {
	// Clean up first
	os.Unsetenv("test_safe_duration")
	os.Unsetenv("test_safe_duration_missing")

	// Set up a clean global config for testing
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)
	testConfig.AddSource(NewEnvSource(""))

	// Manually set environment variables
	os.Setenv("test_safe_duration", "1000")

	// Load the config
	_ = testConfig.Load(t.Context())
	SetGlobalConfig(testConfig)
	defer SetGlobalConfig(nil)

	// Test with valid value
	result, err := GetDurationMsSafe("test_safe_duration", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 1 * time.Second
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Test with default
	result, err = GetDurationMsSafe("test_safe_duration_missing", 2000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected = 2 * time.Second
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestLoadEnvFile(t *testing.T) {
	// Create a temporary .env file
	content := `TEST_ENV_VAR1=value1
TEST_ENV_VAR2=value2
# This is a comment
TEST_ENV_VAR3="quoted value"
`
	tmpFile, err := os.CreateTemp("", "test_env_*.env")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test loading with overwrite
	err = LoadEnvFile(tmpFile.Name(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify values are set
	if os.Getenv("TEST_ENV_VAR1") != "value1" {
		t.Errorf("expected 'value1', got '%s'", os.Getenv("TEST_ENV_VAR1"))
	}
	if os.Getenv("TEST_ENV_VAR2") != "value2" {
		t.Errorf("expected 'value2', got '%s'", os.Getenv("TEST_ENV_VAR2"))
	}
	if os.Getenv("TEST_ENV_VAR3") != "quoted value" {
		t.Errorf("expected 'quoted value', got '%s'", os.Getenv("TEST_ENV_VAR3"))
	}

	// Clean up
	os.Unsetenv("TEST_ENV_VAR1")
	os.Unsetenv("TEST_ENV_VAR2")
	os.Unsetenv("TEST_ENV_VAR3")
}

func TestLoadEnvFileErrors(t *testing.T) {
	// Test with non-existent file
	err := LoadEnvFile("/non/existent/file.env", true)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Test with invalid file content
	tmpFile, err := os.CreateTemp("", "test_invalid_*.env")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid content (no equals sign)
	if _, err := tmpFile.WriteString("invalid_line\n"); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Should not error on invalid lines, just skip them
	err = LoadEnvFile(tmpFile.Name(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitDefault(t *testing.T) {
	// Reset global state
	globalInitOnce = sync.Once{}
	globalConfig = nil

	// Test InitDefault with no config files
	err := InitDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify global config is set
	globalConfigMu.RLock()
	if globalConfig == nil {
		t.Error("global config should be initialized")
	}
	globalConfigMu.RUnlock()
}

func TestSetGlobalConfig(t *testing.T) {
	// Create a test config
	logger := log.NewGLogger()
	testConfig := NewConfigManager(logger)

	// Set global config
	SetGlobalConfig(testConfig)

	// Verify it's set
	globalConfigMu.RLock()
	if globalConfig != testConfig {
		t.Error("global config not set correctly")
	}
	globalConfigMu.RUnlock()

	// Reset for other tests
	SetGlobalConfig(nil)
}
