package contract

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// RequestContext contains request-scoped data that should be shared across middleware and handlers.
// It preserves compatibility with the standard library by living inside the request's context.
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

// RequestConfig holds configuration for request processing.
type RequestConfig struct {
	MaxBodySize       int64
	EnableBodyCache   bool
	EnableCompression bool
	RequestTimeout    time.Duration
}

// Ctx is a unified context object shared by handlers.
// It exposes common request-scoped attributes and helper methods for writing responses.
type Ctx struct {
	W        http.ResponseWriter
	R        *http.Request
	Params   map[string]string
	Query    url.Values
	ClientIP string
	Deadline time.Time
	Config   *RequestConfig

	body         []byte
	bodyErr      error
	bodyReadOnce sync.Once
	cancel       context.CancelFunc
}

// RequestHeaders returns the request's header map.
// This is a direct alias for R.Header: writes to the returned map modify the
// underlying http.Request headers and are visible to all subsequent middleware.
func (c *Ctx) RequestHeaders() http.Header {
	if c == nil || c.R == nil {
		return nil
	}
	return c.R.Header
}

// RequestID returns the request correlation id associated with this request.
func (c *Ctx) RequestID() string {
	if c == nil || c.R == nil {
		return ""
	}
	return RequestIDFromContext(c.R.Context())
}

// bindError represents an error that occurred while binding a request body.
type bindError struct {
	Status  int
	Message string
	Err     error
}

var (
	// ErrRequestBodyTooLarge is returned when the request body exceeds the maximum allowed size.
	ErrRequestBodyTooLarge = errors.New("request body too large")

	// ErrInvalidJSON is returned when the request body contains invalid JSON.
	ErrInvalidJSON = errors.New("invalid JSON payload")

	// ErrEmptyRequestBody is returned when the request body is empty.
	ErrEmptyRequestBody = errors.New("request body is empty")

	// ErrUnexpectedExtraData is returned when the request body contains unexpected extra data.
	ErrUnexpectedExtraData = errors.New("unexpected extra data in request body")

	// ErrMissingParam is returned when a required route parameter is missing.
	ErrMissingParam = errors.New("missing parameter")

	// ErrInvalidParam is returned when a parameter has an invalid value.
	ErrInvalidParam = errors.New("invalid parameter value")

	// ErrValidationFailed is returned when request validation fails.
	ErrValidationFailed = errors.New("validation failed")

	// ErrCompressionNotSupported is returned when compression is not supported.
	ErrCompressionNotSupported = errors.New("compression not supported")

	// ErrHandlerNil is returned when a handler is nil.
	ErrHandlerNil = errors.New("handler cannot be nil")

	// ErrInvalidChunkSize is returned when a streaming chunk size is invalid.
	ErrInvalidChunkSize = errors.New("invalid chunk size")

	// ErrContextNil is returned when a context is nil.
	ErrContextNil = errors.New("context cannot be nil")

	// ErrRequestNil is returned when a request is nil.
	ErrRequestNil = errors.New("request cannot be nil")

	// ErrResponseWriterNil is returned when a response writer is nil.
	ErrResponseWriterNil = errors.New("response writer cannot be nil")

	// ErrConfigNil is returned when a config is nil.
	ErrConfigNil = errors.New("config cannot be nil")

	// ErrInvalidBindDst is returned when a bind destination is nil or otherwise invalid.
	ErrInvalidBindDst = errors.New("invalid bind destination")

	// ErrInvalidQueryParam is returned when a query parameter has an invalid value.
	ErrInvalidQueryParam = errors.New("invalid query parameter")
)

// Error implements the error interface.
func (e *bindError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "binding error"
}

// Unwrap exposes the underlying error.
func (e *bindError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// WithRequestContext stores rc in ctx using the package-internal requestContextKey.
// Use this instead of context.WithValue with the old exported key.
func WithRequestContext(ctx context.Context, rc RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey{}, rc)
}

// RequestContextFromContext returns the RequestContext stored in the given context.
func RequestContextFromContext(ctx context.Context) RequestContext {
	if ctx == nil {
		return RequestContext{}
	}

	if rc, ok := ctx.Value(requestContextKey{}).(RequestContext); ok {
		return rc
	}

	return RequestContext{}
}

