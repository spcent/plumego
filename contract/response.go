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

	buf := getJSONBuffer()
	defer putJSONBuffer(buf)

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(buf.Bytes())
	return err
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
