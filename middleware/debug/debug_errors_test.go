package debug

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestDebugErrorsNotFound(t *testing.T) {
	mw := DebugErrors(DebugErrorConfig{NotFoundHint: "/_debug/routes"})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
	if ct := resp.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected json content type, got %q", ct)
	}

	var payload contract.ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Error.Code != "not_found" {
		t.Fatalf("expected code not_found, got %q", payload.Error.Code)
	}
	if hint, ok := payload.Error.Details["hint"]; !ok || hint != "/_debug/routes" {
		t.Fatalf("expected hint to be set")
	}
}

func TestDebugErrorsPassThroughJSON(t *testing.T) {
	mw := DebugErrors(DefaultDebugErrorConfig())
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad"}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/bad", nil)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if got := strings.TrimSpace(resp.Body.String()); got != `{"error":"bad"}` {
		t.Fatalf("expected body to pass through, got %q", got)
	}
}

// TestDebugErrorsZeroValueConfig confirms that a zero-value DebugErrorConfig
// does not auto-enable IncludeRequest or IncludeQuery (the old piecemeal merge
// silently defaulted those to true).
func TestDebugErrorsZeroValueConfig(t *testing.T) {
	mw := DebugErrors(DebugErrorConfig{})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing?foo=bar", nil)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	var payload contract.ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := payload.Error.Details["method"]; ok {
		t.Error("expected method not present in details when IncludeRequest is false")
	}
	if _, ok := payload.Error.Details["path"]; ok {
		t.Error("expected path not present in details when IncludeRequest is false")
	}
	if _, ok := payload.Error.Details["query"]; ok {
		t.Error("expected query not present in details when IncludeQuery is false")
	}
}

func TestDebugErrorsSkipUpgrade(t *testing.T) {
	mw := DebugErrors(DefaultDebugErrorConfig())
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("upgrade"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if got := strings.TrimSpace(resp.Body.String()); got != "upgrade" {
		t.Fatalf("expected body to pass through, got %q", got)
	}
}
