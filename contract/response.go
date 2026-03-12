package contract

import (
	"encoding/json"
	"net/http"
)

// Response represents a standardized success response payload.
// It can be used together with WriteResponse for consistent JSON output.
type Response struct {
	Data    any            `json:"data,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
}

// WriteJSON writes the payload as JSON with the given HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) error {
	if w == nil {
		return ErrResponseWriterNil
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

// WriteResponse writes a standardized success response and injects trace id when available.
func WriteResponse(w http.ResponseWriter, r *http.Request, status int, data any, meta map[string]any) error {
	resp := Response{
		Data: data,
		Meta: meta,
	}
	if r != nil {
		if traceID := TraceIDFromContext(r.Context()); traceID != "" {
			resp.TraceID = traceID
		}
	}
	return WriteJSON(w, status, resp)
}

func WriteContractResponse(ctx *Ctx, status int, data any) {
	if ctx == nil {
		return
	}
	_ = ctx.Response(status, data, nil)
}

func WriteContractError(ctx *Ctx, status int, code, message string) {
	if ctx == nil {
		return
	}
	apiErr := APIError{
		Status:   status,
		Code:     code,
		Message:  message,
		Category: CategoryForStatus(status),
	}
	WriteError(ctx.W, ctx.R, apiErr)
}

func WriteHTTPResponse(w http.ResponseWriter, r *http.Request, status int, data any) {
	_ = WriteResponse(w, r, status, data, nil)
}

func WriteHTTPError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	apiErr := APIError{
		Status:   status,
		Code:     code,
		Message:  message,
		Category: CategoryForStatus(status),
	}
	WriteError(w, r, apiErr)
}
