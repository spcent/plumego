package contract

import "context"

type TraceIDKey struct{}

// TraceIDFromContext extracts the trace id injected by the Logging middleware.
func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(TraceIDKey{}).(string); ok {
		return v
	}
	return ""
}
