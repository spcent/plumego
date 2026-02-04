// Package logging provides structured logging interfaces for the AI Agent Gateway.
//
// Design Philosophy:
// - Interface-first design for flexibility
// - Zero third-party dependencies by default
// - Structured field support (key-value pairs)
// - Context-aware logging
package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Field represents a structured log field (key-value pair).
type Field struct {
	Key   string
	Value interface{}
}

// Fields creates a slice of fields from key-value pairs.
func Fields(keyValues ...interface{}) []Field {
	if len(keyValues)%2 != 0 {
		panic("Fields requires an even number of arguments")
	}

	fields := make([]Field, 0, len(keyValues)/2)
	for i := 0; i < len(keyValues); i += 2 {
		key, ok := keyValues[i].(string)
		if !ok {
			panic("Field keys must be strings")
		}
		fields = append(fields, Field{Key: key, Value: keyValues[i+1]})
	}
	return fields
}

// Logger is the main interface for structured logging.
type Logger interface {
	// Debug logs a debug-level message
	Debug(msg string, fields ...Field)

	// Info logs an info-level message
	Info(msg string, fields ...Field)

	// Warn logs a warning-level message
	Warn(msg string, fields ...Field)

	// Error logs an error-level message
	Error(msg string, fields ...Field)

	// Fatal logs a fatal-level message and exits
	Fatal(msg string, fields ...Field)

	// With creates a child logger with additional fields
	With(fields ...Field) Logger

	// WithContext creates a child logger with context fields
	WithContext(ctx context.Context) Logger

	// SetLevel sets the minimum log level
	SetLevel(level Level)
}

// NoOpLogger is a logger that does nothing (for disabled logging).
type NoOpLogger struct{}

func (n *NoOpLogger) Debug(msg string, fields ...Field)          {}
func (n *NoOpLogger) Info(msg string, fields ...Field)           {}
func (n *NoOpLogger) Warn(msg string, fields ...Field)           {}
func (n *NoOpLogger) Error(msg string, fields ...Field)          {}
func (n *NoOpLogger) Fatal(msg string, fields ...Field)          { os.Exit(1) }
func (n *NoOpLogger) With(fields ...Field) Logger                { return n }
func (n *NoOpLogger) WithContext(ctx context.Context) Logger     { return n }
func (n *NoOpLogger) SetLevel(level Level)                       {}

// ConsoleLogger is a structured logger that writes to stdout/stderr.
type ConsoleLogger struct {
	level      Level
	baseFields []Field
	mu         sync.RWMutex
	writer     io.Writer
	errWriter  io.Writer
	format     LogFormat
}

// LogFormat determines the output format.
type LogFormat int

const (
	// JSONFormat outputs logs as JSON (machine-readable)
	JSONFormat LogFormat = iota
	// TextFormat outputs logs as human-readable text
	TextFormat
)

// ConsoleLoggerOption configures a ConsoleLogger.
type ConsoleLoggerOption func(*ConsoleLogger)

// WithLevel sets the minimum log level.
func WithLevel(level Level) ConsoleLoggerOption {
	return func(l *ConsoleLogger) {
		l.level = level
	}
}

// WithFormat sets the output format.
func WithFormat(format LogFormat) ConsoleLoggerOption {
	return func(l *ConsoleLogger) {
		l.format = format
	}
}

// WithWriter sets the output writer for non-error logs.
func WithWriter(w io.Writer) ConsoleLoggerOption {
	return func(l *ConsoleLogger) {
		l.writer = w
	}
}

// WithErrorWriter sets the output writer for error logs.
func WithErrorWriter(w io.Writer) ConsoleLoggerOption {
	return func(l *ConsoleLogger) {
		l.errWriter = w
	}
}

// NewConsoleLogger creates a new console logger.
func NewConsoleLogger(opts ...ConsoleLoggerOption) *ConsoleLogger {
	logger := &ConsoleLogger{
		level:      InfoLevel,
		baseFields: []Field{},
		writer:     os.Stdout,
		errWriter:  os.Stderr,
		format:     JSONFormat,
	}

	for _, opt := range opts {
		opt(logger)
	}

	return logger
}

// Debug implements Logger
func (cl *ConsoleLogger) Debug(msg string, fields ...Field) {
	cl.log(DebugLevel, msg, fields)
}

// Info implements Logger
func (cl *ConsoleLogger) Info(msg string, fields ...Field) {
	cl.log(InfoLevel, msg, fields)
}

// Warn implements Logger
func (cl *ConsoleLogger) Warn(msg string, fields ...Field) {
	cl.log(WarnLevel, msg, fields)
}

// Error implements Logger
func (cl *ConsoleLogger) Error(msg string, fields ...Field) {
	cl.log(ErrorLevel, msg, fields)
}

// Fatal implements Logger
func (cl *ConsoleLogger) Fatal(msg string, fields ...Field) {
	cl.log(FatalLevel, msg, fields)
	os.Exit(1)
}

// With implements Logger
func (cl *ConsoleLogger) With(fields ...Field) Logger {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	newFields := make([]Field, len(cl.baseFields)+len(fields))
	copy(newFields, cl.baseFields)
	copy(newFields[len(cl.baseFields):], fields)

	return &ConsoleLogger{
		level:      cl.level,
		baseFields: newFields,
		writer:     cl.writer,
		errWriter:  cl.errWriter,
		format:     cl.format,
	}
}

