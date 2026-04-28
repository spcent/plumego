package config

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
)

type stubSource struct {
	name string
	data map[string]any
	err  error
}

func (s *stubSource) Load(context.Context) (map[string]any, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.data, nil
}

func (s *stubSource) Watch(context.Context) <-chan WatchResult {
	results := make(chan WatchResult)
	close(results)
	return results
}

func (s *stubSource) Name() string { return s.name }

type failingConfigValidator struct{}

func (failingConfigValidator) Validate(any, string) error {
	return errors.New("super-secret-token was invalid")
}

func (failingConfigValidator) Name() string { return "failing" }

type requiredWordValidator struct{}

func (requiredWordValidator) Validate(any, string) error {
	return errors.New("value is required by custom policy")
}

func (requiredWordValidator) Name() string { return "required_word" }

func unsetEnvForTest(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		key := key
		original, existed := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			if existed {
				if err := os.Setenv(key, original); err != nil {
					t.Fatalf("restore %s: %v", key, err)
				}
				return
			}
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("restore unset %s: %v", key, err)
			}
		})
	}
}

func resetGlobalConfigForTest(t *testing.T) {
	t.Helper()
	SetGlobalConfig(nil)
	t.Cleanup(func() {
		SetGlobalConfig(nil)
	})
}

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

func TestGetBoolNumericValues(t *testing.T) {
	cfg := New()
	cfg.data = map[string]any{
		"int_zero":   int8(0),
		"int_one":    int16(1),
		"float_zero": float32(0),
		"float_one":  float64(2.5),
	}

	if cfg.GetBool("int_zero", true) {
		t.Error("expected int_zero to be false")
	}
	if !cfg.GetBool("int_one", false) {
		t.Error("expected int_one to be true")
	}
	if cfg.GetBool("float_zero", true) {
		t.Error("expected float_zero to be false")
	}
	if !cfg.GetBool("float_one", false) {
		t.Error("expected float_one to be true")
	}
}

func TestToIntRejectsOverflow(t *testing.T) {
	const fallback = -7

	if got := toInt(uint64(maxIntValue)+1, fallback); got != fallback {
		t.Fatalf("overflowing uint64 toInt = %d, want fallback %d", got, fallback)
	}
	if got := toInt(math.Inf(1), fallback); got != fallback {
		t.Fatalf("infinite float toInt = %d, want fallback %d", got, fallback)
	}
	if got := toInt(math.NaN(), fallback); got != fallback {
		t.Fatalf("NaN float toInt = %d, want fallback %d", got, fallback)
	}
}

func TestToIntPreservesInRangeConversions(t *testing.T) {
	const fallback = -7

	cases := []struct {
		name  string
		value any
		want  int
	}{
		{name: "int64", value: int64(42), want: 42},
		{name: "uint64", value: uint64(42), want: 42},
		{name: "float64 truncates", value: 42.9, want: 42},
		{name: "bool true", value: true, want: 1},
		{name: "string", value: " 42 ", want: 42},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := toInt(tc.value, fallback); got != tc.want {
				t.Fatalf("toInt(%v) = %d, want %d", tc.value, got, tc.want)
			}
		})
	}
}

