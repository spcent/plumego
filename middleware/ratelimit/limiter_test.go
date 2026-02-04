package ratelimit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
)

func TestRateLimiterQueueFull(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{
		MaxConcurrent: 1,
		QueueDepth:    2,
		QueueTimeout:  500 * time.Millisecond,
	})
	defer rl.Stop()

	started := make(chan struct{})
	release := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-started:
		default:
			close(started)
		}
		<-release
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	done1 := make(chan int, 1)
	go func() {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		done1 <- rec.Code
	}()

	select {
	case <-started:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for first request to start")
	}

	done2 := make(chan int, 1)
	go func() {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		done2 <- rec.Code
	}()

	waitForQueueDepth(t, rl, 2, 200*time.Millisecond)

	rec3 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec3, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec3.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when queue full, got %d", rec3.Code)
	}
	if code := decodeErrorCode(t, rec3); code != "server_busy" {
		t.Fatalf("expected server_busy, got %q", code)
	}

	close(release)

	waitForStatus(t, done1, 500*time.Millisecond)
	waitForStatus(t, done2, 500*time.Millisecond)

	metrics := rl.GetMetrics()
	if metrics.TotalRequests != 3 {
		t.Fatalf("expected total requests 3, got %d", metrics.TotalRequests)
	}
	if metrics.AcceptedRequests != 2 {
		t.Fatalf("expected accepted requests 2, got %d", metrics.AcceptedRequests)
	}
	if metrics.RejectedRequests != 1 {
		t.Fatalf("expected rejected requests 1, got %d", metrics.RejectedRequests)
	}
	if metrics.TimeoutRequests != 0 {
		t.Fatalf("expected timeout requests 0, got %d", metrics.TimeoutRequests)
	}
}

func TestRateLimiterQueueTimeout(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{
		MaxConcurrent: 1,
		QueueDepth:    2,
		QueueTimeout:  20 * time.Millisecond,
	})
	defer rl.Stop()

	started := make(chan struct{})
	release := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-started:
		default:
			close(started)
		}
		<-release
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	done1 := make(chan int, 1)
	go func() {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		done1 <- rec.Code
	}()

	select {
	case <-started:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for first request to start")
	}

	done2 := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		done2 <- rec
	}()

	rec2 := waitForRecorder(t, done2, 200*time.Millisecond)
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for queue timeout, got %d", rec2.Code)
	}
	if code := decodeErrorCode(t, rec2); code != "server_queue_timeout" {
		t.Fatalf("expected server_queue_timeout, got %q", code)
	}

	close(release)
	waitForStatus(t, done1, 200*time.Millisecond)

	metrics := rl.GetMetrics()
	if metrics.TimeoutRequests != 1 {
		t.Fatalf("expected timeout requests 1, got %d", metrics.TimeoutRequests)
	}
}

func decodeErrorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var payload contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	return payload.Error.Code
}

func waitForQueueDepth(t *testing.T, rl *RateLimiter, want int64, timeout time.Duration) {
	t.Helper()

	deadline := time.After(timeout)
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		if rl.GetMetrics().CurrentQueue == want {
			return
		}

		select {
		case <-deadline:
			t.Fatalf("timeout waiting for queue depth %d", want)
		case <-ticker.C:
		}
	}
}

func waitForStatus(t *testing.T, ch <-chan int, timeout time.Duration) {
	t.Helper()

	select {
	case <-ch:
		return
	case <-time.After(timeout):
		t.Fatal("timeout waiting for request to finish")
	}
}

func waitForRecorder(t *testing.T, ch <-chan *httptest.ResponseRecorder, timeout time.Duration) *httptest.ResponseRecorder {
	t.Helper()

	select {
	case rec := <-ch:
		return rec
	case <-time.After(timeout):
		t.Fatal("timeout waiting for request to finish")
		return nil
	}
}
