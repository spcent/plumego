package requestid

import (
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
)

// RequestIDFromRequest extracts the canonical request id from context or headers.
func RequestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if id := contract.RequestIDFromContext(r.Context()); id != "" {
		return id
	}
	if id := strings.TrimSpace(r.Header.Get(contract.RequestIDHeader)); id != "" {
		return id
	}
	return ""
}

// AttachRequestID stores the request id in context and headers.
func AttachRequestID(w http.ResponseWriter, r *http.Request, id string, includeInRequest bool) *http.Request {
	if r == nil {
		return r
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return r
	}
	ctx := contract.WithRequestID(r.Context(), id)
	// Populate the log package's self-contained request-ID carrier so the
	// *Ctx logging methods attach the correlation ID to request-scoped entries.
	ctx = log.WithRequestID(ctx, id)
	if includeInRequest {
		r.Header.Set(contract.RequestIDHeader, id)
	}
	if w != nil {
		w.Header().Set(contract.RequestIDHeader, id)
	}
	return r.WithContext(ctx)
}

// EnsureRequestID returns an existing request id or generates one when missing.
func EnsureRequestID(r *http.Request, generate func() string) string {
	if id := RequestIDFromRequest(r); id != "" {
		return id
	}
	if generate != nil {
		if id := strings.TrimSpace(generate()); id != "" {
			return id
		}
	}
	return NewRequestID()
}
