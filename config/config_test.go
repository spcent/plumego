package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestConfigBasic tests basic Config functionality
func TestConfigBasic(t *testing.T) {
	cfg := New()

	// Test empty config
	if cfg.Get("missing") != "" {
		t.Error("Empty config should return empty string for missing keys")
	}

	if cfg.GetString("missing", "default") != "default" {
		t.Error("GetString should return default value for missing keys")
	}

	if cfg.GetInt("missing", 42) != 42 {
		t.Error("GetInt should return default value for missing keys")
	}

	if cfg.GetBool("missing", true) != true {
		t.Error("GetBool should return default value for missing keys")
	}
}

// TestConfigWithData tests Config with actual data
func TestConfigWithData(t *testing.T) {
	cfg := New()
	ctx := context.Background()

	// Create a simple JSON config file
	configData := map[string]any{
		"app_name":        "TestApp",
		"app_version":     "1.0.0",
		"debug_mode":      true,
		"max_connections": 100,
		"timeout_ms":      5000,
		"api_url":         "https://api.example.com",
	}

	tmpFile := "test_config.json"
	content, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Add file source
	cfg.AddSource(NewFileSource(tmpFile, FormatJSON, false))

	// Load config
	err = cfg.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test values
	if cfg.GetString("app_name", "") != "TestApp" {
		t.Errorf("Expected app_name to be 'TestApp', got '%s'", cfg.GetString("app_name", ""))
	}

	if !cfg.GetBool("debug_mode", false) {
		t.Error("Expected debug_mode to be true")
	}

	if cfg.GetInt("max_connections", 0) != 100 {
		t.Errorf("Expected max_connections to be 100, got %d", cfg.GetInt("max_connections", 0))
	}

	if cfg.GetDurationMs("timeout_ms", 0) != 5000*time.Millisecond {
		t.Errorf("Expected timeout_ms to be 5000ms, got %v", cfg.GetDurationMs("timeout_ms", 0))
	}

	if cfg.GetString("api_url", "") != "https://api.example.com" {
		t.Errorf("Expected api_url to be 'https://api.example.com', got '%s'", cfg.GetString("api_url", ""))
	}
}

// TestValidatorRequired tests Required validator
func TestValidatorRequired(t *testing.T) {
	v := &Required{}

	// Test with empty string
	if err := v.Validate("", "test"); err == nil {
		t.Error("Required validator should reject empty string")
	}

	// Test with non-empty string
	if err := v.Validate("value", "test"); err != nil {
		t.Errorf("Required validator should accept non-empty string: %v", err)
	}

	// Test with nil
	if err := v.Validate(nil, "test"); err == nil {
		t.Error("Required validator should reject nil")
	}
}

// TestValidatorMinLength tests MinLength validator
func TestValidatorMinLength(t *testing.T) {
	v := &MinLength{Min: 5}

	// Test with short string
	if err := v.Validate("abc", "test"); err == nil {
		t.Error("MinLength validator should reject short string")
	}

	// Test with exact length
	if err := v.Validate("hello", "test"); err != nil {
		t.Errorf("MinLength validator should accept exact length string: %v", err)
	}

	// Test with long string
	if err := v.Validate("hello world", "test"); err != nil {
		t.Errorf("MinLength validator should accept long string: %v", err)
	}
}

