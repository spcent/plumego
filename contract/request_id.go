package contract

import (
	"context"
	"strings"
)

type requestIDContextKey struct{}

// WithRequestID stores the canonical request correlation id in ctx.
func WithRequestID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDContextKey{}, id)
}

// RequestIDFromContext returns the canonical request correlation id from ctx.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDContextKey{}).(string); ok {
		return strings.TrimSpace(id)
	}
	return ""
}
