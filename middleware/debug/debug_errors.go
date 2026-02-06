package debug

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/utils"
)

// DebugErrorConfig controls how debug error responses are formatted.
//
// This middleware is useful during development to provide detailed error information.
// It replaces empty or plain text error responses with structured JSON error messages.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/debug"
//
//	config := debug.DebugErrorConfig{
//		IncludeRequest: true,  // Include request method and path
//		IncludeQuery:   true,  // Include query parameters
//		IncludeBody:    false, // Don't include response body (security)
//		NotFoundHint:   "Try /api/v1/users", // Hint for 404 errors
//	}
//	handler := debug.DebugErrors(config)(myHandler)
//
// Security note: This middleware should only be used in development environments.
// In production, consider using a proper error logging and monitoring system.
type DebugErrorConfig struct {
	// IncludeRequest controls whether to include request method and path in error details
	IncludeRequest bool

	// IncludeQuery controls whether to include query parameters in error details
	IncludeQuery bool

	// IncludeBody controls whether to include response body in error details
	// Note: This may expose sensitive information, use with caution
	IncludeBody bool

	// NotFoundHint provides a hint message for 404 errors
	NotFoundHint string
}

// DefaultDebugErrorConfig returns a safe default for debug errors.
func DefaultDebugErrorConfig() DebugErrorConfig {
	return DebugErrorConfig{
		IncludeRequest: true,
		IncludeQuery:   true,
	}
}

// DebugErrors replaces empty/plain error responses with structured JSON.
//
// This middleware intercepts error responses and replaces them with structured JSON
// error messages containing additional debugging information.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware/debug"
//
//	// Use default configuration
//	handler := debug.DebugErrors(debug.DefaultDebugErrorConfig())(myHandler)
//
//	// Or with custom configuration
//	config := debug.DebugErrorConfig{
//		IncludeRequest: true,
//		IncludeQuery:   true,
//		NotFoundHint:   "Try /api/v1/users",
//	}
//	handler := debug.DebugErrors(config)(myHandler)
//
// The middleware skips debugging for:
//   - WebSocket connections
//   - Server-Sent Events (SSE)
//   - CONNECT requests
//   - Responses with non-plain-text content types
//
// Error response format:
//
//	{
//	  "status": 404,
//	  "code": "not_found",
//	  "message": "Not Found",
//	  "category": "client",
//	  "details": {
//	    "method": "GET",
//	    "path": "/api/missing",
//	    "query": "param=value",
//	    "hint": "Try /api/v1/users"
//	  }
//	}
func DebugErrors(config DebugErrorConfig) middleware.Middleware {
	cfg := DefaultDebugErrorConfig()
	if config.NotFoundHint != "" {
		cfg.NotFoundHint = config.NotFoundHint
	}
	if config.IncludeBody {
		cfg.IncludeBody = true
	}
	if !config.IncludeRequest {
		cfg.IncludeRequest = false
	}
	if !config.IncludeQuery {
		cfg.IncludeQuery = false
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldSkipDebugErrors(r) {
				next.ServeHTTP(w, r)
				return
			}

			rec := newDebugErrorRecorder()
			next.ServeHTTP(rec, r)

			status := rec.statusCode()
			body := rec.body.Bytes()

			if status < http.StatusBadRequest || !shouldReplaceError(rec.header, body) {
				rec.flushTo(w)
				return
			}

			copyHeader(w.Header(), rec.header)
			contract.WriteError(w, r, debugErrorPayload(status, r, cfg, body))
		})
	}
}

type debugErrorRecorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newDebugErrorRecorder() *debugErrorRecorder {
	return &debugErrorRecorder{header: make(http.Header)}
}

func (r *debugErrorRecorder) Header() http.Header {
	return r.header
}

func (r *debugErrorRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
}

func (r *debugErrorRecorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.body.Write(p)
}

func (r *debugErrorRecorder) statusCode() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func (r *debugErrorRecorder) flushTo(w http.ResponseWriter) {
	// SECURITY NOTE: This middleware records error responses for debugging.
	// The body contains error information from upstream handlers, not user input.
	// This does not introduce XSS vulnerabilities as it passes through existing responses.
	// XSS protection should be implemented in handlers that generate HTML using utils/html.go.
	copyHeader(w.Header(), r.header)
	utils.EnsureNoSniff(w.Header())
	w.WriteHeader(r.statusCode())
	_, _ = utils.SafeWrite(w, r.body.Bytes())
}

func shouldSkipDebugErrors(r *http.Request) bool {
	if r == nil {
		return false
	}

	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return true
	}

	accept := strings.ToLower(r.Header.Get("Accept"))
	if strings.Contains(accept, "text/event-stream") {
		return true
	}

	if r.Method == http.MethodConnect {
		return true
	}

	return false
}

func shouldReplaceError(header http.Header, body []byte) bool {
	if len(body) == 0 {
		return true
	}

	contentType := strings.ToLower(strings.TrimSpace(header.Get("Content-Type")))
	return strings.HasPrefix(contentType, "text/plain")
}

func debugErrorPayload(status int, r *http.Request, cfg DebugErrorConfig, body []byte) contract.APIError {
	message := http.StatusText(status)
	code := strings.ToLower(strings.ReplaceAll(message, " ", "_"))
	category := contract.CategoryClient
	if status >= http.StatusInternalServerError {
		category = contract.CategoryServer
	}

	details := map[string]any{}
	if cfg.IncludeRequest {
		details["method"] = r.Method
		details["path"] = r.URL.Path
	}
	if cfg.IncludeQuery && r.URL.RawQuery != "" {
		details["query"] = r.URL.RawQuery
	}
	if status == http.StatusNotFound && cfg.NotFoundHint != "" {
		details["hint"] = cfg.NotFoundHint
	}
	if cfg.IncludeBody && len(body) > 0 {
		details["response_preview"] = truncateBody(body, 1024)
	}
	if len(details) == 0 {
		details = nil
	}

	return contract.APIError{
		Status:   status,
		Code:     code,
		Message:  message,
		Category: category,
		Details:  details,
	}
}

func truncateBody(body []byte, limit int) string {
	if limit <= 0 || len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "..."
}

func copyHeader(dst, src http.Header) {
	for key, values := range src {
		cloned := make([]string, len(values))
		copy(cloned, values)
		dst[key] = cloned
	}
}
