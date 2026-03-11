package log

import (
	"context"
)

type contextKey int

const (
	traceIDKey contextKey = iota
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
