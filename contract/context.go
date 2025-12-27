package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/spcent/plumego/log"
)

// RequestContext contains request-scoped data that should be shared across middleware and handlers.
// It preserves compatibility with the standard library by living inside the request's context.
type RequestContext struct {
	Params map[string]string
}

type RequestContextKey struct{}

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

	bodyOnce sync.Once
	body     []byte
	bodyErr  error
}

// BindError represents an error that occurred while binding a request body.
type BindError struct {
	Status  int
	Message string
	Err     error
}

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

type ParamsContextKey struct{}

// ParamsFromContext returns route parameters stored in the request context.
// It returns nil if no parameters were attached.
func ParamsFromContext(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	if params, ok := ctx.Value(ParamsContextKey{}).(map[string]string); ok {
		return params
	}

	if rc, ok := ctx.Value(RequestContextKey{}).(RequestContext); ok {
		return rc.Params
	}

	return nil
}

// RequestContextFrom returns the RequestContext stored in the given context.
// If none is present, it falls back to parameters stored via ParamsFromContext for backward compatibility.
func RequestContextFrom(ctx context.Context) RequestContext {
	if ctx == nil {
		return RequestContext{}
	}

	if rc, ok := ctx.Value(RequestContextKey{}).(RequestContext); ok {
		return rc
	}

	return RequestContext{Params: ParamsFromContext(ctx)}
}

// Param returns a single path parameter from the request's context.
// The boolean indicates whether the parameter was present.
func Param(r *http.Request, key string) (string, bool) {
	rc := RequestContextFrom(r.Context())
	if rc.Params == nil {
		return "", false
	}
	val, ok := rc.Params[key]
	return val, ok
}

// NewCtx builds a unified request context for handlers using the net/http primitives.
func NewCtx(w http.ResponseWriter, r *http.Request, params map[string]string) *Ctx {
	return newCtxWithLogger(w, r, params, nil)
}

// newCtxWithLogger allows injecting a logger while keeping NewCtx minimal for compatibility.
func newCtxWithLogger(w http.ResponseWriter, r *http.Request, params map[string]string, logger log.StructuredLogger) *Ctx {
	if params == nil {
		params = map[string]string{}
	}

	traceID := TraceIDFromContext(r.Context())
	deadline, _ := r.Context().Deadline()

	if logger == nil {
		logger = log.NewGLogger()
	}

	return &Ctx{
		W:        w,
		R:        r,
		Params:   params,
		Query:    r.URL.Query(),
		Headers:  r.Header,
		ClientIP: clientIPFromRequest(r),
		Logger:   logger,
		TraceID:  traceID,
		Deadline: deadline,
	}
}

func (c *Ctx) ErrorJSON(status int, errCode string, message string, details map[string]any) error {
	payload := APIError{
		Status:   status,
		Code:     errCode,
		Message:  message,
		Details:  details,
		TraceID:  c.TraceID,
		Category: CategoryBusiness,
	}
	return c.JSON(status, payload)
}

// JSON writes a JSON response with the given status code.
func (c *Ctx) JSON(status int, data any) error {
	c.W.Header().Set("Content-Type", "application/json")
	c.W.WriteHeader(status)
	return json.NewEncoder(c.W).Encode(data)
}

// Text writes a plain text response with the given status code.
func (c *Ctx) Text(status int, text string) error {
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(status)
	_, err := io.WriteString(c.W, text)
	return err
}

// Bytes writes a binary response with the given status code.
func (c *Ctx) Bytes(status int, data []byte) error {
	c.W.Header().Set("Content-Type", "application/octet-stream")
	c.W.WriteHeader(status)
	_, err := c.W.Write(data)
	return err
}

// Redirect sends a redirect response to the client.
func (c *Ctx) Redirect(status int, location string) error {
	http.Redirect(c.W, c.R, location, status)
	return nil
}

// File serves a file to the client.
func (c *Ctx) File(path string) error {
	http.ServeFile(c.W, c.R, path)
	return nil
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
	if !ok || strings.TrimSpace(val) == "" {
		return "", errors.New("missing param: " + key)
	}
	return val, nil
}

// BindJSON binds the request JSON body to the provided destination structure.
// It performs minimal validation and returns a BindError on failure.
func (c *Ctx) BindJSON(dst any) error {
	data, err := c.bodyBytes()
	if err != nil {
		return &BindError{Status: http.StatusBadRequest, Message: "failed to read request body", Err: err}
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return &BindError{Status: http.StatusBadRequest, Message: "request body is empty"}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	// DisallowUnknownFields could be enabled if you want strict mode:
	// decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return &BindError{Status: http.StatusBadRequest, Message: "invalid JSON payload", Err: err}
	}

	// Ensure no trailing data
	if decoder.Decode(&struct{}{}) != io.EOF {
		return &BindError{Status: http.StatusBadRequest, Message: "unexpected extra JSON data"}
	}

	return nil
}

func (c *Ctx) bodyBytes() ([]byte, error) {
	c.bodyOnce.Do(func() {
		c.body, c.bodyErr = io.ReadAll(c.R.Body)
		if c.bodyErr == nil {
			c.R.Body = io.NopCloser(bytes.NewBuffer(c.body))
		}
	})
	return c.body, c.bodyErr
}

func clientIPFromRequest(r *http.Request) string {
	// Prioritize standard proxy headers while avoiding common pitfalls like multiple values.
	if ip := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); ip != "" {
		return ip
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
		rc := RequestContextFrom(r.Context())
		ctx := newCtxWithLogger(w, r, rc.Params, logger)
		h(ctx)
	})
}

// ValidateCtxHandler returns an error when the handler is nil to give clearer feedback to callers.
func ValidateCtxHandler(h CtxHandlerFunc) error {
	if h == nil {
		return errors.New("context handler cannot be nil")
	}
	return nil
}
