package sharding

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// LogLevelDebug is for debug messages
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for informational messages
	LogLevelInfo
	// LogLevelWarn is for warning messages
	LogLevelWarn
	// LogLevelError is for error messages
	LogLevelError
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger is a structured logger for database sharding operations
type Logger struct {
	mu     sync.Mutex
	output io.Writer
	level  LogLevel
	fields map[string]interface{}
}

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	// Output is where logs are written (default: os.Stdout)
	Output io.Writer

	// Level is the minimum log level to output
	Level LogLevel

	// Fields are default fields added to every log message
	Fields map[string]interface{}
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Output: os.Stdout,
		Level:  LogLevelInfo,
		Fields: make(map[string]interface{}),
	}
}

// NewLogger creates a new logger
func NewLogger(config LoggerConfig) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}

	if config.Fields == nil {
		config.Fields = make(map[string]interface{})
	}

	return &Logger{
		output: config.Output,
		level:  config.Level,
		fields: config.Fields,
	}
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newFields := make(map[string]interface{}, len(l.fields)+len(fields))

	// Copy existing fields
	for k, v := range l.fields {
		newFields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		output: l.output,
		level:  l.level,
		fields: newFields,
	}
}

// WithField returns a new logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, msg string) {
	l.log(ctx, LogLevelDebug, msg, nil)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.log(ctx, LogLevelDebug, fmt.Sprintf(format, args...), nil)
}

// DebugWithFields logs a debug message with additional fields
func (l *Logger) DebugWithFields(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(ctx, LogLevelDebug, msg, fields)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, msg string) {
	l.log(ctx, LogLevelInfo, msg, nil)
}

// Infof logs a formatted info message
func (l *Logger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.log(ctx, LogLevelInfo, fmt.Sprintf(format, args...), nil)
}

// InfoWithFields logs an info message with additional fields
func (l *Logger) InfoWithFields(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(ctx, LogLevelInfo, msg, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, msg string) {
	l.log(ctx, LogLevelWarn, msg, nil)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.log(ctx, LogLevelWarn, fmt.Sprintf(format, args...), nil)
}

// WarnWithFields logs a warning message with additional fields
func (l *Logger) WarnWithFields(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(ctx, LogLevelWarn, msg, fields)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, msg string) {
	l.log(ctx, LogLevelError, msg, nil)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.log(ctx, LogLevelError, fmt.Sprintf(format, args...), nil)
}

// ErrorWithFields logs an error message with additional fields
func (l *Logger) ErrorWithFields(ctx context.Context, msg string, fields map[string]interface{}) {
	l.log(ctx, LogLevelError, msg, fields)
}

// log is the internal logging function
func (l *Logger) log(ctx context.Context, level LogLevel, msg string, fields map[string]interface{}) {
	// Check log level
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Build log entry
	entry := l.buildLogEntry(level, msg, fields)

	// Write to output
	fmt.Fprintln(l.output, entry)
}

// buildLogEntry builds a structured log entry
func (l *Logger) buildLogEntry(level LogLevel, msg string, fields map[string]interface{}) string {
	// Simple JSON-like format
	// In production, use actual JSON encoding
	entry := fmt.Sprintf(`{"time":"%s","level":"%s","msg":"%s"`,
		time.Now().UTC().Format(time.RFC3339),
		level.String(),
		msg,
	)

	// Add default fields
	for k, v := range l.fields {
		entry += fmt.Sprintf(`,"%s":"%v"`, k, v)
	}

	// Add message-specific fields
	for k, v := range fields {
		entry += fmt.Sprintf(`,"%s":"%v"`, k, v)
	}

	entry += "}"

	return entry
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// LoggingRouter wraps a router with logging
type LoggingRouter struct {
	router *Router
	logger *Logger
}

// NewLoggingRouter creates a new logging router
func NewLoggingRouter(router *Router, logger *Logger) *LoggingRouter {
	if logger == nil {
		logger = NewLogger(DefaultLoggerConfig())
	}

	return &LoggingRouter{
		router: router,
		logger: logger.WithFields(map[string]interface{}{
			"component": "sharding",
		}),
	}
}

// Logger returns the logger
func (lr *LoggingRouter) Logger() *Logger {
	return lr.logger
}

// Router returns the underlying router
func (lr *LoggingRouter) Router() *Router {
	return lr.router
}

// LogQuery logs a query execution
func (lr *LoggingRouter) LogQuery(ctx context.Context, query string, shardIndex int, latency time.Duration, err error) {
	fields := map[string]interface{}{
		"query":       query,
		"shard_index": shardIndex,
		"latency_ms":  latency.Milliseconds(),
	}

	if err != nil {
		fields["error"] = err.Error()
		lr.logger.ErrorWithFields(ctx, "query failed", fields)
	} else {
		lr.logger.DebugWithFields(ctx, "query executed", fields)
	}
}

// LogShardResolution logs shard resolution
func (lr *LoggingRouter) LogShardResolution(ctx context.Context, tableName string, shardKey interface{}, shardIndex int) {
	lr.logger.DebugWithFields(ctx, "shard resolved", map[string]interface{}{
		"table":       tableName,
		"shard_key":   fmt.Sprintf("%v", shardKey),
		"shard_index": shardIndex,
	})
}

// LogCrossShardQuery logs a cross-shard query
func (lr *LoggingRouter) LogCrossShardQuery(ctx context.Context, query string, policy string) {
	lr.logger.WarnWithFields(ctx, "cross-shard query", map[string]interface{}{
		"query":  query,
		"policy": policy,
	})
}

// LogRewrite logs SQL rewriting
func (lr *LoggingRouter) LogRewrite(ctx context.Context, original string, rewritten string, cached bool) {
	lr.logger.DebugWithFields(ctx, "SQL rewritten", map[string]interface{}{
		"original":  original,
		"rewritten": rewritten,
		"cached":    cached,
	})
}