// RoutePatternFromContext returns the matched route pattern stored in the request context.
func RoutePatternFromContext(ctx context.Context) string {
	return RequestContextFromContext(ctx).RoutePattern
}

// RouteNameFromContext returns the matched route name stored in the request context.
func RouteNameFromContext(ctx context.Context) string {
	return RequestContextFromContext(ctx).RouteName
}

// NewCtx builds a unified request context for handlers using the net/http primitives.
func NewCtx(w http.ResponseWriter, r *http.Request, params map[string]string) *Ctx {
	return newCtxWithConfig(w, r, params, nil)
}

// NewCtxWithConfig builds a unified request context with a custom RequestConfig,
// allowing callers to override the default without mutating global state.
func NewCtxWithConfig(w http.ResponseWriter, r *http.Request, params map[string]string, cfg RequestConfig) *Ctx {
	return newCtxWithConfig(w, r, params, &cfg)
}

var defaultConfig = RequestConfig{
	MaxBodySize:       10 * 1024 * 1024, // 10MB
	EnableBodyCache:   true,
	EnableCompression: false,
	RequestTimeout:    30 * time.Second,
}

// DefaultConfig returns a copy of the default request processing configuration.
// Modify the returned value freely; it does not affect the package default.
func DefaultConfig() RequestConfig {
	return defaultConfig
}

// newCtxWithConfig is the canonical constructor; nil cfg uses the package default.
func newCtxWithConfig(w http.ResponseWriter, r *http.Request, params map[string]string, cfg *RequestConfig) *Ctx {
	if params == nil {
		params = map[string]string{}
	}

	if cfg == nil {
		c := defaultConfig
		cfg = &c
	}

	var cancel context.CancelFunc
	deadline, hasDeadline := r.Context().Deadline()
	if !hasDeadline && cfg.RequestTimeout > 0 {
		timeoutCtx, cancelFunc := context.WithTimeout(r.Context(), cfg.RequestTimeout)
		cancel = cancelFunc
		r = r.WithContext(timeoutCtx)
		deadline, _ = timeoutCtx.Deadline()
	}

	ctx := &Ctx{
		W:        w,
		R:        r,
		Params:   params,
		Query:    r.URL.Query(),
		ClientIP: clientIPFromRequest(r),
		Deadline: deadline,
		Config:   cfg,
		cancel:   cancel,
	}
	return ctx
}

// release cancels request-scoped resources owned by the context.
func (c *Ctx) release() {
	if c == nil || c.cancel == nil {
		return
	}
	c.cancel()
	c.cancel = nil
}

func (c *Ctx) Param(key string) (string, bool) {
	if c.Params == nil {
		return "", false
	}
	val, ok := c.Params[key]
	return val, ok
}

func (c *Ctx) MustParam(key string) (string, error) {
	val, ok := c.Param(key)
	if !ok || val == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingParam, key)
	}
	return val, nil
}

func clientIPFromRequest(r *http.Request) string {
	// X-Forwarded-For is only trustworthy when appended by infrastructure you control.
	// Use the last non-empty value rather than the first client-supplied value.
	if r == nil {
		return ""
	}
	parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
	for i := len(parts) - 1; i >= 0; i-- {
		if ip := strings.TrimSpace(parts[i]); ip != "" {
			return ip
		}
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// CtxHandlerFunc is the handler signature that receives a unified Ctx.
type CtxHandlerFunc func(*Ctx)

// AdaptCtxHandler converts a CtxHandlerFunc to a standard http.Handler to keep net/http compatibility.
func AdaptCtxHandler(h CtxHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := RequestContextFromContext(r.Context())
		ctx := newCtxWithConfig(w, r, rc.Params, nil)
		defer ctx.release()
		h(ctx)
	})
}

// ValidateCtxHandler returns an error when the handler is nil to give clearer feedback to callers.
func ValidateCtxHandler(h CtxHandlerFunc) error {
	if h == nil {
		return ErrHandlerNil
	}
	return nil
}
