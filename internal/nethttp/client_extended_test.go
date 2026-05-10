package nethttp

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

// ---------------------------------------------------------------------------
// Retry policy unit tests
// ---------------------------------------------------------------------------

func TestCompositeRetryPolicy(t *testing.T) {
	policy := CompositeRetryPolicy{
		Policies: []RetryPolicy{
			TimeoutRetryPolicy{},
			StatusCodeRetryPolicy{Codes: []int{500}},
		},
	}

	if !policy.ShouldRetry(nil, mockTimeoutError{}, 0) {
		t.Error("should retry on timeout error")
	}
	if !policy.ShouldRetry(&http.Response{StatusCode: 500}, nil, 0) {
		t.Error("should retry on 500 status")
	}
	if policy.ShouldRetry(&http.Response{StatusCode: 200}, nil, 0) {
		t.Error("should not retry on success")
	}
}

func TestCompositeRetryPolicySkipsNilPolicy(t *testing.T) {
	policy := CompositeRetryPolicy{
		Policies: []RetryPolicy{
			nil,
			StatusCodeRetryPolicy{Codes: []int{http.StatusTooManyRequests}},
		},
	}

	if !policy.ShouldRetry(&http.Response{StatusCode: http.StatusTooManyRequests}, nil, 0) {
		t.Fatal("expected non-nil policy to decide retry")
	}
	if policy.ShouldRetry(&http.Response{StatusCode: http.StatusOK}, nil, 0) {
		t.Fatal("expected no retry for status not covered by non-nil policies")
	}
}

func TestAlwaysRetryPolicy(t *testing.T) {
	policy := AlwaysRetryPolicy{}

	if !policy.ShouldRetry(nil, errors.New("test error"), 0) {
		t.Error("should retry on any error")
	}
	if !policy.ShouldRetry(&http.Response{StatusCode: 500}, nil, 0) {
		t.Error("should retry on 500 status")
	}
	if policy.ShouldRetry(&http.Response{StatusCode: 200}, nil, 0) {
		t.Error("should not retry on success")
	}
}

// ---------------------------------------------------------------------------
// Client integration tests
// ---------------------------------------------------------------------------

func TestClientRetryExhausted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client := New(
		WithRetryCount(2),
		WithRetryPolicy(StatusCodeRetryPolicy{Codes: []int{500}}),
		WithRetryWait(10*time.Millisecond),
	)

	_, err := client.Get(t.Context(), server.URL)
	if err == nil {
		t.Error("expected error after retries exhausted")
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := New(
		WithTimeout(50*time.Millisecond),
		WithRetryCount(0),
	)

	_, err := client.Get(t.Context(), server.URL)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestClientPostError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	t.Cleanup(server.Close)

	client := New()
	_, err := client.Post(t.Context(), server.URL, []byte("data"), "text/plain")
	if err == nil {
		t.Error("expected error for 400 status")
	}
	if !strings.Contains(err.Error(), "http error") {
		t.Errorf("expected http error, got: %v", err)
	}
}

func TestClientPostJsonMarshalError(t *testing.T) {
	ch := make(chan int) // channels cannot be marshaled to JSON
	client := New()
	_, err := client.PostJson(t.Context(), "http://example.com", ch)
	if err == nil {
		t.Error("expected marshal error")
	}
}

func TestClientRequestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := New()
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // cancel immediately

	_, err := client.Get(ctx, server.URL)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestClientWithTransport(t *testing.T) {
	customTransport := &http.Transport{MaxIdleConns: 100}
	client := New(WithTransport(customTransport))
	if client.client.Transport != customTransport {
		t.Error("custom transport not set")
	}
}

func TestWithRequestOptions(t *testing.T) {
	cfg := &requestConfig{headers: make(map[string]string)}

	timeout := 5 * time.Second
	WithRequestTimeout(timeout)(cfg)
	if cfg.timeout == nil || *cfg.timeout != timeout {
		t.Error("request timeout not set")
	}

	count := 5
	WithRequestRetryCount(count)(cfg)
	if cfg.retryCount == nil || *cfg.retryCount != count {
		t.Error("request retry count not set")
	}

	WithRequestRetryPolicy(AlwaysRetryPolicy{})(cfg)
	if cfg.retryPolicy == nil {
		t.Error("request retry policy not set")
	}

	WithHeader("X-Test", "value")(cfg)
	if cfg.headers["X-Test"] != "value" {
		t.Error("header not set")
	}
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
	t.Cleanup(server.Close)

	client := New(
		WithRetryCount(5),
		WithRetryPolicy(StatusCodeRetryPolicy{Codes: []int{500}}),
		WithRetryWait(10*time.Millisecond),
	)

	resp, err := client.Get(t.Context(), server.URL)
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
	t.Cleanup(server.Close)

	client := New()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
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
	t.Cleanup(server.Close)

	client := New()
	resp, err := client.Get(t.Context(), server.URL,
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

func TestClientPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	t.Cleanup(server.Close)

	client := New()
	resp, err := client.Patch(t.Context(), server.URL, []byte("delta"), "application/json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != "delta" {
		t.Errorf("unexpected body: %s", resp)
	}
}

func TestClientPatchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	t.Cleanup(server.Close)

	client := New()
	_, err := client.Patch(t.Context(), server.URL, []byte("x"), "text/plain")
	if err == nil {
		t.Error("expected error for 422 status")
	}
}
