package contract

import (
	"encoding/json"
	"net/http"

	log "github.com/spcent/plumego/log"
)

// ErrorCategory describes the high-level class of an API error for observability.
type ErrorCategory string

const (
	// CategoryClient covers 4xx errors caused by client input.
	CategoryClient ErrorCategory = "client_error"
	// CategoryServer covers 5xx errors caused by infrastructure or server logic.
	CategoryServer ErrorCategory = "server_error"
	// CategoryBusiness covers domain-specific validation or invariants.
	CategoryBusiness ErrorCategory = "business_error"
)

// APIError represents a normalized error payload for HTTP responses and logging.
type APIError struct {
	Status   int            `json:"-"`
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Category ErrorCategory  `json:"category"`
	TraceID  string         `json:"trace_id,omitempty"`
	Details  map[string]any `json:"details,omitempty"`
}

// ErrorResponse wraps the error payload for consistent JSON responses.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// WriteError writes a structured error response with trace context when available.
func WriteError(w http.ResponseWriter, r *http.Request, err APIError) {
	if err.Status == 0 {
		err.Status = http.StatusInternalServerError
	}

	if err.Code == "" {
		err.Code = http.StatusText(err.Status)
	}

	if err.Category == "" {
		if err.Status >= 500 {
			err.Category = CategoryServer
		} else {
			err.Category = CategoryClient
		}
	}

	if r != nil {
		err.TraceID = TraceIDFromContext(r.Context())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)

	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err})
}

// ErrorLogger converts an error into structured logging fields while preserving correlation ids.
func ErrorLogger(logger log.StructuredLogger, r *http.Request, err APIError) {
	if logger == nil {
		return
	}

	fields := log.Fields{
		"code":     err.Code,
		"status":   err.Status,
		"category": err.Category,
	}

	if err.TraceID == "" && r != nil {
		err.TraceID = TraceIDFromContext(r.Context())
	}

	if err.TraceID != "" {
		fields["trace_id"] = err.TraceID
	}

	for k, v := range err.Details {
		fields[k] = v
	}

	logger.WithFields(fields).Error(err.Message, nil)
}
