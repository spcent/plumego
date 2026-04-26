package debug

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
)

func assertJSONContentType(t *testing.T, contentType string) {
	t.Helper()

	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected json content type, got %q", contentType)
	}
}

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
	assertJSONContentType(t, resp.Header().Get("Content-Type"))

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

func TestDebugErrorsPassThroughResponseDeclaredStream(t *testing.T) {
	mw := DebugErrors(DefaultDebugErrorConfig())
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("data: failed\n\n"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
	if got := strings.TrimSpace(resp.Body.String()); got != "data: failed" {
		t.Fatalf("expected stream body to pass through, got %q", got)
	}
	if got := resp.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("expected stream content type to pass through, got %q", got)
	}
}
