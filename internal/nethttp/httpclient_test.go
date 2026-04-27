package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsTimeoutError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"deadline", context.DeadlineExceeded, true},
		{"netErr", mockTimeoutError{}, true},
		{"message", &wrappedErr{"request timeout"}, true},
		{"other", &wrappedErr{"boom"}, false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTimeoutError(tt.err); got != tt.want {
				t.Fatalf("isTimeoutError = %v, want %v", got, tt.want)
			}
		})
	}
}

// wrappedErr is a minimal error type used to test the string-matching branch of isTimeoutError.
type wrappedErr struct{ msg string }

func (e *wrappedErr) Error() string { return e.msg }

type trackingReadCloser struct {
	closed *atomic.Bool
}

func (b trackingReadCloser) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (b trackingReadCloser) Close() error {
	b.closed.Store(true)
	return nil
}

func TestBackoffWithJitter(t *testing.T) {
	// base=10ms, attempt=2 → raw=40ms; with jitter [0.5,1.5) → [20ms, 60ms)
	dur := backoffWithJitter(10*time.Millisecond, 2, 100*time.Millisecond)
	if dur < 5*time.Millisecond || dur >= 60*time.Millisecond {
		t.Fatalf("unexpected jitter backoff: %s", dur)
	}
}

func TestRetryPolicyComposition(t *testing.T) {
	calls := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, "ok")
	}))
	t.Cleanup(server.Close)

	client := New(
		WithRetryCount(1),
		WithRetryWait(10*time.Millisecond),
		WithMaxRetryWait(15*time.Millisecond),
		WithRetryPolicy(StatusCodeRetryPolicy{Codes: []int{http.StatusInternalServerError}}),
	)

	body, err := client.Get(t.Context(), server.URL)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("unexpected body: %s", string(body))
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 attempts, got %d", calls)
	}
}

func TestMiddlewareOrderAndHeaders(t *testing.T) {
	var order []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Test"); got != "demo" {
			t.Fatalf("missing header from request option, got %q", got)
		}
		io.WriteString(w, strings.Join(order, ","))
	}))
	t.Cleanup(srv.Close)

	mw1 := func(next RoundTripperFunc) RoundTripperFunc {
		return func(req *http.Request) (*http.Response, error) {
			order = append(order, "mw1")
			return next(req)
		}
	}
	mw2 := func(next RoundTripperFunc) RoundTripperFunc {
		return func(req *http.Request) (*http.Response, error) {
			order = append(order, "mw2")
			return next(req)
		}
	}

	client := New(WithMiddleware(mw1), WithMiddleware(mw2))
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	resp, err := client.doRequest(req, WithHeader("X-Test", "demo"))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if got := string(data); got != "mw1,mw2" {
		t.Fatalf("unexpected middleware order: %s", got)
	}
}

func TestPostErrorsPropagate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	client := New()
	if _, err := client.Post(t.Context(), srv.URL, []byte("demo"), "text/plain"); err == nil {
		t.Fatalf("expected http error on non-2xx status")
	}
}

func TestGetErrorsPropagate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client := New()
	if _, err := client.Get(t.Context(), srv.URL); err == nil {
		t.Fatalf("expected http error on non-2xx status")
	}
}

func TestHighLevelRequestClosesResponseBodyOnDoRequestError(t *testing.T) {
	var closed atomic.Bool
	client := New(WithRetryCount(0))
	client.client.Transport = mockRoundTripper(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Status:     "500 Internal Server Error",
			Body:       trackingReadCloser{closed: &closed},
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	if _, err := client.Get(t.Context(), "http://example.com"); err == nil {
		t.Fatal("expected final server error")
	}
	if !closed.Load() {
		t.Fatal("expected high-level request helper to close error response body")
	}
}

func TestTimeoutRetryPolicy(t *testing.T) {
	client := New(
		WithRetryCount(1),
		WithRetryWait(1*time.Millisecond),
		WithMaxRetryWait(2*time.Millisecond),
	)
	client.client.Transport = mockRoundTripper(func(req *http.Request) (*http.Response, error) {
		return nil, mockTimeoutError{}
	})

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := client.doRequest(req)
	if err == nil || !isTimeoutError(err) {
		t.Fatalf("expected timeout error propagation, got %v", err)
	}
}

func TestRetryReplaysBody(t *testing.T) {
	var mu sync.Mutex
	bodies := make([]string, 0, 2)
	calls := int32(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()

		mu.Lock()
		bodies = append(bodies, string(body))
		mu.Unlock()

		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := New(
		WithRetryCount(1),
		WithRetryWait(1*time.Millisecond),
		WithRetryPolicy(StatusCodeRetryPolicy{Codes: []int{http.StatusInternalServerError}}),
	)

	_, err := client.Post(t.Context(), server.URL, []byte("hello"), "text/plain")
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 attempts, got %d", calls)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 2 || bodies[0] != "hello" || bodies[1] != "hello" {
		t.Fatalf("unexpected bodies: %v", bodies)
	}
}

func TestRetryCheckDisablesRetries(t *testing.T) {
	calls := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client := New(
		WithRetryCount(3),
		WithRetryPolicy(StatusCodeRetryPolicy{Codes: []int{http.StatusInternalServerError}}),
		WithRetryCheck(func(req *http.Request) bool {
			return req.Method == http.MethodGet
		}),
	)

	_, err := client.Post(t.Context(), server.URL, []byte("demo"), "text/plain")
	if err == nil {
		t.Fatalf("expected error on server response")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 attempt, got %d", calls)
	}
}
