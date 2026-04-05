package contract

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/log"
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
	Headers  http.Header
	ClientIP string
	Logger   log.StructuredLogger
	TraceID  string
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

// BindError represents an error that occurred while binding a request body.
type BindError struct {
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

	// ErrMissingParam is returned when a required parameter is missing.
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
)

// Error implements the error interface.
func (e *BindError) Error() string {
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
func (e *BindError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type paramsContextKey struct{}

// WithRequestContext stores rc in ctx using the package-internal requestContextKey.
// Use this instead of context.WithValue with the old exported key.
func WithRequestContext(ctx context.Context, rc RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey{}, rc)
}

// WithParams stores params in ctx using the package-internal paramsContextKey.
// Use this instead of context.WithValue with the old exported key.
func WithParams(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, paramsContextKey{}, params)
}

// ParamsFromContext returns route parameters stored in the request context.
// It returns nil if no parameters were attached.
func ParamsFromContext(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	if params, ok := ctx.Value(paramsContextKey{}).(map[string]string); ok {
		return params
	}

	if rc, ok := ctx.Value(requestContextKey{}).(RequestContext); ok {
		return rc.Params
	}

	return nil
}

// RequestContextFromContext returns the RequestContext stored in the given context.
// If none is present, it falls back to parameters stored via ParamsFromContext for backward compatibility.
func RequestContextFromContext(ctx context.Context) RequestContext {
	if ctx == nil {
		return RequestContext{}
	}

	if rc, ok := ctx.Value(requestContextKey{}).(RequestContext); ok {
		return rc
	}

	return RequestContext{Params: ParamsFromContext(ctx)}
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
	return newCtxWithLogger(w, r, params, nil)
}

// NewCtxWithConfig builds a unified request context with a custom RequestConfig,
// allowing callers to override the default without mutating global state.
func NewCtxWithConfig(w http.ResponseWriter, r *http.Request, params map[string]string, cfg RequestConfig) *Ctx {
	return newCtxWithLoggerAndConfig(w, r, params, nil, &cfg)
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

func defaultRequestConfig() *RequestConfig {
	cfg := defaultConfig
	return &cfg
}

// newCtxWithLogger allows injecting a logger while keeping NewCtx minimal for compatibility.
func newCtxWithLogger(w http.ResponseWriter, r *http.Request, params map[string]string, logger log.StructuredLogger) *Ctx {
	return newCtxWithLoggerAndConfig(w, r, params, logger, nil)
}

// newCtxWithLoggerAndConfig is the canonical constructor; nil cfg uses the package default.
func newCtxWithLoggerAndConfig(w http.ResponseWriter, r *http.Request, params map[string]string, logger log.StructuredLogger, cfg *RequestConfig) *Ctx {
	if params == nil {
		params = map[string]string{}
	}

	if cfg == nil {
		cfg = defaultRequestConfig()
	}

	var cancel context.CancelFunc
	deadline, hasDeadline := r.Context().Deadline()
	if !hasDeadline && cfg.RequestTimeout > 0 {
		timeoutCtx, cancelFunc := context.WithTimeout(r.Context(), cfg.RequestTimeout)
		cancel = cancelFunc
		r = r.WithContext(timeoutCtx)
		deadline, _ = timeoutCtx.Deadline()
	}

	traceID := TraceIDFromContext(r.Context())

	if logger == nil {
		logger = log.NewNoOpLogger()
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
		Headers:  r.Header,
		ClientIP: clientIPFromRequest(r),
		Logger:   logger,
		TraceID:  traceID,
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

// AbortWithStatus is a convenience that writes the HTTP status code and
// then marks the context as aborted.
func (c *Ctx) AbortWithStatus(code int) {
	c.W.WriteHeader(code)
	c.Abort()
}

// IsAborted reports whether Abort has been called on this context.
func (c *Ctx) IsAborted() bool {
	return c.aborted.Load()
}

// Error appends a non-fatal error to the context's Errors slice and
// returns the same error for convenient inline use.
func (c *Ctx) Error(err error) error {
	if err != nil {
		c.errors = append(c.errors, err)
	}
	return err
}

// CollectedErrors returns a snapshot of the non-fatal errors collected so far.
// Mutating the returned slice does not affect the context.
func (c *Ctx) CollectedErrors() []error {
	if c == nil || len(c.errors) == 0 {
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
		return "", errors.New("missing param: " + key)
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

// MustGet retrieves a value from the context store and panics if the key does not exist.
// Use this only when the key is guaranteed to have been set by prior middleware.
func (c *Ctx) MustGet(key string) any {
	val, ok := c.Get(key)
	if !ok {
		panic("contract.Ctx: missing key " + key)
	}
	return val
}

// GetRequestDuration returns the time since the request started.
func (c *Ctx) GetRequestDuration() time.Duration {
	return time.Since(c.startedAt)
}

// IsCompressed returns whether the request body is compressed.
func (c *Ctx) IsCompressed() bool {
	return c.compressionEnabled.Load()
}

// GetBodySize returns the size of the request body.
func (c *Ctx) GetBodySize() int64 {
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
func AdaptCtxHandler(h CtxHandlerFunc, logger log.StructuredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := RequestContextFromContext(r.Context())
		ctx := newCtxWithLogger(w, r, rc.Params, logger)
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
