package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/utils/pool"
	"github.com/spcent/plumego/validator"
)

// RequestContext contains request-scoped data that should be shared across middleware and handlers.
// It preserves compatibility with the standard library by living inside the request's context.
type RequestContext struct {
	Params map[string]string
}

type RequestContextKey struct{}

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

	// Request state tracking
	startedAt          time.Time
	bodySize           atomic.Int64
	body               []byte
	bodyErr            error
	bodyReadOnce       sync.Once
	compressionEnabled atomic.Bool
	cancel             context.CancelFunc
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

// DefaultRequestConfig provides sensible defaults for request processing.
var DefaultRequestConfig = &RequestConfig{
	MaxBodySize:       10 * 1024 * 1024, // 10MB
	EnableBodyCache:   true,
	EnableCompression: false,
	RequestTimeout:    30 * time.Second,
}

func defaultRequestConfig() *RequestConfig {
	if DefaultRequestConfig == nil {
		return &RequestConfig{}
	}
	cfg := *DefaultRequestConfig
	return &cfg
}

// newCtxWithLogger allows injecting a logger while keeping NewCtx minimal for compatibility.
func newCtxWithLogger(w http.ResponseWriter, r *http.Request, params map[string]string, logger log.StructuredLogger) *Ctx {
	if params == nil {
		params = map[string]string{}
	}

	config := defaultRequestConfig()

	var cancel context.CancelFunc
	deadline, hasDeadline := r.Context().Deadline()
	if !hasDeadline && config.RequestTimeout > 0 {
		timeoutCtx, cancelFunc := context.WithTimeout(r.Context(), config.RequestTimeout)
		cancel = cancelFunc
		r = r.WithContext(timeoutCtx)
		deadline, _ = timeoutCtx.Deadline()
	}

	traceID := TraceIDFromContext(r.Context())

	if logger == nil {
		logger = log.NewGLogger()
	}

	// Detect compression
	compressionEnabled := false
	if config.EnableCompression {
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
		Config:   config,

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

	// Use pooled buffer for encoding to reduce allocations
	buf := pool.GetBuffer()
	defer pool.PutBuffer(buf)

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return err
	}

	// Write the buffer content to response writer
	_, err := c.W.Write(buf.Bytes())
	return err
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
		if errors.Is(err, ErrRequestBodyTooLarge) {
			return &BindError{Status: http.StatusRequestEntityTooLarge, Message: ErrRequestBodyTooLarge.Error(), Err: err}
		}
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

// BindAndValidateJSON binds the request body to dst and validates it using struct tags.
func (c *Ctx) BindAndValidateJSON(dst any) error {
	if err := c.BindJSON(dst); err != nil {
		return err
	}

	if err := validator.Validate(dst); err != nil {
		return &BindError{Status: http.StatusBadRequest, Message: err.Error(), Err: err}
	}

	return nil
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

func (c *Ctx) bodyBytes() ([]byte, error) {
	c.bodyReadOnce.Do(func() {
		reader := io.Reader(c.R.Body)
		maxBodySize := int64(0)
		if c.Config != nil && c.Config.MaxBodySize > 0 {
			maxBodySize = c.Config.MaxBodySize
			if c.W != nil {
				reader = http.MaxBytesReader(c.W, c.R.Body, maxBodySize)
			} else {
				reader = io.LimitReader(reader, maxBodySize+1)
			}
		}

		c.body, c.bodyErr = io.ReadAll(reader)
		if c.bodyErr != nil {
			var maxErr *http.MaxBytesError
			if errors.As(c.bodyErr, &maxErr) {
				c.bodyErr = ErrRequestBodyTooLarge
				c.body = nil
			}
			return
		}
		if maxBodySize > 0 && int64(len(c.body)) > maxBodySize {
			c.bodyErr = ErrRequestBodyTooLarge
			c.body = nil
			return
		}
		c.bodySize.Store(int64(len(c.body)))
		if c.Config == nil || c.Config.EnableBodyCache {
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
		defer ctx.Close()
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

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	ID    string
	Event string
	Data  string
	Retry time.Duration
}

// SSEWriter is a helper for writing Server-Sent Events.
type SSEWriter struct {
	w http.ResponseWriter
	f http.Flusher
}

// NewSSEWriter creates a new SSE writer.
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil
	}
	return &SSEWriter{w: w, f: f}
}

// Write writes an SSE event to the response.
func (sw *SSEWriter) Write(event SSEEvent) error {
	if event.ID != "" {
		if _, err := fmt.Fprintf(sw.w, "id: %s\n", event.ID); err != nil {
			return err
		}
	}
	if event.Event != "" {
		if _, err := fmt.Fprintf(sw.w, "event: %s\n", event.Event); err != nil {
			return err
		}
	}
	if event.Data != "" {
		if _, err := fmt.Fprintf(sw.w, "data: %s\n\n", event.Data); err != nil {
			return err
		}
	}
	if event.Retry > 0 {
		if _, err := fmt.Fprintf(sw.w, "retry: %d\n\n", int(event.Retry.Milliseconds())); err != nil {
			return err
		}
	}
	sw.f.Flush()
	return nil
}

func (c *Ctx) streamContext() context.Context {
	if c == nil || c.R == nil {
		return context.Background()
	}
	return c.R.Context()
}

func checkStreamContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	if ctx == nil {
		time.Sleep(delay)
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func validateChunkSize(chunkSize int) error {
	if chunkSize <= 0 {
		return ErrInvalidChunkSize
	}
	return nil
}

// RespondWithSSE starts a Server-Sent Events response.
// Returns an SSEWriter for sending events, or nil if SSE is not supported.
func (c *Ctx) RespondWithSSE() *SSEWriter {
	c.W.Header().Set("Content-Type", "text/event-stream")
	c.W.Header().Set("Cache-Control", "no-cache")
	c.W.Header().Set("Connection", "keep-alive")
	c.W.WriteHeader(http.StatusOK)

	return NewSSEWriter(c.W)
}

// RespondWithEventSource starts an event stream response.
// This is an alias for RespondWithSSE.
func (c *Ctx) RespondWithEventSource() *SSEWriter {
	return c.RespondWithSSE()
}

// StreamJSON streams JSON data line by line.
func (c *Ctx) StreamJSON(items []any) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	for _, item := range items {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if err := encoder.Encode(item); err != nil {
			return err
		}
		if f, ok := c.W.(http.Flusher); ok {
			f.Flush()
		}
	}

	return nil
}

// StreamText streams text data line by line.
func (c *Ctx) StreamText(lines []string) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	for _, line := range lines {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
			return err
		}
		if f, ok := c.W.(http.Flusher); ok {
			f.Flush()
		}
	}

	return nil
}

// StreamBinary streams binary data in chunks.
func (c *Ctx) StreamBinary(reader io.Reader, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/octet-stream")
	c.W.WriteHeader(http.StatusOK)

	buf := make([]byte, chunkSize)
	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if n > 0 {
			if err := checkStreamContext(ctx); err != nil {
				return err
			}
			if _, err := c.W.Write(buf[:n]); err != nil {
				return err
			}
			if f, ok := c.W.(http.Flusher); ok {
				f.Flush()
			}
		}
	}

	return nil
}

