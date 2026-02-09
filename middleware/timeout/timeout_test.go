package timeout

import (
	"encoding/json"
	"github.com/spcent/plumego/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
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

	wrapped := middleware.ApplyFunc(handler, Timeout(20*time.Millisecond))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	var resp contract.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if rr.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status %d, got %d", http.StatusGatewayTimeout, rr.Code)
	}

	if resp.Error.Code != "request_timeout" || resp.Error.Category != contract.CategoryServer {
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

	wrapped := middleware.ApplyFunc(handler, Timeout(500*time.Millisecond))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped(rr, req)

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

func TestTimeoutMiddleware_BufferLimit(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("toolarge"))
	}

	wrapped := middleware.ApplyFunc(handler, TimeoutWithConfig(TimeoutConfig{
		Timeout:        500 * time.Millisecond,
		MaxBufferBytes: 4,
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestTimeoutMiddleware_StreamingResponse(t *testing.T) {
	// Test that large responses bypass buffering to avoid memory spikes
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Generate 1MB of data (exceeds 512KB threshold)
		largeData := make([]byte, 1<<20)
		for i := range largeData {
			largeData[i] = byte('A' + (i % 26))
		}
		w.Write(largeData)
	}

	wrapped := middleware.ApplyFunc(handler, TimeoutWithConfig(TimeoutConfig{
		Timeout:            500 * time.Millisecond,
		MaxBufferBytes:     10 << 20,
		StreamingThreshold: 512 << 10, // 512KB
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	// Should return error since we cannot replay bypassed response
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d for bypassed large response, got %d", http.StatusInternalServerError, rr.Code)
	}

	var resp contract.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Message != "response too large for timeout buffering" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func TestTimeoutMiddleware_SmallResponseBuffered(t *testing.T) {
	// Test that small responses are still buffered normally
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "ok")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("small"))
	}

	wrapped := middleware.ApplyFunc(handler, TimeoutWithConfig(TimeoutConfig{
		Timeout:            500 * time.Millisecond,
		MaxBufferBytes:     10 << 20,
		StreamingThreshold: 512 << 10, // 512KB
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped(rr, req)

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

	wrapped := middleware.ApplyFunc(handler, TimeoutWithConfig(TimeoutConfig{
		Timeout:            500 * time.Millisecond,
		MaxBufferBytes:     10 << 20,
		StreamingThreshold: 512 << 10, // 512KB
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	// Should be buffered (not exceed threshold)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.Len() != 512<<10 {
		t.Fatalf("expected body length %d, got %d", 512<<10, rr.Body.Len())
	}
}
