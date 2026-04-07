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
	"sync/atomic"
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

	// errors collects non-fatal errors encountered during request processing.
	// Middleware and handlers should append through Error() and read through
	// CollectedErrors() so mutation stays encapsulated in one canonical path.
	errors []error

	// Request state tracking
	startedAt          time.Time
	bodySize           atomic.Int64
	body               []byte
	bodyErr            error
	bodyReadOnce       sync.Once
	compressionEnabled atomic.Bool
	cancel             context.CancelFunc
	aborted            atomic.Bool

	// Middleware data storage
	mu    sync.RWMutex
	store map[string]any
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

	// ErrRequestTimeout is returned when the request processing times out.
	ErrRequestTimeout = errors.New("request timeout")

	// ErrInvalidJSON is returned when the request body contains invalid JSON.
	ErrInvalidJSON = errors.New("invalid JSON payload")

	// ErrEmptyRequestBody is returned when the request body is empty.
	ErrEmptyRequestBody = errors.New("request body is empty")

	// ErrUnexpectedExtraData is returned when the request body contains unexpected extra data.
	ErrUnexpectedExtraData = errors.New("unexpected extra data in request body")

	// ErrMissingParam is returned when a required route parameter is missing.
	ErrMissingParam = errors.New("missing parameter")

	// ErrMissingKey is returned when a required middleware store key is missing.
	ErrMissingKey = errors.New("missing key")

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

	// Detect compression
	compressionEnabled := false
	if cfg.EnableCompression {
		contentEncoding := strings.ToLower(r.Header.Get("Content-Encoding"))
		if contentEncoding == "gzip" || contentEncoding == "deflate" {
			compressionEnabled = true
		}
	}

	ctx := &Ctx{
		W:        w,
		R:        r,
		Params:   params,
		Query:    r.URL.Query(),
		ClientIP: clientIPFromRequest(r),
		Deadline: deadline,
		Config:   cfg,

		startedAt:          time.Now(),
		compressionEnabled: atomic.Bool{},
		cancel:             cancel,
	}

	ctx.compressionEnabled.Store(compressionEnabled)
	return ctx
}

// Close releases any request-scoped resources owned by the context.
func (c *Ctx) Close() {
	if c == nil || c.cancel == nil {
		return
	}
	c.cancel()
	c.cancel = nil
}

// Abort marks the context as aborted. Subsequent middleware or handlers
// should check IsAborted and skip processing when true. It also cancels
// the underlying request context so that long-running operations are
// notified via context.Done().
func (c *Ctx) Abort() {
	if c.aborted.CompareAndSwap(false, true) {
		if c.cancel != nil {
			c.cancel()
		}
	}
}

// AbortWithStatus writes the HTTP status code and marks the context as aborted atomically.
// If Abort has already been called, this is a no-op (WriteHeader is not called again).
func (c *Ctx) AbortWithStatus(code int) {
	if c.aborted.CompareAndSwap(false, true) {
		c.W.WriteHeader(code)
		if c.cancel != nil {
			c.cancel()
		}
	}
}

// IsAborted reports whether Abort has been called on this context.
func (c *Ctx) IsAborted() bool {
	return c.aborted.Load()
}

// Error appends a non-fatal error to the context's Errors slice and
// returns the same error for convenient inline use.
func (c *Ctx) Error(err error) error {
	if err != nil {
		c.mu.Lock()
		c.errors = append(c.errors, err)
		c.mu.Unlock()
	}
	return err
}

// CollectedErrors returns a snapshot of the non-fatal errors collected so far.
// Mutating the returned slice does not affect the context.
func (c *Ctx) CollectedErrors() []error {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.errors) == 0 {
		return nil
	}
	return append([]error(nil), c.errors...)
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

// Set stores a key-value pair in the context for sharing data between middleware and handlers.
// It is safe for concurrent use.
func (c *Ctx) Set(key string, value any) {
	c.mu.Lock()
	if c.store == nil {
		c.store = make(map[string]any)
	}
	c.store[key] = value
	c.mu.Unlock()
}

// Get retrieves a value from the context store. The boolean indicates whether the key was present.
// It is safe for concurrent use.
func (c *Ctx) Get(key string) (any, bool) {
	c.mu.RLock()
	val, ok := c.store[key]
	c.mu.RUnlock()
	return val, ok
}

// MustGet retrieves a value from the context store.
// It returns ErrMissingKey when the key has not been set, matching MustParam semantics.
// Use Get when a missing key is not an error condition.
func (c *Ctx) MustGet(key string) (any, error) {
	val, ok := c.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMissingKey, key)
	}
	return val, nil
}

// RequestDuration returns the time since the request started.
func (c *Ctx) RequestDuration() time.Duration {
	return time.Since(c.startedAt)
}

// IsCompressed returns whether the request body is compressed.
func (c *Ctx) IsCompressed() bool {
	return c.compressionEnabled.Load()
}

// BodySize returns the size of the request body.
func (c *Ctx) BodySize() int64 {
	return c.bodySize.Load()
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
		defer ctx.Close()
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