// TestConfigWithData tests Config with actual data
func TestConfigWithData(t *testing.T) {
	cfg := New()
	ctx := t.Context()

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

func TestConfigKeyNormalization(t *testing.T) {
	cfg := NewConfigManager(log.NewLogger())
	ctx := t.Context()

	src := &stubSource{
		name: "test",
		data: map[string]any{
			"FooBar": "value",
		},
	}
	if err := cfg.AddSource(src); err != nil {
		t.Fatalf("add source: %v", err)
	}
	if err := cfg.Load(ctx); err != nil {
		t.Fatalf("load: %v", err)
	}

	if got := cfg.Get("foo_bar"); got != "value" {
		t.Fatalf("expected normalized lookup, got %q", got)
	}
	if got := cfg.Get("FOO_BAR"); got != "value" {
		t.Fatalf("expected case-insensitive lookup, got %q", got)
	}
	if got := cfg.Get("FooBar"); got != "value" {
		t.Fatalf("expected original lookup, got %q", got)
	}

	all := cfg.GetAll()
	if _, ok := all["foo_bar"]; !ok {
		t.Fatalf("expected normalized key in GetAll, got %v", all)
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

func TestConfigUnmarshalNestedStruct(t *testing.T) {
	cfg := New()
	cfg.data = map[string]any{
		"app_name":      "demo",
		"db_host":       "localhost",
		"db_port":       "5432",
		"cache_enabled": "true",
	}

	type dbConfig struct {
		Host string `config:"db_host"`
		Port int    `config:"db_port"`
	}

	type cacheConfig struct {
		Enabled bool `config:"cache_enabled"`
	}

	type appConfig struct {
		AppName string `config:"app_name"`
		DB      dbConfig
		Cache   *cacheConfig
	}

	var target appConfig
	if err := cfg.Unmarshal(&target); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if target.AppName != "demo" {
		t.Fatalf("unexpected app name: %s", target.AppName)
	}
	if target.DB.Host != "localhost" || target.DB.Port != 5432 {
		t.Fatalf("unexpected db config: %+v", target.DB)
	}
	if target.Cache == nil || target.Cache.Enabled != true {
		t.Fatalf("unexpected cache config: %#v", target.Cache)
	}
}

func TestLoadBestEffort(t *testing.T) {
	cfg := New()
	cfg.AddSource(&stubSource{name: "bad", err: errors.New("boom")})
	cfg.AddSource(&stubSource{name: "ok", data: map[string]any{"app_name": "demo"}})

	if err := cfg.LoadBestEffort(t.Context()); err != nil {
		t.Fatalf("expected best effort load to succeed, got %v", err)
	}

	if got := cfg.GetString("app_name", ""); got != "demo" {
		t.Fatalf("unexpected value: %q", got)
	}
}

func TestLoadBestEffortAllFail(t *testing.T) {
	cfg := New()
	cfg.AddSource(&stubSource{name: "bad1", err: errors.New("boom1")})
	cfg.AddSource(&stubSource{name: "bad2", err: errors.New("boom2")})

	if err := cfg.LoadBestEffort(t.Context()); err == nil {
		t.Fatalf("expected error when all sources fail")
	}
}

func TestReloadWithValidation(t *testing.T) {
	source := &stubSource{name: "test", data: map[string]any{"value": "ok"}}
	cfg := New()
	cfg.AddSource(source)

	if err := cfg.Load(t.Context()); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	source.data = map[string]any{"value": "bad"}
	err := cfg.ReloadWithValidation(t.Context(), func(data map[string]any) error {
		if data["value"] == "bad" {
			return errors.New("invalid config")
		}
		return nil
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if got := cfg.GetString("value", ""); got != "ok" {
		t.Fatalf("expected rollback to previous value, got %q", got)
	}

	source.data = map[string]any{"value": "good"}
	if err := cfg.ReloadWithValidation(t.Context(), func(data map[string]any) error {
		return nil
	}); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if got := cfg.GetString("value", ""); got != "good" {
		t.Fatalf("expected updated value, got %q", got)
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
	ctx := t.Context()

	// Set up test environment variables
	t.Setenv("TEST_STRING", "hello")
	t.Setenv("TEST_INT", "42")
	t.Setenv("TEST_BOOL", "true")
	t.Setenv("TEST_FLOAT", "3.14")
	t.Setenv("TEST_DURATION", "1000")

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
	// Clear any existing global config (also resets globalInitialized)
	resetGlobalConfigForTest(t)

	// Set up test environment
	t.Setenv("GLOBAL_TEST", "global_value")
	t.Setenv("GLOBAL_INT", "123")

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
	ctx := t.Context()

	// Set up test data
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_VERSION", "2.0.0")
	t.Setenv("DEBUG_MODE", "true")
	t.Setenv("MAX_CONNECTIONS", "200")

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

func TestReloadNotifiesWatchers(t *testing.T) {
	source := &stubSource{name: "test", data: map[string]any{"value": "v1"}}
	cfg := New()
	cfg.AddSource(source)
	cfg.Load(t.Context())

	got := make(chan string, 1)
	cfg.Watch("value", func(_, newVal any) {
		got <- toString(newVal)
	})

	source.data = map[string]any{"value": "v2"}
	if err := cfg.Reload(t.Context()); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	select {
	case v := <-got:
		if v != "v2" {
			t.Fatalf("expected watcher to receive v2, got %q", v)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("watcher not fired after Reload")
	}
}

func TestReloadWithValidationWatcherOnlyOnSuccess(t *testing.T) {
	source := &stubSource{name: "test", data: map[string]any{"value": "ok"}}
	cfg := New()
	cfg.AddSource(source)
	cfg.Load(t.Context())

	fired := make(chan struct{}, 1)
	cfg.Watch("value", func(_, _ any) { fired <- struct{}{} })

	// Failed validation — watcher must NOT fire
	source.data = map[string]any{"value": "bad"}
	_ = cfg.ReloadWithValidation(t.Context(), func(map[string]any) error {
		return errors.New("rejected")
	})
	select {
	case <-fired:
		t.Fatal("watcher fired despite validation failure")
	case <-time.After(50 * time.Millisecond):
	}
	if got := cfg.GetString("value", ""); got != "ok" {
		t.Fatalf("expected rollback to 'ok', got %q", got)
	}

	// Successful validation — watcher MUST fire
	source.data = map[string]any{"value": "good"}
	_ = cfg.ReloadWithValidation(t.Context(), func(map[string]any) error { return nil })
	select {
	case <-fired:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("watcher not fired after successful ReloadWithValidation")
	}
}

func TestUnmarshalDurationAndSlice(t *testing.T) {
	cfg := New()
	cfg.data = map[string]any{
		"timeout":  "5s",
		"tags":     "alpha,beta, gamma",
		"fallback": "2000",
	}

	type T struct {
		Timeout  time.Duration `config:"timeout"`
		Tags     []string      `config:"tags"`
		Fallback time.Duration `config:"fallback"`
	}

	var target T
	if err := cfg.Unmarshal(&target); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if target.Timeout != 5*time.Second {
		t.Errorf("timeout: want 5s, got %v", target.Timeout)
	}
	if len(target.Tags) != 3 || target.Tags[1] != "beta" {
		t.Errorf("tags: want [alpha beta gamma], got %v", target.Tags)
	}
	if target.Fallback != 2000*time.Millisecond {
		t.Errorf("fallback: want 2000ms, got %v", target.Fallback)
	}
}

func TestUnmarshalNumericOverflowReturnsError(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		dst  any
	}{
		{
			name: "int8 overflow",
			data: map[string]any{"value": "128"},
			dst: &struct {
				Value int8 `config:"value"`
			}{},
		},
		{
			name: "uint8 overflow",
			data: map[string]any{"value": "256"},
			dst: &struct {
				Value uint8 `config:"value"`
			}{},
		},
		{
			name: "float32 overflow",
			data: map[string]any{"value": "3.5e39"},
			dst: &struct {
				Value float32 `config:"value"`
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New()
			cfg.data = tt.data

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Unmarshal panicked on overflow: %v", r)
				}
			}()

			err := cfg.Unmarshal(tt.dst)
			if err == nil {
				t.Fatal("expected overflow error")
			}
			if !strings.Contains(err.Error(), "overflows") {
				t.Fatalf("expected overflow error, got %v", err)
			}
		})
	}
}

func TestGetStringSliceAndHas(t *testing.T) {
	cfg := New()
	cfg.data = map[string]any{
		"hosts": "a.com,b.com, c.com",
	}

	if !cfg.Has("hosts") {
		t.Error("Has should return true for existing key")
	}
	if cfg.Has("missing") {
		t.Error("Has should return false for missing key")
	}

	got := cfg.GetStringSlice("hosts", ",", nil)
	if len(got) != 3 || got[2] != "c.com" {
		t.Errorf("GetStringSlice: want [a.com b.com c.com], got %v", got)
	}
	if def := cfg.GetStringSlice("missing", ",", []string{"x"}); len(def) != 1 || def[0] != "x" {
		t.Errorf("GetStringSlice: missing key should return default, got %v", def)
	}
}

func TestGetDuration(t *testing.T) {
	cfg := New()
	cfg.data = map[string]any{
		"go_dur": "10s",
		"ms_dur": "3000",
	}
	if d := cfg.GetDuration("go_dur", 0); d != 10*time.Second {
		t.Errorf("want 10s, got %v", d)
	}
	if d := cfg.GetDuration("ms_dur", 0); d != 3*time.Second {
		t.Errorf("want 3s, got %v", d)
	}
	if d := cfg.GetDuration("missing", 5*time.Second); d != 5*time.Second {
		t.Errorf("want 5s default, got %v", d)
	}
}

func TestFileSourceWithWatchInterval(t *testing.T) {
	src := NewFileSource("config.json", FormatJSON, true).WithWatchInterval(500 * time.Millisecond)
	if src.watchInterval != 500*time.Millisecond {
		t.Errorf("expected 500ms, got %v", src.watchInterval)
	}
}

func TestFileSourceWithWatchIntervalERejectsInvalidInterval(t *testing.T) {
	src := NewFileSource("config.json", FormatJSON, true)
	if got, err := src.WithWatchIntervalE(0); !errors.Is(err, ErrInvalidWatchInterval) || got != nil {
		t.Fatalf("WithWatchIntervalE invalid interval = (%v, %v), want nil ErrInvalidWatchInterval", got, err)
	}
	if src.watchInterval != time.Second {
		t.Fatalf("invalid interval changed watch interval to %v", src.watchInterval)
	}
}

func TestFileSourceWithWatchIntervalPanicsForCompatibility(t *testing.T) {
	src := NewFileSource("config.json", FormatJSON, true)
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected panic for invalid watch interval")
		}
	}()
	_ = src.WithWatchInterval(0)
}

func TestNewManagerERejectsNilLogger(t *testing.T) {
	m, err := NewManagerE(nil)
	if !errors.Is(err, ErrLoggerRequired) {
		t.Fatalf("NewManagerE error = %v, want ErrLoggerRequired", err)
	}
	if m != nil {
		t.Fatalf("NewManagerE manager = %v, want nil", m)
	}
}

func TestNewManagerPanicsForCompatibility(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected panic for nil logger")
		}
	}()
	_ = NewManager(nil)
}

func TestConfigSchemaValidateAllUsesSafeErrors(t *testing.T) {
	schemas := NewConfigSchemaManager()
	schemas.Register("api_token", ConfigSchemaEntry{
		Required:   true,
		Validators: []Validator{failingConfigValidator{}},
	})

	errs := schemas.ValidateAll(map[string]any{"api_token": "super-secret-token"})
	if len(errs) != 1 {
		t.Fatalf("expected one validation error, got %d", len(errs))
	}
	if errs[0].Code != codeConfigValidationFailed {
		t.Fatalf("code=%s, want %s", errs[0].Code, codeConfigValidationFailed)
	}
	if errs[0].Message != "configuration value failed validation" {
		t.Fatalf("message=%q, want safe validation message", errs[0].Message)
	}
	if errs[0].Details["value"] != nil {
		t.Fatalf("validation details expose raw value: %v", errs[0].Details)
	}
	if errs[0].Details["key"] != "api_token" {
		t.Fatalf("expected key detail, got %v", errs[0].Details)
	}
	if errs[0].Details["validator"] != "failing" {
		t.Fatalf("expected validator detail, got %v", errs[0].Details)
	}
	if errs[0].Type != contract.TypeValidation {
		t.Fatalf("type=%s, want %s", errs[0].Type, contract.TypeValidation)
	}

	errs = schemas.ValidateAll(map[string]any{})
	if len(errs) != 1 {
		t.Fatalf("expected one required error, got %d", len(errs))
	}
	if errs[0].Code != codeConfigRequired {
		t.Fatalf("code=%s, want %s", errs[0].Code, codeConfigRequired)
	}
	if errs[0].Message != "required configuration is missing" {
		t.Fatalf("message=%q, want safe required message", errs[0].Message)
	}
}

func TestConfigSchemaValidateMissingFieldUsesRequiredValidatorType(t *testing.T) {
	schema := NewConfigSchema()
	schema.AddField("optional", requiredWordValidator{})
	schema.AddField("required", &Required{})

	err := schema.Validate(map[string]any{})
	if err == nil {
		t.Fatal("expected missing required field error")
	}
	msg := err.Error()
	if strings.Contains(msg, "optional") {
		t.Fatalf("optional custom validator should not be treated as required: %v", err)
	}
	if !strings.Contains(msg, "required") {
		t.Fatalf("expected required field error, got %v", err)
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
	data, err := source.Load(t.Context())
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
