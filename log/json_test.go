package log

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/spcent/plumego/contract"
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
		Output:           &buf,
		Level:            INFO,
		RespectVerbosity: true,
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
		Output:           &buf,
		Level:            INFO,
		RespectVerbosity: true,
	})

	ctx := context.Background()
	requestID := "req-123"
	ctx = contract.WithRequestID(ctx, requestID)

	logger.InfoCtx(ctx, "test message", Fields{"key": "value"})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["request_id"] != requestID {
		t.Errorf("expected request_id %q, got %v", requestID, entry["request_id"])
	}

	if entry["msg"] != "test message" {
		t.Errorf("expected msg 'test message', got %v", entry["msg"])
	}
}

func TestJSONLogger_ContextLevels(t *testing.T) {
	tests := []struct {
		name      string
		logFunc   func(logger *JSONLogger, ctx context.Context)
		wantLevel string
	}{
		{
			name: "debug ctx",
			logFunc: func(l *JSONLogger, ctx context.Context) {
				l.DebugCtx(ctx, "debug msg", nil)
			},
			wantLevel: "DEBUG",
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
				Level:  DEBUG,
			})

			ctx := context.Background()
			ctx = contract.WithRequestID(ctx, "test-request")

			tt.logFunc(logger, ctx)

			var entry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("failed to parse JSON output: %v", err)
			}

			if entry["level"] != tt.wantLevel {
				t.Errorf("expected level %s, got %v", tt.wantLevel, entry["level"])
			}

			if entry["request_id"] != "test-request" {
				t.Errorf("expected request_id 'test-request', got %v", entry["request_id"])
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

func TestJSONLogger_DebugVerbosityGate(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(JSONLoggerConfig{
		Output:           &buf,
		Level:            DEBUG,
		RespectVerbosity: true,
		Verbosity:        0,
	})

	logger.Debug("hidden debug", nil)
	if strings.TrimSpace(buf.String()) != "" {
		t.Fatalf("expected debug log to be filtered when verbosity is 0")
	}

	logger.verbosity = 1
	logger.Debug("visible debug", nil)

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if entry["msg"] != "visible debug" {
		t.Fatalf("expected visible debug message, got %v", entry["msg"])
	}
	if entry["level"] != "DEBUG" {
		t.Fatalf("expected DEBUG level, got %v", entry["level"])
	}
}

func TestJSONLogger_ReservedFieldsCannotOverrideCoreKeys(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(JSONLoggerConfig{
		Output: &buf,
		Level:  INFO,
		Fields: Fields{
			"level": "OVERRIDE",
			"msg":   "override",
			"time":  "not-time",
		},
	})

	ctx := contract.WithRequestID(context.Background(), "req-from-ctx")
	logger.InfoCtx(ctx, "actual message", Fields{
		"level":      "SHOULD_NOT_APPLY",
		"msg":        "SHOULD_NOT_APPLY",
		"time":       "SHOULD_NOT_APPLY",
		"request_id": "SHOULD_NOT_APPLY",
	})

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if entry["level"] != "INFO" {
		t.Fatalf("expected reserved level to remain INFO, got %v", entry["level"])
	}
	if entry["msg"] != "actual message" {
		t.Fatalf("expected reserved msg to remain actual message, got %v", entry["msg"])
	}
	if entry["time"] == "SHOULD_NOT_APPLY" || entry["time"] == "not-time" {
		t.Fatalf("expected reserved time to be generated by logger, got %v", entry["time"])
	}
	if entry["request_id"] != "req-from-ctx" {
		t.Fatalf("expected request_id from context, got %v", entry["request_id"])
	}
}

func TestJSONLogger_WithFieldsSharesWriterLock(t *testing.T) {
	var buf bytes.Buffer
	base := NewJSONLogger(JSONLoggerConfig{
		Output: &buf,
		Level:  INFO,
	})
	child := base.WithFields(Fields{"scope": "child"})

	const n = 200
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			base.Info("base", Fields{"idx": i})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			child.Info("child", Fields{"idx": i})
		}
	}()

	wg.Wait()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2*n {
		t.Fatalf("expected %d lines, got %d", 2*n, len(lines))
	}
	for i, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("line %d should be valid JSON, err=%v line=%q", i, err, line)
		}
	}
}
