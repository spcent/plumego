package glog

import "context"

type contextKey int

const (
	traceIDKey contextKey = iota
	loggerKey
)

// WithTraceID adds a trace ID to the context.
// This trace ID will be automatically included in all Context-aware log methods.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext extracts the trace ID from context.
// Returns an empty string if no trace ID is present.
func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithLogger attaches a logger instance to the context.
// This allows child functions to retrieve and use the same logger.
func WithLogger(ctx context.Context, logger StructuredLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext extracts the logger from context.
// Returns a default gLogger if no logger is present.
func LoggerFromContext(ctx context.Context) StructuredLogger {
	if logger, ok := ctx.Value(loggerKey).(StructuredLogger); ok {
		return logger
	}
	return NewGLogger()
}

// LoggerFromContextOrNew extracts the logger from context, or creates a new one.
// If no trace ID is present in the context, it automatically generates one.
// Returns both the logger and an updated context with trace ID and logger attached.
func LoggerFromContextOrNew(ctx context.Context) (StructuredLogger, context.Context) {
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