// StreamJSONWithChannel streams JSON data from a channel.
func (c *Ctx) StreamJSONWithChannel(itemChan <-chan any) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-itemChan:
			if !ok {
				return nil
			}
			if err := encoder.Encode(item); err != nil {
				return err
			}
			if f, ok := c.W.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// StreamTextWithChannel streams text data from a channel.
func (c *Ctx) StreamTextWithChannel(lineChan <-chan string) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case line, ok := <-lineChan:
			if !ok {
				return nil
			}
			if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
				return err
			}
			if f, ok := c.W.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// StreamSSEWithChannel streams Server-Sent Events from a channel.
func (c *Ctx) StreamSSEWithChannel(eventChan <-chan SSEEvent) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	sseWriter := c.RespondWithSSE()
	if sseWriter == nil {
		return errors.New("SSE not supported")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-eventChan:
			if !ok {
				return nil
			}
			if err := sseWriter.Write(event); err != nil {
				return err
			}
		}
	}
}

// StreamJSONWithGenerator streams JSON data using a generator function.
func (c *Ctx) StreamJSONWithGenerator(generator func() (any, error)) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		item, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := encoder.Encode(item); err != nil {
			return err
		}
		if f, ok := c.W.(http.Flusher); ok {
			f.Flush()
		}
	}

	return nil
}

// StreamTextWithGenerator streams text data using a generator function.
func (c *Ctx) StreamTextWithGenerator(generator func() (string, error)) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		line, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
			return err
		}
		if f, ok := c.W.(http.Flusher); ok {
			f.Flush()
		}
	}

	return nil
}

// StreamSSEWithGenerator streams Server-Sent Events using a generator function.
func (c *Ctx) StreamSSEWithGenerator(generator func() (SSEEvent, error)) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	sseWriter := c.RespondWithSSE()
	if sseWriter == nil {
		return errors.New("SSE not supported")
	}

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		event, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := sseWriter.Write(event); err != nil {
			return err
		}
	}

	return nil
}

// IsSSESupported checks if the response writer supports SSE.
func (c *Ctx) IsSSESupported() bool {
	_, ok := c.W.(http.Flusher)
	return ok
}

