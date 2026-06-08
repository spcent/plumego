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
	if !validResponseHTTPStatus(status) {
		return ErrInvalidResponseStatus
	}
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

func validResponseHTTPStatus(status int) bool {
	return status >= http.StatusOK && status <= 299
}

// errorPayload is the "error" field inside an error response envelope.
type errorPayload struct {
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Category ErrorCategory  `json:"category"`
	Type     ErrorType      `json:"type,omitempty"`
	Details  map[string]any `json:"details,omitempty"`
}

// errorResponse is the top-level error envelope:
//
//	{"error": {"code":"...", "message":"...", "category":"..."}, "request_id":"..."}
type errorResponse struct {
	Error     errorPayload `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
}

// WriteError writes a structured error response and injects the request id when available.
// It returns the encoding error, if any; callers may ignore it when the response
// headers have already been sent.
//
// Prefer building APIError values through NewErrorBuilder() so that required
// fields are always populated. WriteError still normalizes incomplete APIError
// values, but it does so deterministically with no package-global side effects.
func WriteError(w http.ResponseWriter, r *http.Request, err APIError) error {
	if w == nil {
		return ErrResponseWriterNil
	}

	err = normalizeAPIError(err)

	if err.requestID == "" && r != nil {
		if requestID := RequestIDFromContext(r.Context()); requestID != "" {
			err.requestID = requestID
		}
	}

	resp := errorResponse{
		Error:     errorPayloadFrom(err),
		RequestID: err.requestID,
	}

	buf := getJSONBuffer()
	defer putJSONBuffer(buf)

	if encErr := json.NewEncoder(buf).Encode(resp); encErr != nil {
		return encErr
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(err.status)
	_, writeErr := w.Write(buf.Bytes())
	return writeErr
}

func errorPayloadFrom(err APIError) errorPayload {
	return errorPayload{
		Code:     err.code,
		Message:  err.message,
		Category: err.category,
		Type:     err.errorType,
		Details:  cloneAnyMap(err.details),
	}
}
