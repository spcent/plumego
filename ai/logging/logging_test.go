package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("Level.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestFields(t *testing.T) {
	fields := Fields("key1", "value1", "key2", 42)

	if len(fields) != 2 {
		t.Fatalf("Fields() returned %d fields, want 2", len(fields))
	}

	if fields[0].Key != "key1" || fields[0].Value != "value1" {
		t.Errorf("Field 0 = {%v, %v}, want {key1, value1}", fields[0].Key, fields[0].Value)
	}

	if fields[1].Key != "key2" || fields[1].Value != 42 {
		t.Errorf("Field 1 = {%v, %v}, want {key2, 42}", fields[1].Key, fields[1].Value)
	}
}

func TestFields_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Fields() should panic with odd number of arguments")
		}
	}()

	Fields("key1", "value1", "key2")
}

func TestFields_NonStringKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Fields() should panic with non-string key")
		}
	}()

	Fields(123, "value")
}

func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}

	// Should not panic
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")

	child := logger.With(Fields("key", "value")...)
	if child != logger {
		t.Error("NoOpLogger.With() should return same instance")
	}

	withCtx := logger.WithContext(context.Background())
	if withCtx != logger {
		t.Error("NoOpLogger.WithContext() should return same instance")
	}

	logger.SetLevel(ErrorLevel) // Should not panic
}

func TestConsoleLogger_JSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewConsoleLogger(
		WithFormat(JSONFormat),
		WithWriter(&buf),
		WithLevel(DebugLevel),
	)

	logger.Info("test message", Fields("key1", "value1", "key2", 42)...)

	output := buf.String()
	if output == "" {
		t.Fatal("No output written")
	}

	// Parse JSON
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry["level"] != "INFO" {
		t.Errorf("level = %v, want INFO", entry["level"])
	}

	if entry["message"] != "test message" {
		t.Errorf("message = %v, want 'test message'", entry["message"])
	}

	if entry["key1"] != "value1" {
		t.Errorf("key1 = %v, want 'value1'", entry["key1"])
	}

	if entry["key2"] != float64(42) { // JSON numbers are float64
		t.Errorf("key2 = %v, want 42", entry["key2"])
	}
}

func TestConsoleLogger_Text(t *testing.T) {
	var buf bytes.Buffer
	logger := NewConsoleLogger(
		WithFormat(TextFormat),
		WithWriter(&buf),
		WithLevel(DebugLevel),
	)

	logger.Warn("warning message", Fields("user", "alice")...)

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Error("Output should contain 'WARN'")
	}
	if !strings.Contains(output, "warning message") {
		t.Error("Output should contain 'warning message'")
	}
	if !strings.Contains(output, "user=alice") {
		t.Error("Output should contain 'user=alice'")
	}
}

func TestConsoleLogger_Level(t *testing.T) {
	var buf bytes.Buffer
	logger := NewConsoleLogger(
		WithFormat(JSONFormat),
		WithWriter(&buf),
		WithErrorWriter(&buf), // Use same buffer for error messages
		WithLevel(WarnLevel),
	)

	// These should not be logged
	logger.Debug("debug message")
	logger.Info("info message")

	if buf.Len() > 0 {
		t.Error("Debug and Info should not be logged at Warn level")
	}

	// These should be logged
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be logged")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should be logged")
	}
}

func TestConsoleLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewConsoleLogger(
		WithFormat(JSONFormat),
		WithWriter(&buf),
		WithLevel(DebugLevel),
	)

	logger.Debug("debug1")
	logger.SetLevel(InfoLevel)
	logger.Debug("debug2")
	logger.Info("info1")

	output := buf.String()
	if !strings.Contains(output, "debug1") {
		t.Error("debug1 should be logged")
	}
	if strings.Contains(output, "debug2") {
		t.Error("debug2 should not be logged after SetLevel")
	}
	if !strings.Contains(output, "info1") {
		t.Error("info1 should be logged")
	}
}

func TestConsoleLogger_With(t *testing.T) {
	var buf bytes.Buffer
	logger := NewConsoleLogger(
		WithFormat(JSONFormat),
		WithWriter(&buf),
		WithLevel(DebugLevel),
	)

	child := logger.With(Fields("request_id", "123")...)
	child.Info("test message", Fields("action", "create")...)

	output := buf.String()
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry["request_id"] != "123" {
		t.Error("Base field 'request_id' should be included")
	}
	if entry["action"] != "create" {
		t.Error("Message field 'action' should be included")
	}
}

func TestConsoleLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewConsoleLogger(
		WithFormat(JSONFormat),
		WithWriter(&buf),
		WithLevel(DebugLevel),
	)

	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	child := logger.WithContext(ctx)
	child.Info("test message")

	output := buf.String()
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry["request_id"] != "req-123" {
		t.Error("Context field 'request_id' should be included")
	}
}

func TestConsoleLogger_ErrorWriter(t *testing.T) {
	var stdBuf bytes.Buffer
	var errBuf bytes.Buffer
	logger := NewConsoleLogger(
		WithFormat(TextFormat),
		WithWriter(&stdBuf),
		WithErrorWriter(&errBuf),
		WithLevel(DebugLevel),
	)

	logger.Info("info message")
	logger.Error("error message")

	if !strings.Contains(stdBuf.String(), "info message") {
		t.Error("Info message should go to standard writer")
	}
	if !strings.Contains(errBuf.String(), "error message") {
		t.Error("Error message should go to error writer")
	}
	if strings.Contains(errBuf.String(), "info message") {
		t.Error("Info message should not go to error writer")
	}
}

func TestBufferedLogger(t *testing.T) {
	logger := NewBufferedLogger()

	logger.Debug("debug message")
	logger.Info("info message", Fields("key", "value")...)
	logger.Warn("warn message")
	logger.Error("error message")

	entries := logger.Entries()
	if len(entries) != 4 {
		t.Fatalf("Expected 4 entries, got %d", len(entries))
	}

	// Check first entry
	if entries[0].Level != DebugLevel {
		t.Errorf("Entry 0 level = %v, want DebugLevel", entries[0].Level)
	}
	if entries[0].Message != "debug message" {
		t.Errorf("Entry 0 message = %v, want 'debug message'", entries[0].Message)
	}

	// Check second entry with fields
	if entries[1].Level != InfoLevel {
		t.Errorf("Entry 1 level = %v, want InfoLevel", entries[1].Level)
	}
	if len(entries[1].Fields) != 1 {
		t.Errorf("Entry 1 should have 1 field, got %d", len(entries[1].Fields))
	}
	if entries[1].Fields[0].Key != "key" || entries[1].Fields[0].Value != "value" {
		t.Error("Entry 1 field mismatch")
	}
}

func TestBufferedLogger_Count(t *testing.T) {
	logger := NewBufferedLogger()

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error1")
	logger.Error("error2")

	if count := logger.Count(); count != 5 {
		t.Errorf("Count() = %d, want 5", count)
	}

	if count := logger.CountByLevel(ErrorLevel); count != 2 {
		t.Errorf("CountByLevel(ErrorLevel) = %d, want 2", count)
	}

	if count := logger.CountByLevel(InfoLevel); count != 1 {
		t.Errorf("CountByLevel(InfoLevel) = %d, want 1", count)
	}
}

func TestBufferedLogger_Clear(t *testing.T) {
	logger := NewBufferedLogger()

	logger.Info("test1")
	logger.Info("test2")

	if logger.Count() != 2 {
		t.Fatal("Expected 2 entries")
	}

	logger.Clear()

	if logger.Count() != 0 {
		t.Error("Count() should be 0 after Clear()")
	}
}

func TestBufferedLogger_Level(t *testing.T) {
	logger := NewBufferedLogger()
	logger.SetLevel(WarnLevel)

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	if count := logger.Count(); count != 2 {
		t.Errorf("Count() = %d, want 2 (only Warn and Error)", count)
	}
}

func BenchmarkConsoleLogger_JSON(b *testing.B) {
	var buf bytes.Buffer
	logger := NewConsoleLogger(
		WithFormat(JSONFormat),
		WithWriter(&buf),
		WithLevel(InfoLevel),
	)

	fields := Fields("request_id", "123", "user", "alice", "duration", 42.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", fields...)
	}
}

func BenchmarkConsoleLogger_Text(b *testing.B) {
	var buf bytes.Buffer
	logger := NewConsoleLogger(
		WithFormat(TextFormat),
		WithWriter(&buf),
		WithLevel(InfoLevel),
	)

	fields := Fields("request_id", "123", "user", "alice", "duration", 42.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", fields...)
	}
}

func BenchmarkBufferedLogger(b *testing.B) {
	logger := NewBufferedLogger()
	fields := Fields("request_id", "123", "user", "alice", "duration", 42.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", fields...)
	}
}
