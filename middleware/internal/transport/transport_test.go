package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- EnsureNoSniff ---

func TestEnsureNoSniff_SetsHeader(t *testing.T) {
	h := make(http.Header)
	EnsureNoSniff(h)
	if got := h.Get(HeaderContentTypeNoSniff); got != ContentTypeNoSniffValue {
		t.Errorf("X-Content-Type-Options = %q, want %q", got, ContentTypeNoSniffValue)
	}
}

func TestEnsureNoSniff_PreservesExisting(t *testing.T) {
	h := make(http.Header)
	h.Set(HeaderContentTypeNoSniff, "custom")
	EnsureNoSniff(h)
	if got := h.Get(HeaderContentTypeNoSniff); got != "custom" {
		t.Errorf("existing header was overwritten, got %q", got)
	}
}

func TestEnsureNoSniff_NilHeader(t *testing.T) {
	// Must not panic.
	EnsureNoSniff(nil)
}

// --- SafeWrite ---

func TestSafeWrite_WritesBody(t *testing.T) {
	w := httptest.NewRecorder()
	n, err := SafeWrite(w, []byte("hello"))
	if err != nil {
		t.Fatalf("SafeWrite error: %v", err)
	}
	if n != 5 {
		t.Errorf("n = %d, want 5", n)
	}
	if w.Body.String() != "hello" {
		t.Errorf("body = %q, want hello", w.Body.String())
	}
	if w.Header().Get(HeaderContentTypeNoSniff) != ContentTypeNoSniffValue {
		t.Error("X-Content-Type-Options not set by SafeWrite")
	}
}

func TestSafeWrite_NilWriter(t *testing.T) {
	n, err := SafeWrite(nil, []byte("data"))
	if err != nil || n != 0 {
		t.Errorf("SafeWrite(nil) = (%d, %v), want (0, nil)", n, err)
	}
}

// --- AddVary ---

func TestAddVary_AppendsUniqueTokens(t *testing.T) {
	h := make(http.Header)
	h.Add("Vary", "Origin, Accept-Encoding")
	h.Add("Vary", "X-Existing")

	AddVary(h, "Accept-Encoding", "Access-Control-Request-Method", "origin", "X-New, X-Existing")

	got := h.Values("Vary")
	if len(got) != 4 {
		t.Fatalf("Vary values = %v, want 4 header values", got)
	}
	for _, want := range []string{"Origin", "Accept-Encoding", "X-Existing", "Access-Control-Request-Method", "X-New"} {
		if !headerValuesContainToken(got, want) {
			t.Fatalf("Vary values = %v, missing %q", got, want)
		}
	}
}

func TestAddVary_NilHeader(t *testing.T) {
	AddVary(nil, "Origin")
}

// --- ClientIP ---

func TestClientIP_ForwardedFor(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderForwardedFor, "1.2.3.4, 5.6.7.8")
	if ip := ClientIP(r); ip != "1.2.3.4" {
		t.Errorf("ClientIP = %q, want 1.2.3.4", ip)
	}
}

func TestClientIP_RealIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderRealIP, "9.10.11.12")
	if ip := ClientIP(r); ip != "9.10.11.12" {
		t.Errorf("ClientIP = %q, want 9.10.11.12", ip)
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:9999"
	if ip := ClientIP(r); ip != "10.0.0.1" {
		t.Errorf("ClientIP = %q, want 10.0.0.1", ip)
	}
}

func TestClientIP_NilRequest(t *testing.T) {
	if ip := ClientIP(nil); ip != "" {
		t.Errorf("ClientIP(nil) = %q, want empty", ip)
	}
}

// --- ResponseRecorder ---

