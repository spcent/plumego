package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return false }

func TestIsTimeoutError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"deadline", context.DeadlineExceeded, true},
		{"netErr", timeoutErr{}, true},
		{"message", errors.New("request timeout"), true},
		{"other", errors.New("boom"), false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTimeoutError(tt.err); got != tt.want {
				t.Fatalf("isTimeoutError = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackoffWithJitter(t *testing.T) {
	// base 10ms, attempt 2 => 40ms cap with jitter in [0.5,1.5)
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

	body, err := client.Get(context.Background(), server.URL)
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
	resp, err := client.doRequest(
		req,
		WithHeader("X-Test", "demo"),
	)
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
	if _, err := client.Post(context.Background(), srv.URL, []byte("demo"), "text/plain"); err == nil {
		t.Fatalf("expected http error on non-2xx status")
	}
}

func TestTimeoutRetryPolicy(t *testing.T) {
	client := New(WithRetryCount(1), WithRetryWait(1*time.Millisecond), WithMaxRetryWait(2*time.Millisecond))
	// replace transport to simulate timeout errors
	client.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, timeoutErr{}
	})

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := client.doRequest(req)
	if err == nil || !isTimeoutError(err) {
		t.Fatalf("expected timeout error propagation, got %v", err)
	}
}

// roundTripperFunc adapts a func into an http.RoundTripper-like call
// to be used with the HttpClient internal pipeline.
type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }
