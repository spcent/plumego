package conformance_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type callCountingHandler struct {
	calls int
}

type canonicalErrorEnvelope struct {
	Error struct {
		Code     string `json:"code"`
		Message  string `json:"message"`
		Category string `json:"category"`
	} `json:"error"`
}

func (h *callCountingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.calls++
	w.Header().Set("X-Conformance-Handler", "called")
	w.WriteHeader(http.StatusNoContent)
}

func assertCanonicalErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, expectedCode string) {
	t.Helper()

	if rec.Code < 400 {
		t.Fatalf("expected error status, got %d", rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json content type, got %q", got)
	}

	var payload canonicalErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse JSON error response: %v", err)
	}
	if payload.Error.Message == "" {
		t.Fatalf("expected error.message field in payload %#v", payload)
	}
	if payload.Error.Category == "" {
		t.Fatalf("expected error.category field in payload %#v", payload)
	}
	if payload.Error.Code != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, payload.Error.Code)
	}
}

func execute(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
