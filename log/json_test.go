package glog

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewJSONLogger(t *testing.T) {
	t.Run("with defaults", func(t *testing.T) {
		logger := NewJSONLogger(JSONLoggerConfig{})
		if logger == nil {
			t.Error("expected logger to be created")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewJSONLogger(JSONLoggerConfig{
			Output: &buf,
			Level:  WARNING,
			Fields: Fields{"app": "test"},
		})

		if logger == nil {
			t.Error("expected logger to be created")
		}
	})
}

func TestJSONLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(JSONLoggerConfig{
		Output: &buf,
		Level:  INFO,
	})

	logger.Info("test message", Fields{"key": "value"})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["level"] != "INFO" {
		t.Errorf("expected level INFO, got %v", entry["level"])
	}

	if entry["msg"] != "test message" {
		t.Errorf("expected msg 'test message', got %v", entry["msg"])
	}

	if entry["key"] != "value" {
		t.Errorf("expected key 'value', got %v", entry["key"])
	}

	if _, ok := entry["time"]; !ok {
		t.Error("expected time field to be present")
	}
}

func TestJSONLogger_Levels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel Level
		logFunc  func(logger *JSONLogger)
		want     string
	}{
		{
			name:     "info level",
			logLevel: INFO,
			logFunc: func(l *JSONLogger) {
				l.Info("info msg", nil)
			},
			want: "INFO",
		},
		{
			name:     "warning level",
			logLevel: INFO,
			logFunc: func(l *JSONLogger) {
				l.Warn("warn msg", nil)
			},
			want: "WARN",
		},
		{
			name:     "error level",
			logLevel: INFO,
			logFunc: func(l *JSONLogger) {
				l.Error("error msg", nil)
			},
			want: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewJSONLogger(JSONLoggerConfig{
				Output: &buf,
				Level:  tt.logLevel,
			})

			tt.logFunc(logger)

			var entry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("failed to parse JSON output: %v", err)
			}

			if entry["level"] != tt.want {
				t.Errorf("expected level %s, got %v", tt.want, entry["level"])
			}
		})
	}
}

func TestJSONLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(JSONLoggerConfig{
		Output: &buf,
		Level:  INFO,
		Fields: Fields{"app": "myapp"},
	})

	childLogger := logger.WithFields(Fields{"request_id": "123"})
	childLogger.Info("test", Fields{"status": "ok"})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	// Should have all fields
	if entry["app"] != "myapp" {
		t.Errorf("expected app 'myapp', got %v", entry["app"])
	}
	if entry["request_id"] != "123" {
		t.Errorf("expected request_id '123', got %v", entry["request_id"])
	}
	if entry["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", entry["status"])
	}
}

func TestJSONLogger_InfoCtx(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(JSONLoggerConfig{
		Output: &buf,
		Level:  INFO,
	})

	ctx := context.Background()
	traceID := "trace-123"
	ctx = WithTraceID(ctx, traceID)

	logger.InfoCtx(ctx, "test message", Fields{"key": "value"})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["trace_id"] != traceID {
		t.Errorf("expected trace_id %q, got %v", traceID, entry["trace_id"])
	}

	if entry["msg"] != "test message" {
		t.Errorf("expected msg 'test message', got %v", entry["msg"])
	}
}

func TestJSONLogger_ContextLevels(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(logger *JSONLogger, ctx context.Context)
		wantLevel string
	}{
		{
			name: "debug ctx",
			logFunc: func(l *JSONLogger, ctx context.Context) {
				l.DebugCtx(ctx, "debug msg", nil)
			},
			wantLevel: "INFO",
		},
		{
			name: "info ctx",
			logFunc: func(l *JSONLogger, ctx context.Context) {
				l.InfoCtx(ctx, "info msg", nil)
			},
			wantLevel: "INFO",
		},
		{
			name: "warn ctx",
			logFunc: func(l *JSONLogger, ctx context.Context) {
				l.WarnCtx(ctx, "warn msg", nil)
			},
			wantLevel: "WARN",
		},
		{
			name: "error ctx",
			logFunc: func(l *JSONLogger, ctx context.Context) {
				l.ErrorCtx(ctx, "error msg", nil)
			},
			wantLevel: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewJSONLogger(JSONLoggerConfig{
				Output: &buf,
				Level:  INFO,
			})

			ctx := context.Background()
			ctx = WithTraceID(ctx, "test-trace")

			tt.logFunc(logger, ctx)

			var entry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("failed to parse JSON output: %v", err)
			}

			if entry["level"] != tt.wantLevel {
				t.Errorf("expected level %s, got %v", tt.wantLevel, entry["level"])
			}

			if entry["trace_id"] != "test-trace" {
				t.Errorf("expected trace_id 'test-trace', got %v", entry["trace_id"])
			}
		})
	}
}

func TestJSONLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(JSONLoggerConfig{
		Output: &buf,
		Level:  ERROR, // Only log ERROR and above
	})

	logger.Info("should not appear", nil)
	logger.Warn("should not appear", nil)
	logger.Error("should appear", nil)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should only have one line (the ERROR)
	if len(lines) != 1 {
		t.Errorf("expected 1 log line, got %d", len(lines))
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["level"] != "ERROR" {
		t.Errorf("expected level ERROR, got %v", entry["level"])
	}
}

func TestJSONLogger_FieldOverride(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(JSONLoggerConfig{
		Output: &buf,
		Level:  INFO,
		Fields: Fields{"env": "dev", "version": "1.0"},
	})

	// Override 'env' field
	logger.Info("test", Fields{"env": "prod", "request": "123"})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	// env should be overridden to 'prod'
	if entry["env"] != "prod" {
		t.Errorf("expected env 'prod', got %v", entry["env"])
	}

	// version should still be '1.0'
	if entry["version"] != "1.0" {
		t.Errorf("expected version '1.0', got %v", entry["version"])
	}

	// request should be '123'
	if entry["request"] != "123" {
		t.Errorf("expected request '123', got %v", entry["request"])
	}
}

func TestJSONLogger_NilFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(JSONLoggerConfig{
		Output: &buf,
		Level:  INFO,
	})

	// Should not panic with nil fields
	logger.Info("test message", nil)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["msg"] != "test message" {
		t.Errorf("expected msg 'test message', got %v", entry["msg"])
	}
}
