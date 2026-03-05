package glog

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Fields represents structured log fields to attach to a log entry.
type Fields map[string]any

// StructuredLogger defines the minimal logging interface used by the application.
//
// Implementations can wrap zap/logrus/zerolog/slog or any other logger.
// The logger should treat a nil Fields map the same as an empty one.
type StructuredLogger interface {
	WithFields(fields Fields) StructuredLogger

	Debug(msg string, fields Fields)
	Info(msg string, fields Fields)
	Warn(msg string, fields Fields)
	Error(msg string, fields Fields)
	// Fatal logs a message at FATAL level then calls os.Exit(1).
	Fatal(msg string, fields Fields)

	// Context-aware logging methods that can extract trace IDs and other context values
	DebugCtx(ctx context.Context, msg string, fields Fields)
	InfoCtx(ctx context.Context, msg string, fields Fields)
	WarnCtx(ctx context.Context, msg string, fields Fields)
	ErrorCtx(ctx context.Context, msg string, fields Fields)
	FatalCtx(ctx context.Context, msg string, fields Fields)
}

// Lifecycle allows a logger to participate in application start/stop hooks
// (e.g. to initialize flags or flush buffers). Methods are optional.
type Lifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// gLogger adapts the existing glog implementation to the StructuredLogger interface.
type gLogger struct {
	fields Fields
}

// NewGLogger creates a StructuredLogger backed by the package-level glog implementation.
// It implements Lifecycle to run glog.Init/Flush/Close automatically.
func NewGLogger() StructuredLogger {
	return &gLogger{fields: Fields{}}
}

// Start initializes the underlying glog system.
func (l *gLogger) Start(ctx context.Context) error {
	Init()
	return nil
}

// Stop flushes and closes glog resources.
func (l *gLogger) Stop(ctx context.Context) error {
	Flush()
	Close()
	return nil
}

func (l *gLogger) WithFields(fields Fields) StructuredLogger {
	return &gLogger{fields: mergeFields(l.fields, fields)}
}

func (l *gLogger) Debug(msg string, fields Fields) {
	if !std.vAt(1, 3) {
		return
	}
	l.logWithLevel(DEBUG, msg, fields, nil)
}

func (l *gLogger) Info(msg string, fields Fields) {
	l.logWithLevel(INFO, msg, fields, nil)
}

func (l *gLogger) Warn(msg string, fields Fields) {
	l.logWithLevel(WARNING, msg, fields, nil)
}

func (l *gLogger) Error(msg string, fields Fields) {
	l.logWithLevel(ERROR, msg, fields, nil)
}

func (l *gLogger) Fatal(msg string, fields Fields) {
	l.logWithLevel(FATAL, msg, fields, nil)
}

// DebugCtx logs a debug message with context support.
func (l *gLogger) DebugCtx(ctx context.Context, msg string, fields Fields) {
	if !std.vAt(1, 3) {
		return
	}
	l.logWithLevel(DEBUG, msg, fields, ctx)
}

// InfoCtx logs an info message with context support.
func (l *gLogger) InfoCtx(ctx context.Context, msg string, fields Fields) {
	l.logWithLevel(INFO, msg, fields, ctx)
}

// WarnCtx logs a warning message with context support.
func (l *gLogger) WarnCtx(ctx context.Context, msg string, fields Fields) {
	l.logWithLevel(WARNING, msg, fields, ctx)
}

// ErrorCtx logs an error message with context support.
func (l *gLogger) ErrorCtx(ctx context.Context, msg string, fields Fields) {
	l.logWithLevel(ERROR, msg, fields, ctx)
}

// FatalCtx logs a fatal message with context support then calls os.Exit(1).
func (l *gLogger) FatalCtx(ctx context.Context, msg string, fields Fields) {
	l.logWithLevel(FATAL, msg, fields, ctx)
}

func (l *gLogger) logWithLevel(level Level, msg string, fields Fields, ctx context.Context) {
	combined := mergeFields(l.fields, fields)
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		combined["trace_id"] = traceID
	}
	formatted := l.formatFields(combined)
	if formatted != "" {
		msg += " " + formatted
	}

	switch level {
	case DEBUG:
		// calldepth=3: logWithLevel → Debug/DebugCtx → caller
		std.log(DEBUG, 3, msg)
	case INFO:
		Info(msg)
	case WARNING:
		Warning(msg)
	case ERROR:
		Error(msg)
	case FATAL:
		Fatal(msg)
	}
}

func (l *gLogger) formatFields(fields Fields) string {
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
