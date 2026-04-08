package contract

import (
	"context"
	"strings"
)

type requestIDKey struct{}

const (
	// RequestIDHeader is the canonical request id header.
	RequestIDHeader = "X-Request-ID"
)

// WithRequestID stores the canonical request correlation id in ctx.
func WithRequestID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the canonical request correlation id from ctx.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return strings.TrimSpace(id)
	}
	return ""
}
