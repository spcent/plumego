package glog

import (
	"context"
	"sync"
)

type contextKey int

const (
	traceIDKey contextKey = iota
	loggerKey
)

var (
	defaultLoggerOnce sync.Once
	defaultLogger     StructuredLogger
)

// WithTraceID adds a trace ID to the context.
// This trace ID will be automatically included in all Context-aware log methods.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext extracts the trace ID from context.
// Returns an empty string if no trace ID is present.
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithLogger attaches a logger instance to the context.
// This allows child functions to retrieve and use the same logger.
func WithLogger(ctx context.Context, logger StructuredLogger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext extracts the logger from context.
// Returns a default gLogger if no logger is present.
func LoggerFromContext(ctx context.Context) StructuredLogger {
	if ctx == nil {
		return getDefaultLogger()
	}
	if logger, ok := ctx.Value(loggerKey).(StructuredLogger); ok {
		return logger
	}
	return getDefaultLogger()
}

// LoggerFromContextOrNew extracts the logger from context, or creates a new one.
// If no trace ID is present in the context, it automatically generates one.
// Returns both the logger and an updated context with trace ID and logger attached.
//
// Note: this function has the side effect of attaching a new trace ID and logger to
// the returned context when they are absent. Always use the returned context for
// subsequent operations so those values are not lost.
func LoggerFromContextOrNew(ctx context.Context) (StructuredLogger, context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger, ok := ctx.Value(loggerKey).(StructuredLogger); ok {
		return logger, ctx
	}

	// Generate trace ID if not present
	if TraceIDFromContext(ctx) == "" {
		ctx = WithTraceID(ctx, NewTraceID())
	}

	logger := NewGLogger()
	ctx = WithLogger(ctx, logger)
	return logger, ctx
}

// NewRequestLogger creates a fresh request-scoped logger with a new trace ID, attaches
// both to the context, and returns them. Unlike LoggerFromContextOrNew, this always
// creates a new logger and trace ID regardless of what is already in the context.
// Use this at the start of each incoming request to establish a clean logging scope.
func NewRequestLogger(ctx context.Context) (StructuredLogger, context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	traceID := NewTraceID()
	ctx = WithTraceID(ctx, traceID)
	logger := NewGLogger()
	ctx = WithLogger(ctx, logger)
	return logger, ctx
}

func getDefaultLogger() StructuredLogger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = NewGLogger()
	})
	return defaultLogger
}
