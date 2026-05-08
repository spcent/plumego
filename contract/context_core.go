package contract

import (
	"context"
	"errors"
)

// RequestContext contains router-owned request metadata shared across middleware
// and handlers. It preserves compatibility with the standard library by living
// inside the request's context.
type RequestContext struct {
	Params       map[string]string
	RoutePattern string
	RouteName    string
}

// Context keys are unexported zero-value structs to avoid collisions with other
// packages. External callers must use the With* and *FromContext accessor functions
// (e.g. WithRequestContext, RequestContextFromContext) rather than context.WithValue
// with the key types directly.
type requestContextKey struct{}

var (
	// ErrHandlerNil is returned when a handler is nil.
	ErrHandlerNil = errors.New("handler cannot be nil")

	// ErrResponseWriterNil is returned when a response writer is nil.
	ErrResponseWriterNil = errors.New("response writer cannot be nil")
)

// WithRequestContext stores rc in ctx using the package-internal requestContextKey.
// Use this instead of context.WithValue with the old exported key.
func WithRequestContext(ctx context.Context, rc RequestContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	rc.Params = cloneStringMap(rc.Params)
	return context.WithValue(ctx, requestContextKey{}, rc)
}

// RequestContextFromContext returns the RequestContext stored in the given context.
func RequestContextFromContext(ctx context.Context) RequestContext {
	if ctx == nil {
		return RequestContext{}
	}

	if rc, ok := ctx.Value(requestContextKey{}).(RequestContext); ok {
		rc.Params = cloneStringMap(rc.Params)
		return rc
	}

	return RequestContext{}
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
