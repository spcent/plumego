package contract

import (
	"context"
	"strings"
)

type requestIDContextKey struct{}

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
	if id == "" || containsControlChar(id) {
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

func containsControlChar(id string) bool {
	for _, ch := range id {
		if ch < 0x20 || ch == 0x7f {
			return true
		}
	}
	return false
}
