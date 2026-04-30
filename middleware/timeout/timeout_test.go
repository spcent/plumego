package timeout

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/middleware"
)

func TestTimeoutMiddleware_TimesOut(t *testing.T) {
	canceled := make(chan struct{})

	handler := func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			close(canceled)
		}
	}

	wrapped := middleware.Apply(http.HandlerFunc(handler), Timeout(TimeoutConfig{Timeout: 20 * time.Millisecond}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	var resp contract.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if rr.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status %d, got %d", http.StatusGatewayTimeout, rr.Code)
	}

	if resp.Error.Code != contract.CodeTimeout || resp.Error.Category != contract.CategoryTimeout {
		t.Fatalf("unexpected response payload: %+v", resp)
	}

	select {
	case <-canceled:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("handler context was not canceled")
	}
}

func TestTimeoutMiddleware_PassThrough(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "ok")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("done"))
	}

	wrapped := middleware.Apply(http.HandlerFunc(handler), Timeout(TimeoutConfig{Timeout: 500 * time.Millisecond}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	if rr.Body.String() != "done" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}

	if rr.Header().Get("X-Test") != "ok" {
		t.Fatalf("header not propagated")
	}
}

func TestTimeoutMiddleware_DisabledWhenTimeoutNonPositive(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if err := r.Context().Err(); err != nil {
			t.Fatalf("unexpected canceled request context: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("disabled"))
	}

	wrapped := middleware.Apply(http.HandlerFunc(handler), Timeout(TimeoutConfig{}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rr.Code)
	}
	if rr.Body.String() != "disabled" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func TestTimeoutMiddleware_BufferLimit(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("toolarge"))
	}

	wrapped := middleware.Apply(http.HandlerFunc(handler), Timeout(TimeoutConfig{
		Timeout:        500 * time.Millisecond,
		MaxBufferBytes: 4,
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestTimeoutMiddleware_RejectsResponseAboveStreamingThreshold(t *testing.T) {
	writeErr := make(chan error, 1)
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		largeData := make([]byte, 1<<20)
		for i := range largeData {
			largeData[i] = byte('A' + (i % 26))
		}
		_, err := w.Write(largeData)
		writeErr <- err
	}

	wrapped := middleware.Apply(http.HandlerFunc(handler), Timeout(TimeoutConfig{
		Timeout:            500 * time.Millisecond,
		MaxBufferBytes:     10 << 20,
		StreamingThreshold: 512 << 10, // 512KB
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d for oversized response, got %d", http.StatusInternalServerError, rr.Code)
	}

	select {
	case err := <-writeErr:
		if err == nil {
			t.Fatalf("expected oversized write to return an error")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("handler did not report write result")
	}

	var resp contract.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Message != "response exceeded timeout buffering threshold" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
	if resp.Error.Code != contract.CodeInternalError || resp.Error.Category != contract.CategoryServer {
		t.Fatalf("unexpected response payload: %+v", resp)
	}
}

func TestTimeoutMiddleware_SmallResponseBuffered(t *testing.T) {
	// Test that small responses are still buffered normally
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "ok")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("small"))
	}

	wrapped := middleware.Apply(http.HandlerFunc(handler), Timeout(TimeoutConfig{
		Timeout:            500 * time.Millisecond,
		MaxBufferBytes:     10 << 20,
		StreamingThreshold: 512 << 10, // 512KB
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	if rr.Body.String() != "small" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}

	if rr.Header().Get("X-Test") != "ok" {
		t.Fatalf("header not propagated")
	}
}

func TestTimeoutMiddleware_StreamingThreshold(t *testing.T) {
	// Test response exactly at threshold boundary
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// 512KB exactly
		data := make([]byte, 512<<10)
		w.Write(data)
	}

	wrapped := middleware.Apply(http.HandlerFunc(handler), Timeout(TimeoutConfig{
		Timeout:            500 * time.Millisecond,
		MaxBufferBytes:     10 << 20,
		StreamingThreshold: 512 << 10, // 512KB
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Should be buffered (not exceed threshold)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.Len() != 512<<10 {
		t.Fatalf("expected body length %d, got %d", 512<<10, rr.Body.Len())
	}
}

func TestTimeoutMiddleware_TimesOutAfterPartialBufferedWrite(t *testing.T) {
	released := make(chan struct{})
	handlerStarted := make(chan struct{})

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte("partial")); err != nil {
			t.Errorf("unexpected partial write error: %v", err)
		}
		close(handlerStarted)
		select {
		case <-released:
		case <-r.Context().Done():
		}
	}

	wrapped := middleware.Apply(http.HandlerFunc(handler), Timeout(TimeoutConfig{
		Timeout:            20 * time.Millisecond,
		MaxBufferBytes:     1024,
		StreamingThreshold: 1024,
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)
	close(released)

	select {
	case <-handlerStarted:
	default:
		t.Fatalf("handler did not start")
	}

	if rr.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status %d, got %d", http.StatusGatewayTimeout, rr.Code)
	}
	if body := rr.Body.String(); body == "partial" {
		t.Fatalf("partial buffered response was committed")
	}

	var resp contract.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Code != contract.CodeTimeout {
		t.Fatalf("expected timeout code, got %+v", resp)
	}
}
