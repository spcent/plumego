package log

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spcent/plumego/contract"
)

// Compile-time checks that all concrete logger types satisfy StructuredLogger.
var (
	_ StructuredLogger = (*gLogger)(nil)
	_ StructuredLogger = (*JSONLogger)(nil)
	_ StructuredLogger = (*TestLogger)(nil)
	_ StructuredLogger = (*NoOpLogger)(nil)
)

// Fields represents structured log fields to attach to a log entry.
type Fields map[string]any

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

	// Context-aware variants extract trace IDs and other values from ctx.
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

func (l *gLogger) With(key string, value any) StructuredLogger {
	return l.WithFields(Fields{key: value})
}

// Debug logs at DEBUG level.
// Unlike JSONLogger (which checks its own local verbosity), gLogger gates Debug
// on the global glog verbosity flag (-v). Both require at least V(1) to emit.
func (l *gLogger) Debug(msg string, fields ...Fields) {
	if !std.vAt(1, 3) {
		return
	}
	l.logWithLevel(DEBUG, msg, firstFields(fields), nil)
}

func (l *gLogger) Info(msg string, fields ...Fields) {
	l.logWithLevel(INFO, msg, firstFields(fields), nil)
}

func (l *gLogger) Warn(msg string, fields ...Fields) {
	l.logWithLevel(WARNING, msg, firstFields(fields), nil)
}

func (l *gLogger) Error(msg string, fields ...Fields) {
	l.logWithLevel(ERROR, msg, firstFields(fields), nil)
}

func (l *gLogger) Fatal(msg string, fields ...Fields) {
	l.logWithLevel(FATAL, msg, firstFields(fields), nil)
}

// DebugCtx logs at DEBUG level with context.
// Verbosity is gated on the global glog flag, consistent with Debug.
func (l *gLogger) DebugCtx(ctx context.Context, msg string, fields ...Fields) {
	if !std.vAt(1, 3) {
		return
	}
	l.logWithLevel(DEBUG, msg, firstFields(fields), ctx)
}

func (l *gLogger) InfoCtx(ctx context.Context, msg string, fields ...Fields) {
	l.logWithLevel(INFO, msg, firstFields(fields), ctx)
}

func (l *gLogger) WarnCtx(ctx context.Context, msg string, fields ...Fields) {
	l.logWithLevel(WARNING, msg, firstFields(fields), ctx)
}

func (l *gLogger) ErrorCtx(ctx context.Context, msg string, fields ...Fields) {
	l.logWithLevel(ERROR, msg, firstFields(fields), ctx)
}

func (l *gLogger) FatalCtx(ctx context.Context, msg string, fields ...Fields) {
	l.logWithLevel(FATAL, msg, firstFields(fields), ctx)
}

func (l *gLogger) logWithLevel(level Level, msg string, fields Fields, ctx context.Context) {
	combined := mergeFields(l.fields, fields)
	if traceID := contract.TraceIDFromContext(ctx); traceID != "" {
		combined["trace_id"] = traceID
	}
	formatted := l.formatFields(combined)
	if formatted != "" {
		msg += " " + formatted
	}

	// calldepth=3: std.log adds 1 → logInternal calls runtime.Caller(4)
	// Frame 0: logInternal, 1: std.log, 2: logWithLevel, 3: public method (Info/Error/…), 4: actual caller
	std.log(level, 3, msg)
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
