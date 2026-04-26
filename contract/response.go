package contract

import (
	"encoding/json"
	"net/http"
)

// HeaderContentType is the HTTP Content-Type header name.
const HeaderContentType = "Content-Type"

// ContentTypeJSON is the MIME type for JSON responses.
const ContentTypeJSON = "application/json"

// Response represents a standardized success response payload.
// It can be used together with WriteResponse for consistent JSON output.
type Response struct {
	Data      any            `json:"data,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
}

// WriteJSON writes the payload as JSON with the given HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) error {
	if w == nil {
		return ErrResponseWriterNil
	}
	status, _ = normalizeResponseHTTPStatus(status)
	if statusDisallowsBody(status) {
		w.WriteHeader(status)
		return nil
	}

	buf := getJSONBuffer()
	defer putJSONBuffer(buf)

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return err
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(status)
	_, err := w.Write(buf.Bytes())
	return err
}

// WriteResponse writes a standardized success response and injects request id when available.
func WriteResponse(w http.ResponseWriter, r *http.Request, status int, data any, meta map[string]any) error {
	resp := Response{
		Data: data,
		Meta: meta,
	}
	if r != nil {
		if requestID := RequestIDFromContext(r.Context()); requestID != "" {
			resp.RequestID = requestID
		}
	}
	return WriteJSON(w, status, resp)
}

func statusDisallowsBody(status int) bool {
	return (status >= 100 && status < 200) || status == http.StatusNoContent || status == http.StatusNotModified
}

func normalizeResponseHTTPStatus(status int) (int, bool) {
	if status < 100 || status > 599 {
		return http.StatusInternalServerError, true
	}
	return status, false
}
