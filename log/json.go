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
	mu               sync.Mutex
	writeErrOnce     sync.Once
	output           io.Writer
	level            Level
	verbosity        int
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
	// Verbosity controls local debug gating when RespectVerbosity is true.
	Verbosity int
}

// NewJSONLogger creates a new JSONLogger with the given configuration.
func NewJSONLogger(config JSONLoggerConfig) *JSONLogger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	return &JSONLogger{
		output:           config.Output,
		level:            config.Level,
		verbosity:        config.Verbosity,
		fields:           cloneFields(config.Fields),
		respectVerbosity: config.RespectVerbosity,
	}
}

// WithFields returns a new logger with additional fields merged in.
// Each derived logger gets its own independent write-error tracker.
func (l *JSONLogger) WithFields(fields Fields) StructuredLogger {
	l.mu.Lock()
	output := l.output
	level := l.level
	verbosity := l.verbosity
	respectVerbosity := l.respectVerbosity
	merged := mergeFields(l.fields, fields)
	l.mu.Unlock()

	return &JSONLogger{
		output:           output,
		level:            level,
		verbosity:        verbosity,
		fields:           merged,
		respectVerbosity: respectVerbosity,
		// writeErrOnce is zero-value (fresh) for every derived logger
	}
}

// Debug logs a debug message with optional fields.
func (l *JSONLogger) Debug(msg string, fields Fields) {
	if l.respectVerbosity && !l.vAt(1) {
		return
	}
	l.log(DEBUG, msg, fields, nil)
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

// Fatal logs a fatal message then calls os.Exit(1).
func (l *JSONLogger) Fatal(msg string, fields Fields) {
	l.log(FATAL, msg, fields, nil)
	os.Exit(1)
}

// DebugCtx logs a debug message with context and optional fields.
func (l *JSONLogger) DebugCtx(ctx context.Context, msg string, fields Fields) {
	if l.respectVerbosity && !l.vAt(1) {
		return
	}
	l.log(DEBUG, msg, fields, ctx)
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

// FatalCtx logs a fatal message with context then calls os.Exit(1).
func (l *JSONLogger) FatalCtx(ctx context.Context, msg string, fields Fields) {
	l.log(FATAL, msg, fields, ctx)
	os.Exit(1)
}

func (l *JSONLogger) log(level Level, msg string, fields Fields, ctx context.Context) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := l.buildEntry(level, msg, fields, ctx)
	l.writeEntry(entry)
}

// buildEntry constructs the log entry map.
func (l *JSONLogger) buildEntry(level Level, msg string, fields Fields, ctx context.Context) map[string]any {
	combined := mergeFields(l.fields, fields)
	entry := make(map[string]any, len(combined)+4)
	for k, v := range combined {
		entry[k] = v
	}

	// Reserved keys are always controlled by logger internals.
	entry["time"] = time.Now().UTC().Format(time.RFC3339Nano)
	entry["level"] = levelName(level)
	entry["msg"] = msg
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		entry["trace_id"] = traceID
	}

	return entry
}

// writeEntry marshals and writes the log entry.
func (l *JSONLogger) writeEntry(entry map[string]any) {
	data, err := json.Marshal(entry)
	if err != nil {
		data = []byte(`{"level":"ERROR","msg":"failed to marshal log entry"}`)
	}
	if err := writeFull(l.output, data); err != nil {
		l.reportWriteError(err)
		return
	}
	if err := writeFull(l.output, []byte("\n")); err != nil {
		l.reportWriteError(err)
	}
}

// Start implements the Lifecycle interface (no-op for JSONLogger).
func (l *JSONLogger) Start(ctx context.Context) error {
	return nil
}

// Stop implements the Lifecycle interface (flushes if output supports it).
func (l *JSONLogger) Stop(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if syncer, ok := l.output.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}

func (l *JSONLogger) vAt(level int) bool {
	return level <= l.verbosity
}

func (l *JSONLogger) reportWriteError(err error) {
	l.writeErrOnce.Do(func() {
		_, _ = os.Stderr.WriteString("json logger: failed to write log output: " + err.Error() + "\n")
	})
}
