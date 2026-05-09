package contract

import (
	"encoding/json"
	"net/http"
)

// HeaderContentType is the HTTP Content-Type header name.
const HeaderContentType = "Content-Type"

// ContentTypeJSON is the MIME type for JSON responses.
const ContentTypeJSON = "application/json"

type response struct {
	Data      any            `json:"data,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) error {
	if w == nil {
		return ErrResponseWriterNil
	}
	status = normalizeResponseHTTPStatus(status)
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
// For body-eligible statuses, nil data and nil meta encode as an empty JSON
// object because the response envelope omits empty fields.
func WriteResponse(w http.ResponseWriter, r *http.Request, status int, data any, meta map[string]any) error {
	resp := response{
		Data: data,
		Meta: meta,
	}
	if r != nil {
		if requestID := RequestIDFromContext(r.Context()); requestID != "" {
			resp.RequestID = requestID
		}
	}
	return writeJSON(w, status, resp)
}

func statusDisallowsBody(status int) bool {
	return (status >= 100 && status < 200) || status == http.StatusNoContent || status == http.StatusNotModified
}

func normalizeResponseHTTPStatus(status int) int {
	if status < 100 || status > 599 {
		return http.StatusInternalServerError
	}
	return status
}
