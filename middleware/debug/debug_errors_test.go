package debug

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net"
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
	mw := Middleware(Config{NotFoundHint: "/_debug/routes"})
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
	mw := Middleware(DefaultConfig())
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

func TestDebugErrorsCaptureLimitPassesThroughOriginalResponse(t *testing.T) {
	mw := Middleware(Config{
		IncludeBody:  true,
		MaxBodyBytes: 4,
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("abcdef"))
		_, _ = w.Write([]byte("gh"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/large-error", nil)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.Code)
	}
	if got := resp.Body.String(); got != "abcdefgh" {
		t.Fatalf("expected original body to pass through after capture limit, got %q", got)
	}
	if strings.Contains(resp.Body.String(), "response_preview") {
		t.Fatalf("debug replacement should be skipped after capture limit, got %q", resp.Body.String())
	}
}

func TestDebugErrorsIncludeBodyPreviewStillTruncates(t *testing.T) {
	mw := Middleware(Config{
		IncludeBody:  true,
		MaxBodyBytes: 4096,
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(strings.Repeat("x", 2048)))
	}))

	req := httptest.NewRequest(http.MethodGet, "/preview", nil)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	var payload contract.ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	preview, ok := payload.Error.Details["response_preview"].(string)
	if !ok {
		t.Fatalf("expected response preview detail, got %+v", payload.Error.Details)
	}
	if len(preview) != 1027 {
		t.Fatalf("preview length = %d, want 1027", len(preview))
	}
	if !strings.HasSuffix(preview, "...") {
		t.Fatalf("expected truncated preview suffix, got %q", preview[len(preview)-8:])
	}
}

func TestDebugErrorsReplacementDropsStaleContentLength(t *testing.T) {
	mw := Middleware(DefaultConfig())
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", "4")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failure"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/stale-length", nil)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusInternalServerError)
	}
	if got := resp.Header().Get("Content-Length"); got != "" {
		t.Fatalf("Content-Length = %q, want empty", got)
	}
	assertJSONContentType(t, resp.Header().Get("Content-Type"))
	var payload contract.ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode replacement body: %v", err)
	}
}

// TestDebugErrorsZeroValueConfig confirms that a zero-value Config
// does not auto-enable IncludeRequest or IncludeQuery (the old piecemeal merge
// silently defaulted those to true).
func TestDebugErrorsZeroValueConfig(t *testing.T) {
	mw := Middleware(Config{})
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
	mw := Middleware(DefaultConfig())
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
	mw := Middleware(DefaultConfig())
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

func TestDebugErrorsFlushCommitsPassthrough(t *testing.T) {
	mw := Middleware(DefaultConfig())
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("data: one\n\n"))
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("debug recorder did not expose flusher")
		}
		flusher.Flush()
		_, _ = w.Write([]byte("data: two\n\n"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	resp := &debugFlushWriter{ResponseRecorder: httptest.NewRecorder()}
	h.ServeHTTP(resp, req)

	if resp.flushed == 0 {
		t.Fatal("flush was not forwarded")
	}
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusInternalServerError)
	}
	if got := resp.Body.String(); got != "data: one\n\ndata: two\n\n" {
		t.Fatalf("body = %q, want original stream body", got)
	}
	if strings.Contains(resp.Body.String(), `"code"`) {
		t.Fatalf("flush path should not replace body with debug JSON: %q", resp.Body.String())
	}
}

func TestDebugErrorsHijackPassesThrough(t *testing.T) {
	mw := Middleware(DefaultConfig())
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("debug recorder did not expose hijacker")
		}
		conn, _, err := hijacker.Hijack()
		if err != nil {
			t.Fatalf("hijack failed: %v", err)
		}
		if conn != nil {
			_ = conn.Close()
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/raw", nil)
	writer := &debugHijackWriter{}
	h.ServeHTTP(writer, req)

	if !writer.hijacked {
		t.Fatal("underlying writer was not hijacked")
	}
	if writer.status != 0 {
		t.Fatalf("debug hijack path wrote status %d, want no HTTP status", writer.status)
	}
}

type debugFlushWriter struct {
	*httptest.ResponseRecorder
	flushed int
}

func (w *debugFlushWriter) Flush() {
	w.flushed++
	w.ResponseRecorder.Flush()
}

type debugHijackWriter struct {
	header   http.Header
	status   int
	hijacked bool
}

func (w *debugHijackWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *debugHijackWriter) WriteHeader(status int) {
	w.status = status
}

func (w *debugHijackWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (w *debugHijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	server, client := net.Pipe()
	_ = client.Close()
	var rw bytes.Buffer
	return server, bufio.NewReadWriter(bufio.NewReader(&rw), bufio.NewWriter(&rw)), nil
}