// SetSSEHeaders sets the standard SSE headers.
func (c *Ctx) SetSSEHeaders() {
	c.W.Header().Set("Content-Type", "text/event-stream")
	c.W.Header().Set("Cache-Control", "no-cache")
	c.W.Header().Set("Connection", "keep-alive")
}

// SetEventSourceHeaders sets the standard event source headers.
func (c *Ctx) SetEventSourceHeaders() {
	c.SetSSEHeaders()
}

// WriteSSE writes a single SSE event and flushes.
func (c *Ctx) WriteSSE(event SSEEvent) error {
	sseWriter := NewSSEWriter(c.W)
	if sseWriter == nil {
		return errors.New("SSE not supported")
	}
	return sseWriter.Write(event)
}

// WriteEventSource writes a single event source event and flushes.
func (c *Ctx) WriteEventSource(event SSEEvent) error {
	return c.WriteSSE(event)
}

// StreamJSONChunked streams JSON data in chunks with custom chunk size.
func (c *Ctx) StreamJSONChunked(items []any, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	for i, item := range items {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if err := encoder.Encode(item); err != nil {
			return err
		}

		// Flush after each chunk
		if (i+1)%chunkSize == 0 {
			if f, ok := c.W.(http.Flusher); ok {
				f.Flush()
			}
		}
	}

	// Final flush
	if f, ok := c.W.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

// StreamTextChunked streams text data in chunks with custom chunk size.
func (c *Ctx) StreamTextChunked(lines []string, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	for i, line := range lines {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
			return err
		}

		// Flush after each chunk
		if (i+1)%chunkSize == 0 {
			if f, ok := c.W.(http.Flusher); ok {
				f.Flush()
			}
		}
	}

	// Final flush
	if f, ok := c.W.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

// StreamSSEChunked streams Server-Sent Events in chunks with custom chunk size.
func (c *Ctx) StreamSSEChunked(events []SSEEvent, chunkSize int) error {
	if err := validateChunkSize(chunkSize); err != nil {
		return err
	}
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	sseWriter := c.RespondWithSSE()
	if sseWriter == nil {
		return errors.New("SSE not supported")
	}

	for i, event := range events {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		if err := sseWriter.Write(event); err != nil {
			return err
		}

		// Flush after each chunk
		if (i+1)%chunkSize == 0 {
			if f, ok := c.W.(http.Flusher); ok {
				f.Flush()
			}
		}
	}

	// Final flush
	if f, ok := c.W.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

// StreamJSONWithRetry streams JSON data with retry logic on error.
func (c *Ctx) StreamJSONWithRetry(generator func() (any, error), maxRetries int, retryDelay time.Duration) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "application/x-ndjson")
	c.W.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(c.W)
	retries := 0

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		item, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			if retries < maxRetries {
				if err := sleepWithContext(ctx, retryDelay); err != nil {
					return err
				}
				retries++
				continue
			}
			return err
		}

		retries = 0 // Reset retries on success

		if err := encoder.Encode(item); err != nil {
			return err
		}
		if f, ok := c.W.(http.Flusher); ok {
			f.Flush()
		}
	}

	return nil
}

// StreamTextWithRetry streams text data with retry logic on error.
func (c *Ctx) StreamTextWithRetry(generator func() (string, error), maxRetries int, retryDelay time.Duration) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(http.StatusOK)

	retries := 0

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		line, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			if retries < maxRetries {
				if err := sleepWithContext(ctx, retryDelay); err != nil {
					return err
				}
				retries++
				continue
			}
			return err
		}

		retries = 0 // Reset retries on success

		if _, err := fmt.Fprintf(c.W, "%s\n", line); err != nil {
			return err
		}
		if f, ok := c.W.(http.Flusher); ok {
			f.Flush()
		}
	}

	return nil
}

// StreamSSEWithRetry streams Server-Sent Events with retry logic on error.
func (c *Ctx) StreamSSEWithRetry(generator func() (SSEEvent, error), maxRetries int, retryDelay time.Duration) error {
	ctx := c.streamContext()
	if err := checkStreamContext(ctx); err != nil {
		return err
	}
	sseWriter := c.RespondWithSSE()
	if sseWriter == nil {
		return errors.New("SSE not supported")
	}

	retries := 0

	for {
		if err := checkStreamContext(ctx); err != nil {
			return err
		}
		event, err := generator()
		if err != nil {
			if err == io.EOF {
				break
			}
			if retries < maxRetries {
				if err := sleepWithContext(ctx, retryDelay); err != nil {
					return err
				}
				retries++
				continue
			}
			return err
		}

		retries = 0 // Reset retries on success

		if err := sseWriter.Write(event); err != nil {
			return err
		}
	}

	return nil
}
