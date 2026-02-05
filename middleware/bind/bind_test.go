package bind

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

type testPayload struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required" mask:"true"`
}

func TestBindJSONSuccess(t *testing.T) {
	body := []byte(`{"name":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler := BindJSON[testPayload](JSONOptions{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := FromRequest[testPayload](r)
		if !ok || p == nil {
			t.Fatalf("expected payload in context")
		}
		if p.Name != "alice" {
			t.Fatalf("unexpected name: %s", p.Name)
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestBindJSONValidationError(t *testing.T) {
	body := []byte(`{"name":"","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler := BindJSON[testPayload](JSONOptions{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Category != contract.CategoryValidation {
		t.Fatalf("expected validation category, got %s", resp.Error.Category)
	}
	fieldsRaw, ok := resp.Error.Details["fields"]
	if !ok {
		t.Fatalf("expected field errors in details")
	}
	fields, ok := fieldsRaw.([]any)
	if !ok || len(fields) == 0 {
		t.Fatalf("expected non-empty field errors, got %T", fieldsRaw)
	}
}

func TestRedactor(t *testing.T) {
	redactor := DefaultRedactor()
	payload := testPayload{Name: "alice", Password: "secret"}
	result := redactor.Redact(payload)

	data, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if data["password"] == "secret" {
		t.Fatalf("expected password to be redacted")
	}
}
