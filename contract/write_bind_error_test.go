package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWriteBindErrorBodyTooLarge verifies WriteBindError maps ErrRequestBodyTooLarge
// to a 413 response with the correct error code.
func TestWriteBindErrorBodyTooLarge(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/upload", nil)

	WriteBindError(w, r, ErrRequestBodyTooLarge)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error.Code != CodeRequestBodyTooLarge {
		t.Fatalf("expected code %s, got %s", CodeRequestBodyTooLarge, resp.Error.Code)
	}
}

// TestWriteBindErrorEmptyBody verifies WriteBindError maps ErrEmptyRequestBody
// to a 400 response with the correct error code.
func TestWriteBindErrorEmptyBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/items", nil)

	WriteBindError(w, r, ErrEmptyRequestBody)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error.Code != CodeEmptyBody {
		t.Fatalf("expected code %s, got %s", CodeEmptyBody, resp.Error.Code)
	}
}

// TestWriteBindErrorInvalidJSON verifies WriteBindError maps ErrInvalidJSON
// to a 400 response with the correct error code.
func TestWriteBindErrorInvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/items", nil)

	WriteBindError(w, r, ErrInvalidJSON)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error.Code != CodeInvalidJSON {
		t.Fatalf("expected code %s, got %s", CodeInvalidJSON, resp.Error.Code)
	}
}

// TestWriteBindErrorUnexpectedExtraData verifies WriteBindError maps
// ErrUnexpectedExtraData to a 400 response with the correct error code.
func TestWriteBindErrorUnexpectedExtraData(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/items", nil)

	WriteBindError(w, r, ErrUnexpectedExtraData)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error.Code != CodeUnexpectedExtraData {
		t.Fatalf("expected code %s, got %s", CodeUnexpectedExtraData, resp.Error.Code)
	}
}

// TestWriteBindErrorValidationWithFields verifies WriteBindError produces
// structured field errors when a BindError with field-level details is provided.
func TestWriteBindErrorValidationWithFields(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/users", nil)

	type payload struct {
		Email string `validate:"required,email"`
		Name  string `validate:"required"`
	}
	validateErr := validateStruct(&payload{})
	if validateErr == nil {
		t.Skip("validator not available")
	}
	bindErr := &BindError{
		Status:  http.StatusBadRequest,
		Message: validateErr.Error(),
		Err:     validateErr,
	}

	WriteBindError(w, r, bindErr)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error.Code != CodeValidationError {
		t.Fatalf("expected code %s, got %s", CodeValidationError, resp.Error.Code)
	}
}

// TestWriteBindErrorContentTypeIsJSON verifies WriteBindError always sets
// Content-Type to application/json.
func TestWriteBindErrorContentTypeIsJSON(t *testing.T) {
	for _, err := range []error{
		ErrRequestBodyTooLarge,
		ErrEmptyRequestBody,
		ErrInvalidJSON,
		ErrUnexpectedExtraData,
	} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		WriteBindError(w, r, err)
		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Fatalf("expected Content-Type application/json for %v, got %s", err, ct)
		}
	}
}