// TestValidatorPattern tests Pattern validator
func TestValidatorPattern(t *testing.T) {
	v := &Pattern{Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`}

	// Test with valid email
	if err := v.Validate("user@example.com", "email"); err != nil {
		t.Errorf("Pattern validator should accept valid email: %v", err)
	}

	// Test with invalid email
	if err := v.Validate("invalid-email", "email"); err == nil {
		t.Error("Pattern validator should reject invalid email")
	}
}

// TestValidatorURL tests URL validator
func TestValidatorURL(t *testing.T) {
	v := &URL{}

	// Test with valid URL
	if err := v.Validate("https://example.com/path", "api_url"); err != nil {
		t.Errorf("URL validator should accept valid URL: %v", err)
	}

	// Test with invalid URL
	if err := v.Validate("not-a-url", "api_url"); err == nil {
		t.Error("URL validator should reject invalid URL")
	}

	// Test with empty string (should be allowed)
	if err := v.Validate("", "api_url"); err != nil {
		t.Errorf("URL validator should allow empty string: %v", err)
	}
}

// TestValidatorRange tests Range validator
func TestValidatorRange(t *testing.T) {
	v := &Range{Min: 0, Max: 100}

	// Test with value in range
	if err := v.Validate(50, "test"); err != nil {
		t.Errorf("Range validator should accept value in range: %v", err)
	}

	// Test with value below range
	if err := v.Validate(-1, "test"); err == nil {
		t.Error("Range validator should reject value below range")
	}

	// Test with value above range
	if err := v.Validate(101, "test"); err == nil {
		t.Error("Range validator should reject value above range")
	}
}

// TestValidatorOneOf tests OneOf validator
func TestValidatorOneOf(t *testing.T) {
	v := &OneOf{Values: []string{"development", "staging", "production"}}

	// Test with valid value
	if err := v.Validate("production", "env"); err != nil {
		t.Errorf("OneOf validator should accept valid value: %v", err)
	}

	// Test with invalid value
	if err := v.Validate("invalid", "env"); err == nil {
		t.Error("OneOf validator should reject invalid value")
	}
}

// TestTypeSafeAccessors tests type-safe configuration accessors
func TestTypeSafeAccessors(t *testing.T) {
	cfg := New()
	ctx := context.Background()

	// Set up test environment variables
	os.Setenv("TEST_STRING", "hello")
	os.Setenv("TEST_INT", "42")
	os.Setenv("TEST_BOOL", "true")
	os.Setenv("TEST_FLOAT", "3.14")
	os.Setenv("TEST_DURATION", "1000")

	// Add environment source with empty prefix to load all env vars
	cfg.AddSource(NewEnvSource(""))

	err := cfg.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify that environment variables are loaded
	allConfig := cfg.GetAll()
	t.Logf("Loaded config data: %v", allConfig)

	// Check if test variables exist
	if _, exists := allConfig["test_string"]; !exists {
		t.Logf("Available keys: %v", func() []string {
			keys := make([]string, 0, len(allConfig))
			for k := range allConfig {
				keys = append(keys, k)
			}
			return keys
		}())
	}

	// Test String accessor
	result, err := cfg.String("test_string", "default", &Required{})
	if err != nil {
		t.Errorf("String accessor should not return error: %v", err)
	}
	if result != "hello" {
		t.Errorf("Expected string 'hello', got '%s'", result)
	}

	// Test Int accessor
	resultInt, err := cfg.Int("test_int", 0, &Range{Min: 0, Max: 1000})
	if err != nil {
		t.Errorf("Int accessor should not return error: %v", err)
	}
	if resultInt != 42 {
		t.Errorf("Expected int 42, got %d", resultInt)
	}

	// Test Bool accessor
	resultBool, err := cfg.Bool("test_bool", false)
	if err != nil {
		t.Errorf("Bool accessor should not return error: %v", err)
	}
	if !resultBool {
		t.Error("Expected bool true, got false")
	}

	// Test Float accessor
	resultFloat, err := cfg.Float("test_float", 0.0)
	if err != nil {
		t.Errorf("Float accessor should not return error: %v", err)
	}
	if resultFloat != 3.14 {
		t.Errorf("Expected float 3.14, got %f", resultFloat)
	}

	// Test Duration accessor
	resultDuration, err := cfg.DurationMs("test_duration", 0)
	if err != nil {
		t.Errorf("Duration accessor should not return error: %v", err)
	}
	if resultDuration != 1000*time.Millisecond {
		t.Errorf("Expected duration 1000ms, got %v", resultDuration)
	}
}

// TestGlobalFunctions tests global helper functions
func TestGlobalFunctions(t *testing.T) {
	// Save original env vars
	origGlobalTest := os.Getenv("GLOBAL_TEST")
	origGlobalInt := os.Getenv("GLOBAL_INT")

	// Clean up after test
	defer func() {
		if origGlobalTest != "" {
			os.Setenv("GLOBAL_TEST", origGlobalTest)
		} else {
			os.Unsetenv("GLOBAL_TEST")
		}
		if origGlobalInt != "" {
			os.Setenv("GLOBAL_INT", origGlobalInt)
		} else {
			os.Unsetenv("GLOBAL_INT")
		}
	}()

	// Clear any existing global config
	SetGlobalConfig(nil)

	// Reset init once for testing
	globalInitOnce = sync.Once{}

	// Set up test environment
	os.Setenv("GLOBAL_TEST", "global_value")
	os.Setenv("GLOBAL_INT", "123")

	// Test global config initialization
	err := InitDefault()
	if err != nil {
		t.Fatalf("InitDefault failed: %v", err)
	}

	// Test global accessors
	result := GetString("global_test", "")
	if result != "global_value" {
		t.Errorf("Global GetString should return correct value, got '%s'", result)
	}

	resultInt := GetInt("global_int", 0)
	if resultInt != 123 {
		t.Errorf("Global GetInt should return correct value, got %d", resultInt)
	}

	// Test global safe accessors
	value, err := GetStringSafe("global_test", "", &Required{})
	if err != nil {
		t.Errorf("Global GetStringSafe should not return error: %v", err)
	}
	if value != "global_value" {
		t.Errorf("Expected global value 'global_value', got '%s'", value)
	}
}

// TestConfigUnmarshal tests struct unmarshalling
func TestConfigUnmarshal(t *testing.T) {
	cfg := New()
	ctx := context.Background()

	// Set up test data
	os.Setenv("APP_NAME", "TestApp")
	os.Setenv("APP_VERSION", "2.0.0")
	os.Setenv("DEBUG_MODE", "true")
	os.Setenv("MAX_CONNECTIONS", "200")

	cfg.AddSource(NewEnvSource(""))

	err := cfg.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Define test struct
	type AppConfig struct {
		AppName        string `config:"app_name"`
		AppVersion     string `config:"app_version"`
		DebugMode      bool   `config:"debug_mode"`
		MaxConnections int    `config:"max_connections"`
	}

	var config AppConfig
	err = cfg.Unmarshal(&config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.AppName != "TestApp" {
		t.Errorf("Expected AppName 'TestApp', got '%s'", config.AppName)
	}

	if config.AppVersion != "2.0.0" {
		t.Errorf("Expected AppVersion '2.0.0', got '%s'", config.AppVersion)
	}

	if !config.DebugMode {
		t.Error("Expected DebugMode to be true")
	}

	if config.MaxConnections != 200 {
		t.Errorf("Expected MaxConnections 200, got %d", config.MaxConnections)
	}
}

// TestFileSource tests FileSource functionality
func TestFileSource(t *testing.T) {
	// Create a test JSON file
	configData := map[string]any{
		"server_port":  8080,
		"database_url": "postgres://localhost:5432/test",
		"enable_ssl":   false,
	}

	tmpFile := filepath.Join(t.TempDir(), "test_config.json")
	content, err := json.Marshal(configData)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	source := NewFileSource(tmpFile, FormatJSON, false)

	// Test Load
	data, err := source.Load(context.Background())
	if err != nil {
		t.Fatalf("FileSource Load failed: %v", err)
	}

	// Check server_port (numbers in JSON are parsed as float64)
	serverPort, ok := data["server_port"].(float64)
	if !ok || int(serverPort) != 8080 {
		t.Errorf("Expected server_port 8080, got %v (type: %T)", data["server_port"], data["server_port"])
	} else {
		t.Logf("server_port correctly loaded: %v", int(serverPort))
	}

	// Check database_url
	databaseURL, ok := data["database_url"].(string)
	if !ok || databaseURL != "postgres://localhost:5432/test" {
		t.Errorf("Expected database_url 'postgres://localhost:5432/test', got %v", data["database_url"])
	}

	// Check enable_ssl
	enableSSL, ok := data["enable_ssl"].(bool)
	if !ok || enableSSL != false {
		t.Errorf("Expected enable_ssl false, got %v", data["enable_ssl"])
	}
}
