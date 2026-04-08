package log

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Compile-time checks that all concrete logger types satisfy StructuredLogger.
var (
	_ StructuredLogger = (*defaultLogger)(nil)
	_ StructuredLogger = (*jsonLogger)(nil)
	_ StructuredLogger = (*discardLogger)(nil)
)

// Fields represents structured log fields to attach to a log entry.
type Fields map[string]any

// LoggerFormat selects the concrete logger implementation behind NewLogger.
type LoggerFormat string

const (
	LoggerFormatText    LoggerFormat = "text"
	LoggerFormatJSON    LoggerFormat = "json"
	LoggerFormatDiscard LoggerFormat = "discard"
)

// LoggerConfig configures the canonical NewLogger constructor.
type LoggerConfig struct {
	Format           LoggerFormat
	Output           io.Writer
	ErrorOutput      io.Writer
	Level            Level
	Fields           Fields
	RespectVerbosity bool
	Verbosity        int
}

// StructuredLogger defines the minimal logging interface used by the application.
//
// All logging methods accept an optional Fields argument. Callers may omit
// it entirely when there are no extra fields to attach:
//
//	logger.Info("server started")
//	logger.Info("server started", glog.Fields{"addr": ":8080"})
//
// Implementations must treat a nil or absent Fields the same as an empty one.
type StructuredLogger interface {
	// WithFields returns a new logger with the given fields merged into every
	// subsequent log entry. The original logger is not modified.
	WithFields(fields Fields) StructuredLogger

	// With is a convenience shortcut for WithFields(Fields{key: value}).
	// Prefer WithFields when attaching multiple fields at once.
	With(key string, value any) StructuredLogger

	Debug(msg string, fields ...Fields)
	Info(msg string, fields ...Fields)
	Warn(msg string, fields ...Fields)
	Error(msg string, fields ...Fields)
	// Fatal logs a message at FATAL level then calls os.Exit(1).
	Fatal(msg string, fields ...Fields)

	// Context-aware variants preserve the call shape for request-scoped logging.
	// Logger implementations must not infer transport metadata from ctx.
	DebugCtx(ctx context.Context, msg string, fields ...Fields)
	InfoCtx(ctx context.Context, msg string, fields ...Fields)
	WarnCtx(ctx context.Context, msg string, fields ...Fields)
	ErrorCtx(ctx context.Context, msg string, fields ...Fields)
	FatalCtx(ctx context.Context, msg string, fields ...Fields)
}

// Lifecycle allows a logger to participate in application start/stop hooks
// (e.g. to initialize flags or flush buffers). Methods are optional.
type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// firstFields returns the first Fields argument if present, otherwise nil.
func firstFields(extra []Fields) Fields {
	if len(extra) > 0 {
		return extra[0]
	}
	return nil
}

// defaultLogger adapts the default text logger backend to StructuredLogger.
type defaultLogger struct {
	fields Fields
}

// NewLogger creates the canonical structured logger.
// With no config it returns the default text logger. Alternate formats such as
// JSON or discard mode are selected through LoggerConfig.Format so there is
// one constructor path for all stable logger variants.
func NewLogger(configs ...LoggerConfig) StructuredLogger {
	cfg := LoggerConfig{}
	if len(configs) > 0 {
		cfg = configs[0]
	}

	switch cfg.Format {
	case LoggerFormatJSON:
		return newJSONLogger(cfg)
	case LoggerFormatDiscard:
		return newDiscardLogger()
	case "", LoggerFormatText:
		return &defaultLogger{fields: cloneFields(cfg.Fields)}
	default:
		return &defaultLogger{fields: cloneFields(cfg.Fields)}
	}
}

// Start initializes the underlying default logger backend.
func (l *defaultLogger) Start(ctx context.Context) error {
	initDefaultFromFlags()
	return nil
}

// Stop flushes and closes backend resources.
func (l *defaultLogger) Stop(ctx context.Context) error {
	flushDefault()
	closeDefault()
	return nil
}

func (l *defaultLogger) WithFields(fields Fields) StructuredLogger {
	return &defaultLogger{fields: mergeFields(l.fields, fields)}
}

func (l *defaultLogger) With(key string, value any) StructuredLogger {
	return l.WithFields(Fields{key: value})
}

// Debug logs at DEBUG level.
// The canonical logger path gates debug on V(1).
func (l *defaultLogger) Debug(msg string, fields ...Fields) {
	if !std.vAt(1, 2) {
		return
	}
	l.logWithLevel(DEBUG, msg, firstFields(fields))
}

func (l *defaultLogger) Info(msg string, fields ...Fields) {
	l.logWithLevel(INFO, msg, firstFields(fields))
}

func (l *defaultLogger) Warn(msg string, fields ...Fields) {
	l.logWithLevel(WARNING, msg, firstFields(fields))
}

func (l *defaultLogger) Error(msg string, fields ...Fields) {
	l.logWithLevel(ERROR, msg, firstFields(fields))
}

func (l *defaultLogger) Fatal(msg string, fields ...Fields) {
	l.logWithLevel(FATAL, msg, firstFields(fields))
}

// DebugCtx logs at DEBUG level with context.
// Verbosity is gated on the global glog flag, consistent with Debug.
func (l *defaultLogger) DebugCtx(ctx context.Context, msg string, fields ...Fields) {
	_ = ctx
	if !std.vAt(1, 2) {
		return
	}
	l.logWithLevel(DEBUG, msg, firstFields(fields))
}

func (l *defaultLogger) InfoCtx(ctx context.Context, msg string, fields ...Fields) {
	_ = ctx
	l.logWithLevel(INFO, msg, firstFields(fields))
}

func (l *defaultLogger) WarnCtx(ctx context.Context, msg string, fields ...Fields) {
	_ = ctx
	l.logWithLevel(WARNING, msg, firstFields(fields))
}

func (l *defaultLogger) ErrorCtx(ctx context.Context, msg string, fields ...Fields) {
	_ = ctx
	l.logWithLevel(ERROR, msg, firstFields(fields))
}

func (l *defaultLogger) FatalCtx(ctx context.Context, msg string, fields ...Fields) {
	_ = ctx
	l.logWithLevel(FATAL, msg, firstFields(fields))
}

func (l *defaultLogger) logWithLevel(level Level, msg string, fields Fields) {
	combined := mergeFields(l.fields, fields)
	formatted := l.formatFields(combined)
	if formatted != "" {
		msg += " " + formatted
	}

	// calldepth=3: std.log adds 1 → logInternal calls runtime.Caller(4)
	// Frame 0: logInternal, 1: std.log, 2: logWithLevel, 3: public method (Info/Error/…), 4: actual caller
	std.log(level, 3, msg)
}

func (l *defaultLogger) formatFields(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, fields[k]))
	}

	return strings.Join(parts, " ")
}