// WithContext implements Logger
func (cl *ConsoleLogger) WithContext(ctx context.Context) Logger {
	// Extract common context values
	fields := []Field{}

	// Check for common context keys (request ID, trace ID, etc.)
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, Field{Key: "request_id", Value: requestID})
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, Field{Key: "trace_id", Value: traceID})
	}

	return cl.With(fields...)
}

// SetLevel implements Logger
func (cl *ConsoleLogger) SetLevel(level Level) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.level = level
}

// log writes a log message with the given level and fields
func (cl *ConsoleLogger) log(level Level, msg string, fields []Field) {
	cl.mu.RLock()
	minLevel := cl.level
	cl.mu.RUnlock()

	if level < minLevel {
		return
	}

	// Combine base fields and message fields
	allFields := make([]Field, 0, len(cl.baseFields)+len(fields))
	allFields = append(allFields, cl.baseFields...)
	allFields = append(allFields, fields...)

	// Choose writer
	writer := cl.writer
	if level >= ErrorLevel {
		writer = cl.errWriter
	}

	// Format and write
	switch cl.format {
	case JSONFormat:
		cl.writeJSON(writer, level, msg, allFields)
	case TextFormat:
		cl.writeText(writer, level, msg, allFields)
	}
}

// writeJSON writes a log message in JSON format
func (cl *ConsoleLogger) writeJSON(w io.Writer, level Level, msg string, fields []Field) {
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"level":     level.String(),
		"message":   msg,
	}

	for _, field := range fields {
		entry[field.Key] = field.Value
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple output
		fmt.Fprintf(w, `{"timestamp":"%s","level":"%s","message":"%s","error":"failed to marshal fields"}%s`,
			time.Now().UTC().Format(time.RFC3339Nano), level.String(), msg, "\n")
		return
	}

	w.Write(data)
	w.Write([]byte("\n"))
}

// writeText writes a log message in human-readable text format
func (cl *ConsoleLogger) writeText(w io.Writer, level Level, msg string, fields []Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(w, "[%s] %s: %s", timestamp, level.String(), msg)

	if len(fields) > 0 {
		fmt.Fprint(w, " {")
		for i, field := range fields {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprintf(w, "%s=%v", field.Key, field.Value)
		}
		fmt.Fprint(w, "}")
	}

	fmt.Fprintln(w)
}

// BufferedLogger is an in-memory logger for testing.
type BufferedLogger struct {
	entries []LogEntry
	level   Level
	mu      sync.RWMutex
}

// LogEntry represents a single log entry.
type LogEntry struct {
	Level     Level
	Message   string
	Fields    []Field
	Timestamp time.Time
}

// NewBufferedLogger creates a new buffered logger.
func NewBufferedLogger() *BufferedLogger {
	return &BufferedLogger{
		entries: []LogEntry{},
		level:   DebugLevel,
	}
}

// Debug implements Logger
func (bl *BufferedLogger) Debug(msg string, fields ...Field) {
	bl.log(DebugLevel, msg, fields)
}

// Info implements Logger
func (bl *BufferedLogger) Info(msg string, fields ...Field) {
	bl.log(InfoLevel, msg, fields)
}

// Warn implements Logger
func (bl *BufferedLogger) Warn(msg string, fields ...Field) {
	bl.log(WarnLevel, msg, fields)
}

// Error implements Logger
func (bl *BufferedLogger) Error(msg string, fields ...Field) {
	bl.log(ErrorLevel, msg, fields)
}

// Fatal implements Logger
func (bl *BufferedLogger) Fatal(msg string, fields ...Field) {
	bl.log(FatalLevel, msg, fields)
}

// With implements Logger
func (bl *BufferedLogger) With(fields ...Field) Logger {
	// For buffered logger, we just return the same logger
	// In a more advanced implementation, this could create a child logger
	return bl
}

// WithContext implements Logger
func (bl *BufferedLogger) WithContext(ctx context.Context) Logger {
	return bl
}

// SetLevel implements Logger
func (bl *BufferedLogger) SetLevel(level Level) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.level = level
}

// log adds an entry to the buffer
func (bl *BufferedLogger) log(level Level, msg string, fields []Field) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if level < bl.level {
		return
	}

	bl.entries = append(bl.entries, LogEntry{
		Level:     level,
		Message:   msg,
		Fields:    fields,
		Timestamp: time.Now(),
	})
}

// Entries returns all logged entries
func (bl *BufferedLogger) Entries() []LogEntry {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	result := make([]LogEntry, len(bl.entries))
	copy(result, bl.entries)
	return result
}

// Clear removes all entries
func (bl *BufferedLogger) Clear() {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.entries = []LogEntry{}
}

// Count returns the number of logged entries
func (bl *BufferedLogger) Count() int {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return len(bl.entries)
}

// CountByLevel returns the number of entries for a specific level
func (bl *BufferedLogger) CountByLevel(level Level) int {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	count := 0
	for _, entry := range bl.entries {
		if entry.Level == level {
			count++
		}
	}
	return count
}
