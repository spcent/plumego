package middleware

import (
	"encoding/json"
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

	wrapped := ApplyFunc(handler, Timeout(20*time.Millisecond))

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

	wrapped := ApplyFunc(handler, Timeout(500*time.Millisecond))

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

	wrapped := ApplyFunc(handler, TimeoutWithConfig(TimeoutConfig{
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