func TestResponseRecorder_WriteAndStatus(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := NewResponseRecorder(underlying)

	rec.WriteHeader(http.StatusCreated)
	n, err := rec.Write([]byte("body"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != 4 {
		t.Errorf("n = %d, want 4", n)
	}
	if rec.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode = %d, want 201", rec.StatusCode())
	}
	if string(rec.Body()) != "body" {
		t.Errorf("Body = %q, want body", string(rec.Body()))
	}
	if rec.BytesWritten() != 4 {
		t.Errorf("BytesWritten = %d, want 4", rec.BytesWritten())
	}
	if underlying.Code != http.StatusCreated {
		t.Errorf("underlying status = %d, want 201", underlying.Code)
	}
}

func TestResponseRecorder_DefaultStatus200(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := NewResponseRecorder(underlying)

	_, _ = rec.Write([]byte("x"))
	if rec.StatusCode() != http.StatusOK {
		t.Errorf("default status = %d, want 200", rec.StatusCode())
	}
}

func TestResponseRecorder_WriteHeaderOnce(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := NewResponseRecorder(underlying)

	rec.WriteHeader(http.StatusAccepted)
	rec.WriteHeader(http.StatusTeapot) // second call should be ignored
	if rec.StatusCode() != http.StatusAccepted {
		t.Errorf("status after double WriteHeader = %d, want 202", rec.StatusCode())
	}
}

func TestResponseRecorder_HeaderForwarded(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := NewResponseRecorder(underlying)
	rec.Header().Set("X-Custom", "val")
	rec.WriteHeader(http.StatusOK)

	if underlying.Header().Get("X-Custom") != "val" {
		t.Error("custom header was not copied to underlying writer")
	}
}

// --- BufferedResponse ---

func TestBufferedResponse_BasicWrite(t *testing.T) {
	buf := NewBufferedResponse(0)
	n, err := buf.Write([]byte("buffered"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != 8 {
		t.Errorf("n = %d, want 8", n)
	}
	if string(buf.Body()) != "buffered" {
		t.Errorf("Body = %q, want buffered", string(buf.Body()))
	}
	if buf.BytesWritten() != 8 {
		t.Errorf("BytesWritten = %d, want 8", buf.BytesWritten())
	}
	if buf.Len() != 8 {
		t.Errorf("Len = %d, want 8", buf.Len())
	}
}

func TestBufferedResponse_MaxBytesOverflow(t *testing.T) {
	buf := NewBufferedResponse(5)
	_, _ = buf.Write([]byte("hello"))

	_, err := buf.Write([]byte("overflow"))
	if err == nil {
		t.Fatal("expected error on overflow, got nil")
	}
	if !buf.Overflowed() {
		t.Error("Overflowed() should be true after overflow")
	}
}

func TestBufferedResponse_StatusCode(t *testing.T) {
	buf := NewBufferedResponse(0)
	buf.WriteHeader(http.StatusNotFound)
	if buf.StatusCode() != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", buf.StatusCode())
	}
}

func TestBufferedResponse_DefaultStatus200(t *testing.T) {
	buf := NewBufferedResponse(0)
	_, _ = buf.Write([]byte("x"))
	if buf.StatusCode() != http.StatusOK {
		t.Errorf("default StatusCode = %d, want 200", buf.StatusCode())
	}
}

func TestBufferedResponse_WriteHeaderOnce(t *testing.T) {
	buf := NewBufferedResponse(0)
	buf.WriteHeader(http.StatusCreated)
	buf.WriteHeader(http.StatusTeapot)
	if buf.StatusCode() != http.StatusCreated {
		t.Errorf("status after double WriteHeader = %d, want 201", buf.StatusCode())
	}
}

func TestBufferedResponse_ClearBody(t *testing.T) {
	buf := NewBufferedResponse(0)
	_, _ = buf.Write([]byte("data"))
	buf.ClearBody()
	if len(buf.Body()) != 0 {
		t.Errorf("expected empty body after ClearBody, got %d bytes", len(buf.Body()))
	}
	if buf.BytesWritten() != 0 {
		t.Errorf("BytesWritten after ClearBody = %d, want 0", buf.BytesWritten())
	}
}

func TestBufferedResponse_WriteTo(t *testing.T) {
	buf := NewBufferedResponse(0)
	buf.Header().Set("X-Test", "1")
	buf.WriteHeader(http.StatusAccepted)
	_, _ = buf.Write([]byte("payload"))

	dst := httptest.NewRecorder()
	n, err := buf.WriteTo(dst)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if n != 7 {
		t.Errorf("n = %d, want 7", n)
	}
	if dst.Code != http.StatusAccepted {
		t.Errorf("dst status = %d, want 202", dst.Code)
	}
	if dst.Body.String() != "payload" {
		t.Errorf("dst body = %q, want payload", dst.Body.String())
	}
	if dst.Header().Get("X-Test") != "1" {
		t.Error("header X-Test not forwarded")
	}
	if dst.Header().Get(HeaderContentTypeNoSniff) != ContentTypeNoSniffValue {
		t.Error("X-Content-Type-Options not set by WriteTo")
	}
}

func TestBufferedResponse_WriteTo_ReplacesExistingHeader(t *testing.T) {
	buf := NewBufferedResponse(0)
	buf.Header().Set("X-Test", "buffered")
	_, _ = buf.Write([]byte("payload"))

	dst := httptest.NewRecorder()
	dst.Header().Set("X-Test", "stale")

	_, err := buf.WriteTo(dst)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if got := dst.Header().Values("X-Test"); len(got) != 1 || got[0] != "buffered" {
		t.Fatalf("X-Test values = %v, want [buffered]", got)
	}
}

func TestBufferedResponse_WriteTo_PreservesMultiValueHeader(t *testing.T) {
	buf := NewBufferedResponse(0)
	buf.Header().Add("Set-Cookie", "a=1")
	buf.Header().Add("Set-Cookie", "b=2")
	_, _ = buf.Write([]byte("payload"))

	dst := httptest.NewRecorder()
	_, err := buf.WriteTo(dst)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if got := dst.Header().Values("Set-Cookie"); len(got) != 2 || got[0] != "a=1" || got[1] != "b=2" {
		t.Fatalf("Set-Cookie values = %v, want [a=1 b=2]", got)
	}
}

func TestBufferedResponse_WriteTo_NilDst(t *testing.T) {
	buf := NewBufferedResponse(0)
	_, _ = buf.Write([]byte("x"))
	n, err := buf.WriteTo(nil)
	if err != nil || n != 0 {
		t.Errorf("WriteTo(nil) = (%d, %v), want (0, nil)", n, err)
	}
}

func headerValuesContainToken(values []string, want string) bool {
	for _, value := range values {
		for _, token := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(token), want) {
				return true
			}
		}
	}
	return false
}
