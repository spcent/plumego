package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCompositeRetryPolicy(t *testing.T) {
	policy := CompositeRetryPolicy{
		Policies: []RetryPolicy{
			TimeoutRetryPolicy{},
			StatusCodeRetryPolicy{Codes: []int{500}},
		},
	}

	// Test timeout error
	err := &timeoutError{}
	if !policy.ShouldRetry(nil, err, 0) {
		t.Error("should retry on timeout error")
	}

	// Test 500 status
	resp := &http.Response{StatusCode: 500}
	if !policy.ShouldRetry(resp, nil, 0) {
		t.Error("should retry on 500 status")
	}

	// Test success
	resp = &http.Response{StatusCode: 200}
	if policy.ShouldRetry(resp, nil, 0) {
		t.Error("should not retry on success")
	}
}

func TestAlwaysRetryPolicy(t *testing.T) {
	policy := AlwaysRetryPolicy{}

	// Test with error
	if !policy.ShouldRetry(nil, errors.New("test error"), 0) {
		t.Error("should retry on any error")
	}

	// Test with 500 status
	resp := &http.Response{StatusCode: 500}
	if !policy.ShouldRetry(resp, nil, 0) {
		t.Error("should retry on 500 status")
	}

	// Test with success
	resp = &http.Response{StatusCode: 200}
	if policy.ShouldRetry(resp, nil, 0) {
		t.Error("should not retry on success")
	}
}

func TestClientRetryExhausted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(
		WithRetryCount(2),
		WithRetryPolicy(StatusCodeRetryPolicy{Codes: []int{500}}),
		WithRetryWait(10*time.Millisecond),
	)

	_, err := client.Get(context.Background(), server.URL)
	if err == nil {
		t.Error("expected error after retries exhausted")
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(
		WithTimeout(50*time.Millisecond),
		WithRetryCount(0),
	)

	_, err := client.Get(context.Background(), server.URL)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestClientPostError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client := New()
	_, err := client.Post(context.Background(), server.URL, []byte("data"), "text/plain")
	if err == nil {
		t.Error("expected error for 400 status")
	}

	if !strings.Contains(err.Error(), "http error") {
		t.Errorf("expected http error, got: %v", err)
	}
}

func TestClientPostJsonMarshalError(t *testing.T) {
	// Create a channel which cannot be marshaled
	ch := make(chan int)

	client := New()
	_, err := client.PostJson(context.Background(), "http://example.com", ch)
	if err == nil {
		t.Error("expected marshal error")
	}
}

func TestClientRequestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, server.URL)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestClientWithTransport(t *testing.T) {
	// Create custom transport
	customTransport := &http.Transport{
		MaxIdleConns: 100,
	}

	client := New(
		WithTransport(customTransport),
	)

	if client.client.Transport != customTransport {
		t.Error("custom transport not set")
	}
}

func TestWithRequestOptions(t *testing.T) {
	// Test request-specific timeout
	cfg := &requestConfig{
		headers: make(map[string]string),
	}

	timeout := 5 * time.Second
	WithRequestTimeout(timeout)(cfg)
	if cfg.timeout == nil || *cfg.timeout != timeout {
		t.Error("request timeout not set")
	}

	// Test request-specific retry count
	count := 5
	WithRequestRetryCount(count)(cfg)
	if cfg.retryCount == nil || *cfg.retryCount != count {
		t.Error("request retry count not set")
	}

	// Test request-specific retry policy
	policy := AlwaysRetryPolicy{}
	WithRequestRetryPolicy(policy)(cfg)
	if cfg.retryPolicy == nil {
		t.Error("request retry policy not set")
	}

	// Test header
	WithHeader("X-Test", "value")(cfg)
	if cfg.headers["X-Test"] != "value" {
		t.Error("header not set")
	}
}

func TestLoggingMiddleware(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create client with logging middleware
	client := New(
		WithMiddleware(Logging),
		WithTransport(http.DefaultTransport),
	)

	// Make request
	_, err := client.Get(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The logging middleware prints to stdout, we can't easily capture it
	// but we can verify the request succeeded
}

func TestMetricsMiddleware(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create client with metrics middleware
	client := New(
		WithMiddleware(Metrics),
		WithTransport(http.DefaultTransport),
	)

	// Make request
	_, err := client.Get(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Metrics middleware is a placeholder, just verify it works
}

func TestClientRetryOn500(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := New(
		WithRetryCount(5),
		WithRetryPolicy(StatusCodeRetryPolicy{Codes: []int{500}}),
		WithRetryWait(10*time.Millisecond),
	)

	resp, err := client.Get(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(resp) != "success" {
		t.Errorf("unexpected response: %s", string(resp))
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClientDo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("custom response"))
	}))
	defer server.Close()

	client := New()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "custom response" {
		t.Errorf("unexpected response: %s", string(body))
	}
}

func TestClientWithRequestOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "test" {
			t.Error("custom header not set")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := New()
	resp, err := client.Get(
		context.Background(),
		server.URL,
		WithHeader("X-Custom", "test"),
		WithRequestTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(resp) != "ok" {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

// Helper types
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
