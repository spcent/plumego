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
	mu               *sync.Mutex
	output           io.Writer
	level            Level
	fields           Fields
	respectVerbosity bool
}

// JSONLoggerConfig configures a JSONLogger instance.
type JSONLoggerConfig struct {
	// Output destination (defaults to os.Stdout)
	Output io.Writer
	// Minimum log level (defaults to INFO)
	Level Level
	// Default fields to include in every log entry
	Fields Fields
	// RespectVerbosity applies V(1) filtering to Debug/DebugCtx when enabled.
	// Default false keeps backward-compatible behavior.
	RespectVerbosity bool
}

// NewJSONLogger creates a new JSONLogger with the given configuration.
func NewJSONLogger(config JSONLoggerConfig) *JSONLogger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	return &JSONLogger{
		mu:               &sync.Mutex{},
		output:           config.Output,
		level:            config.Level,
		fields:           cloneFields(config.Fields),
		respectVerbosity: config.RespectVerbosity,
	}
}

// WithFields returns a new logger with additional fields.
func (l *JSONLogger) WithFields(fields Fields) StructuredLogger {
	mu := l.mu
	if mu == nil {
		mu = &sync.Mutex{}
	}
	return &JSONLogger{
		mu:               mu,
		output:           l.output,
		level:            l.level,
		fields:           mergeFields(l.fields, fields),
		respectVerbosity: l.respectVerbosity,
	}
}

// Debug logs a debug message with optional fields.
func (l *JSONLogger) Debug(msg string, fields Fields) {
	if l.respectVerbosity && !std.vAt(1, 3) {
		return
	}
	// Keep backward compatibility with glog adapter where debug is emitted as INFO.
	l.log(INFO, msg, fields, nil)
}

// Info logs an info message with optional fields.
func (l *JSONLogger) Info(msg string, fields Fields) {
	l.log(INFO, msg, fields, nil)
}

// Warn logs a warning message with optional fields.
func (l *JSONLogger) Warn(msg string, fields Fields) {
	l.log(WARNING, msg, fields, nil)
}

// Error logs an error message with optional fields.
func (l *JSONLogger) Error(msg string, fields Fields) {
	l.log(ERROR, msg, fields, nil)
}

// DebugCtx logs a debug message with context and optional fields.
func (l *JSONLogger) DebugCtx(ctx context.Context, msg string, fields Fields) {
	if l.respectVerbosity && !std.vAt(1, 3) {
		return
	}
	// Keep backward compatibility with glog adapter where debug is emitted as INFO.
	l.log(INFO, msg, fields, ctx)
}

// InfoCtx logs an info message with context and optional fields.
func (l *JSONLogger) InfoCtx(ctx context.Context, msg string, fields Fields) {
	l.log(INFO, msg, fields, ctx)
}

// WarnCtx logs a warning message with context and optional fields.
func (l *JSONLogger) WarnCtx(ctx context.Context, msg string, fields Fields) {
	l.log(WARNING, msg, fields, ctx)
}

// ErrorCtx logs an error message with context and optional fields.
func (l *JSONLogger) ErrorCtx(ctx context.Context, msg string, fields Fields) {
	l.log(ERROR, msg, fields, ctx)
}

func (l *JSONLogger) log(level Level, msg string, fields Fields, ctx context.Context) {
	if level < l.level {
		return
	}

	mu := l.mu
	if mu == nil {
		mu = &sync.Mutex{}
		l.mu = mu
	}
	mu.Lock()
	defer mu.Unlock()

	entry := l.buildEntry(level, msg, fields, ctx)
	l.writeEntry(entry)
}

// buildEntry constructs the log entry map.
func (l *JSONLogger) buildEntry(level Level, msg string, fields Fields, ctx context.Context) map[string]any {
	entry := make(map[string]any)
	entry["time"] = time.Now().UTC().Format(time.RFC3339Nano)
	entry["level"] = levelNames[level]
	entry["msg"] = msg

	// Add trace ID if present
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		entry["trace_id"] = traceID
	}

	combined := mergeFields(l.fields, fields)
	for k, v := range combined {
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
	mu := l.mu
	if mu == nil {
		mu = &sync.Mutex{}
		l.mu = mu
	}
	mu.Lock()
	defer mu.Unlock()

	if syncer, ok := l.output.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}
