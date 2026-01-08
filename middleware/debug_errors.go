package middleware

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
)

// DebugErrorConfig controls how debug error responses are formatted.
type DebugErrorConfig struct {
	IncludeRequest bool
	IncludeQuery   bool
	IncludeBody    bool
	NotFoundHint   string
}

// DefaultDebugErrorConfig returns a safe default for debug errors.
func DefaultDebugErrorConfig() DebugErrorConfig {
	return DebugErrorConfig{
		IncludeRequest: true,
		IncludeQuery:   true,
	}
}

// DebugErrors replaces empty/plain error responses with structured JSON.
func DebugErrors(config DebugErrorConfig) Middleware {
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
	copyHeader(w.Header(), r.header)
	w.WriteHeader(r.statusCode())
	_, _ = w.Write(r.body.Bytes())
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
