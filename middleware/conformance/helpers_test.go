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

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse JSON error response: %v", err)
	}

	errObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected top-level error object, got %#v", payload)
	}

	for _, field := range []string{"code", "message", "category"} {
		if _, ok := errObj[field]; !ok {
			t.Fatalf("expected error.%s field in payload %#v", field, payload)
		}
	}

	if got, _ := errObj["code"].(string); got != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, got)
	}
}

func execute(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
