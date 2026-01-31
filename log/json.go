package glog

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// JSONLogger implements StructuredLogger with JSON output format.
// It's thread-safe and suitable for structured logging in production environments.
type JSONLogger struct {
	mu     sync.Mutex
	output io.Writer
	level  Level
	fields Fields
}

// JSONLoggerConfig configures a JSONLogger instance.
type JSONLoggerConfig struct {
	// Output destination (defaults to os.Stdout)
	Output io.Writer
	// Minimum log level (defaults to INFO)
	Level Level
	// Default fields to include in every log entry
	Fields Fields
}

// NewJSONLogger creates a new JSONLogger with the given configuration.
func NewJSONLogger(config JSONLoggerConfig) *JSONLogger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.Fields == nil {
		config.Fields = make(Fields)
	}
	return &JSONLogger{
		output: config.Output,
		level:  config.Level,
		fields: config.Fields,
	}
}

// WithFields returns a new logger with additional fields.
func (l *JSONLogger) WithFields(fields Fields) StructuredLogger {
	merged := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &JSONLogger{
		output: l.output,
		level:  l.level,
		fields: merged,
	}
}

// Debug logs a debug message with optional fields.
func (l *JSONLogger) Debug(msg string, fields Fields) {
	l.log(INFO, msg, fields, "") // Map DEBUG to INFO level with verbosity check
}

// Info logs an info message with optional fields.
func (l *JSONLogger) Info(msg string, fields Fields) {
	l.log(INFO, msg, fields, "")
}

// Warn logs a warning message with optional fields.
func (l *JSONLogger) Warn(msg string, fields Fields) {
	l.log(WARNING, msg, fields, "")
}

// Error logs an error message with optional fields.
func (l *JSONLogger) Error(msg string, fields Fields) {
	l.log(ERROR, msg, fields, "")
}

// DebugCtx logs a debug message with context and optional fields.
func (l *JSONLogger) DebugCtx(ctx context.Context, msg string, fields Fields) {
	l.logCtx(ctx, INFO, msg, fields)
}

// InfoCtx logs an info message with context and optional fields.
func (l *JSONLogger) InfoCtx(ctx context.Context, msg string, fields Fields) {
	l.logCtx(ctx, INFO, msg, fields)
}

// WarnCtx logs a warning message with context and optional fields.
func (l *JSONLogger) WarnCtx(ctx context.Context, msg string, fields Fields) {
	l.logCtx(ctx, WARNING, msg, fields)
}

// ErrorCtx logs an error message with context and optional fields.
func (l *JSONLogger) ErrorCtx(ctx context.Context, msg string, fields Fields) {
	l.logCtx(ctx, ERROR, msg, fields)
}

// log is the internal logging method without context.
func (l *JSONLogger) log(level Level, msg string, fields Fields, traceID string) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := l.buildEntry(level, msg, fields, traceID)
	l.writeEntry(entry)
}

// logCtx is the internal logging method with context support.
func (l *JSONLogger) logCtx(ctx context.Context, level Level, msg string, fields Fields) {
	if level < l.level {
		return
	}

	// Extract trace ID from context
	traceID := TraceIDFromContext(ctx)

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := l.buildEntry(level, msg, fields, traceID)
	l.writeEntry(entry)
}

// buildEntry constructs the log entry map.
func (l *JSONLogger) buildEntry(level Level, msg string, fields Fields, traceID string) map[string]any {
	entry := make(map[string]any)
	entry["time"] = time.Now().UTC().Format(time.RFC3339Nano)
	entry["level"] = levelNames[level]
	entry["msg"] = msg

	// Add trace ID if present
	if traceID != "" {
		entry["trace_id"] = traceID
	}

	// Add default fields
	for k, v := range l.fields {
		entry[k] = v
	}

	// Add message-specific fields (can override defaults)
	for k, v := range fields {
		entry[k] = v
	}

	return entry
}

// writeEntry marshals and writes the log entry.
func (l *JSONLogger) writeEntry(entry map[string]any) {
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple error message if marshaling fails
		data = []byte(`{"level":"ERROR","msg":"failed to marshal log entry"}`)
	}
	l.output.Write(data)
	l.output.Write([]byte("\n"))
}

// Start implements the Lifecycle interface (no-op for JSONLogger).
func (l *JSONLogger) Start(ctx context.Context) error {
	return nil
}

// Stop implements the Lifecycle interface (flushes if output supports it).
func (l *JSONLogger) Stop(ctx context.Context) error {
	if syncer, ok := l.output.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}
