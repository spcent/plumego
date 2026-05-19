package contract

import (
	"context"
	"errors"
	"sync"
)

// RequestContext contains router-owned request metadata shared across middleware
// and handlers. It preserves compatibility with the standard library by living
// inside the request's context.
type RequestContext struct {
	Params       map[string]string
	RoutePattern string
	RouteName    string
}

// maxInlineParams is the maximum number of route parameters stored inline
// in RouteState without a heap-allocated map.
const maxInlineParams = 8

// RouteState carries per-request routing data using inline fixed-size storage.
// It is stored in the request context by the router and read by the contract
// accessor functions. Callers must not retain the RouteState pointer beyond the
// handler's ServeHTTP call; use RequestContextFromContext to obtain an isolated
// copy for goroutines.
type RouteState struct {
	paramKeys [maxInlineParams]string
	paramVals [maxInlineParams]string
	paramN    int
	pattern   string
	name      string
}

// SetPattern records the matched route pattern.
func (rs *RouteState) SetPattern(p string) { rs.pattern = p }

// SetName records the route name.
func (rs *RouteState) SetName(n string) { rs.name = n }

// AddParam appends a key/value route parameter. Extra params beyond
// maxInlineParams are silently dropped (real routes rarely exceed 8 params).
func (rs *RouteState) AddParam(key, val string) {
	if rs.paramN < maxInlineParams {
		rs.paramKeys[rs.paramN] = key
		rs.paramVals[rs.paramN] = val
		rs.paramN++
	}
}

// Reset zeros the RouteState so it can be returned to a sync.Pool.
func (rs *RouteState) Reset() { *rs = RouteState{} }

// RouteStatePool is a package-level pool for RouteState objects. The router
// package uses this to avoid per-request allocations.
var RouteStatePool = sync.Pool{
	New: func() any { return new(RouteState) },
}

// Context keys are unexported zero-value structs to avoid collisions with other
// packages. External callers must use the With* and *FromContext accessor functions
// (e.g. WithRequestContext, RequestContextFromContext) rather than context.WithValue
// with the key types directly.
type requestContextKey struct{}
type routeStateKey struct{}

var (
	// ErrHandlerNil is returned when a handler is nil.
	ErrHandlerNil = errors.New("handler cannot be nil")

	// ErrResponseWriterNil is returned when a response writer is nil.
	ErrResponseWriterNil = errors.New("response writer cannot be nil")

	// ErrInvalidResponseStatus is returned when WriteResponse receives a
	// non-success HTTP status. Use WriteError for error statuses.
	ErrInvalidResponseStatus = errors.New("response status must be 2xx")
)

// WithRouteState stores rs in ctx under the internal routeStateKey. The router
// uses this as the fast path for attaching per-request route data without map
// allocations. Handlers must not retain the RouteState pointer beyond ServeHTTP.
func WithRouteState(ctx context.Context, rs *RouteState) context.Context {
	return context.WithValue(ctx, routeStateKey{}, rs)
}

// WithRequestContext stores rc in ctx using the package-internal requestContextKey.
// Callers should use this accessor instead of writing raw context keys.
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

	// Fast path: inline RouteState stored by the router.
	if rs, ok := ctx.Value(routeStateKey{}).(*RouteState); ok && rs != nil {
		rc := RequestContext{
			RoutePattern: rs.pattern,
			RouteName:    rs.name,
		}
		if rs.paramN > 0 {
			rc.Params = make(map[string]string, rs.paramN)
			for i := 0; i < rs.paramN; i++ {
				rc.Params[rs.paramKeys[i]] = rs.paramVals[i]
			}
		}
		return rc
	}

	// Fallback: legacy requestContextKey path used by direct callers of
	// WithRequestContext (e.g. middleware, tests).
	if rc, ok := ctx.Value(requestContextKey{}).(RequestContext); ok {
		rc.Params = cloneStringMap(rc.Params)
		return rc
	}

	return RequestContext{}
}

// RequestParamFromContext returns one route parameter from the stored
// RequestContext without cloning the full params map.
func RequestParamFromContext(ctx context.Context, name string) string {
	if ctx == nil || name == "" {
		return ""
	}

	// Fast path: inline RouteState — linear scan, zero allocations.
	if rs, ok := ctx.Value(routeStateKey{}).(*RouteState); ok && rs != nil {
		for i := 0; i < rs.paramN; i++ {
			if rs.paramKeys[i] == name {
				return rs.paramVals[i]
			}
		}
		return ""
	}

	// Fallback: legacy requestContextKey.
	if rc, ok := ctx.Value(requestContextKey{}).(RequestContext); ok {
		return rc.Params[name]
	}
	return ""
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
