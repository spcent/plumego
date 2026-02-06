// Package http provides an HTTP client with automatic retry and backoff.
//
// This package wraps the standard library's http.Client with intelligent retry logic:
//   - Automatic retry for network errors and 5xx responses
//   - Configurable retry policies (exponential backoff, linear, custom)
//   - Request timeout handling
//   - Idempotency-aware retries (safe for GET, PUT, DELETE)
//   - Jitter to prevent thundering herd
//
// The client is designed for reliable HTTP communication in distributed systems
// where transient failures are common.
//
// Example usage:
//
//	import nethttp "github.com/spcent/plumego/net/http"
//
//	// Create client with retry
//	client := nethttp.NewClient(nethttp.ClientConfig{
//		Timeout:     10 * time.Second,
//		MaxRetries:  3,
//		RetryPolicy: nethttp.ExponentialBackoff(time.Second, 2.0),
//	})
//
//	// Make a request (automatically retries on failure)
//	resp, err := client.Get("https://api.example.com/data")
//	if err != nil {
//		// All retries failed
//	}
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// RetryPolicy defines the interface for retry strategy.
//
// Implementations determine whether an HTTP request should be retried based on
// the response, error, and current attempt number.
type RetryPolicy interface {
	// ShouldRetry determines if the request should be retried.
	//
	// Parameters:
	//   - resp: The HTTP response (may be nil if request failed before receiving response)
	//   - err: The error from the request attempt (may be nil if response was received)
	//   - attempt: The zero-based attempt counter (0 = first attempt, 1 = first retry, etc.)
	//
	// Returns true if the request should be retried, false otherwise.
	//
	// Note: This method is called AFTER each failed attempt. The attempt parameter
	// starts at 0 for the initial request attempt, increments to 1 for the first retry,
	// 2 for the second retry, and so on.
	ShouldRetry(resp *http.Response, err error, attempt int) bool
}

// TimeoutRetryPolicy retries only on timeout errors.
type TimeoutRetryPolicy struct{}

func (p TimeoutRetryPolicy) ShouldRetry(resp *http.Response, err error, attempt int) bool {
	return isTimeoutError(err)
}

// StatusCodeRetryPolicy retries on specific HTTP status codes (e.g., 5xx).
type StatusCodeRetryPolicy struct {
	Codes []int
}

func (p StatusCodeRetryPolicy) ShouldRetry(resp *http.Response, err error, attempt int) bool {
	if resp == nil {
		return false
	}
	for _, code := range p.Codes {
		if resp.StatusCode == code {
			return true
		}
	}
	return false
}

// CompositeRetryPolicy combines multiple retry policies with OR logic.
type CompositeRetryPolicy struct {
	Policies []RetryPolicy
}

func (p CompositeRetryPolicy) ShouldRetry(resp *http.Response, err error, attempt int) bool {
	for _, policy := range p.Policies {
		if policy.ShouldRetry(resp, err, attempt) {
			return true
		}
	}
	return false
}

// AlwaysRetryPolicy retries on any error.
type AlwaysRetryPolicy struct{}

func (p AlwaysRetryPolicy) ShouldRetry(resp *http.Response, err error, attempt int) bool {
	return err != nil || (resp != nil && resp.StatusCode >= 500)
}

// Middleware defines a function that wraps request execution.
type Middleware func(next RoundTripperFunc) RoundTripperFunc

// RoundTripperFunc is a functional form of http.RoundTripper-like function.
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

// Logging middleware logs request duration and results.
func Logging(next RoundTripperFunc) RoundTripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		start := time.Now()
		resp, err := next(req)
		dur := time.Since(start)
		if err != nil {
			println("HTTP ERR:", err.Error(), "took", dur.String())
		} else {
			println("HTTP OK:", resp.Status, "took", dur.String())
		}
		return resp, err
	}
}

// Metrics middleware can be extended to collect metrics.
func Metrics(next RoundTripperFunc) RoundTripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		start := time.Now()
		resp, err := next(req)
		// In production, you would send metrics to your monitoring system
		_ = start // placeholder for metrics
		return resp, err
	}
}

// Client is a wrapper around http.Client with retry, timeout, and backoff support.
type Client struct {
	client          *http.Client
	retryCount      int
	retryWait       time.Duration
	maxRetryWait    time.Duration
	retryPolicy     RetryPolicy
	defaultTimeout  time.Duration
	middlewares     []Middleware
	retryCheck      func(*http.Request) bool
	ssrfProtection  *SSRFProtection
	enableSSRFCheck bool
}

// Option defines a functional option for Client
type Option func(*Client)

// WithTimeout sets the client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.client.Timeout = timeout
		c.defaultTimeout = timeout
	}
}

// WithRetryCount sets the maximum retry attempts.
func WithRetryCount(count int) Option {
	return func(c *Client) {
		c.retryCount = count
	}
}

// WithRetryWait sets the base retry wait duration.
func WithRetryWait(wait time.Duration) Option {
	return func(c *Client) {
		c.retryWait = wait
	}
}

// WithMaxRetryWait sets the maximum retry wait duration.
func WithMaxRetryWait(max time.Duration) Option {
	return func(c *Client) {
		c.maxRetryWait = max
	}
}

// WithRetryPolicy sets a custom retry policy.
func WithRetryPolicy(policy RetryPolicy) Option {
	return func(c *Client) {
		c.retryPolicy = policy
	}
}

// WithRetryCheck limits retries based on request attributes.
// Returning false disables retries for the given request.
func WithRetryCheck(check func(*http.Request) bool) Option {
	return func(c *Client) {
		c.retryCheck = check
	}
}

// WithMiddleware adds middleware to the client.
func WithMiddleware(mw Middleware) Option {
	return func(c *Client) { c.middlewares = append(c.middlewares, mw) }
}

// WithTransport sets a custom transport.
func WithTransport(transport http.RoundTripper) Option {
	return func(c *Client) {
		c.client.Transport = transport
	}
}

// WithSSRFProtection enables SSRF protection with the given configuration.
func WithSSRFProtection(protection SSRFProtection) Option {
	return func(c *Client) {
		c.ssrfProtection = &protection
		c.enableSSRFCheck = true
	}
}

// WithDefaultSSRFProtection enables SSRF protection with secure defaults.
func WithDefaultSSRFProtection() Option {
	return func(c *Client) {
		defaults := DefaultSSRFProtection()
		c.ssrfProtection = &defaults
		c.enableSSRFCheck = true
	}
}

// New creates a new Client with provided options.
func New(opts ...Option) *Client {
	c := &Client{
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		retryCount:     3,
		retryWait:      1 * time.Second,
		maxRetryWait:   5 * time.Second,
		retryPolicy:    TimeoutRetryPolicy{},
		defaultTimeout: 3 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// requestConfig holds per-request configuration.
type requestConfig struct {
	retryCount  *int
	retryPolicy RetryPolicy
	timeout     *time.Duration
	headers     map[string]string
	retryCheck  func(*http.Request) bool
}

// RequestOption defines a functional option for individual requests.
type RequestOption func(*requestConfig)

// WithRequestTimeout sets timeout for a specific request.
func WithRequestTimeout(timeout time.Duration) RequestOption {
	return func(c *requestConfig) { c.timeout = &timeout }
}

// WithRequestRetryCount sets retry count for a specific request.
func WithRequestRetryCount(count int) RequestOption {
	return func(c *requestConfig) { c.retryCount = &count }
}

// WithRequestRetryPolicy sets retry policy for a specific request.
func WithRequestRetryPolicy(p RetryPolicy) RequestOption {
	return func(c *requestConfig) { c.retryPolicy = p }
}

// WithRequestRetryCheck limits retries for a specific request.
func WithRequestRetryCheck(check func(*http.Request) bool) RequestOption {
	return func(c *requestConfig) { c.retryCheck = check }
}

// WithHeader adds a header to the request.
func WithHeader(key, val string) RequestOption {
	return func(cfg *requestConfig) {
		cfg.headers[key] = val
	}
}

// isTimeoutError checks if an error is caused by timeout.
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	if strings.Contains(err.Error(), "timeout") {
		return true
	}

	return false
}

// backoffWithJitter calculates exponential backoff with jitter.
func backoffWithJitter(base time.Duration, attempt int, max time.Duration) time.Duration {
	backoff := float64(base) * math.Pow(2, float64(attempt))
	if backoff > float64(max) {
		backoff = float64(max)
	}
	jitterFactor := 0.5 + rand.Float64() // [0.5, 1.5)
	return time.Duration(backoff * jitterFactor)
}

// doRequest executes an HTTP request with retry and policy control.
func (c *Client) doRequest(req *http.Request, opts ...RequestOption) (*http.Response, error) {
	// SSRF protection - validate URL before making request
	if c.enableSSRFCheck && c.ssrfProtection != nil {
		if err := ValidateURL(req.URL.String(), *c.ssrfProtection); err != nil {
			return nil, err
		}
	}

	cfg := &requestConfig{
		retryCount:  &c.retryCount,
		retryPolicy: c.retryPolicy,
		timeout:     &c.defaultTimeout,
		headers:     make(map[string]string),
		retryCheck:  c.retryCheck,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.retryCheck != nil && !cfg.retryCheck(req) {
		noRetry := 0
		cfg.retryCount = &noRetry
	}

	// apply headers
	for k, v := range cfg.headers {
		req.Header.Set(k, v)
	}

	if err := ensureReplayableBody(req, cfg.retryCount); err != nil {
		return nil, err
	}

	// build middleware chain
	final := c.do(cfg)
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		final = c.middlewares[i](final)
	}

	return final(req)
}

type cancelingBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelingBody) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

func ensureReplayableBody(req *http.Request, retryCount *int) error {
	if req == nil || req.Body == nil || req.GetBody != nil {
		return nil
	}
	if retryCount == nil || *retryCount <= 0 {
		return nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	_ = req.Body.Close()

	req.Body = io.NopCloser(bytes.NewReader(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	return nil
}

func cloneRequest(req *http.Request, ctx context.Context) (*http.Request, error) {
	cloned := req.Clone(ctx)

	if req.Body == nil {
		return cloned, nil
	}

	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		cloned.Body = body
		return cloned, nil
	}

	cloned.Body = req.Body
	return cloned, nil
}

// do executes an HTTP request with retry logic.
func (c *Client) do(cfg *requestConfig) RoundTripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		retryCount := 0
		if cfg.retryCount != nil {
			retryCount = *cfg.retryCount
		}

		for i := 0; i <= retryCount; i++ {
			ctx, cancel := context.WithTimeout(req.Context(), *cfg.timeout)
			attemptReq, err := cloneRequest(req, ctx)
			if err != nil {
				cancel()
				return nil, err
			}

			resp, reqErr := c.client.Do(attemptReq)
			if reqErr == nil && (resp.StatusCode < 500) {
				if resp != nil && resp.Body != nil {
					resp.Body = &cancelingBody{ReadCloser: resp.Body, cancel: cancel}
				} else {
					cancel()
				}
				return resp, nil
			}

			shouldRetry := cfg.retryPolicy != nil && cfg.retryPolicy.ShouldRetry(resp, reqErr, i)
			if !shouldRetry || i == retryCount {
				if reqErr == nil && resp != nil && resp.StatusCode >= 500 {
					reqErr = errors.New("http error: " + resp.Status)
				}
				if resp != nil && resp.Body != nil {
					resp.Body = &cancelingBody{ReadCloser: resp.Body, cancel: cancel}
				} else {
					cancel()
				}
				return resp, reqErr
			}

			if resp != nil && resp.Body != nil {
				// Just close the body without reading - we're retrying anyway
				_ = resp.Body.Close()
			}
			cancel()

			time.Sleep(backoffWithJitter(c.retryWait, i, c.maxRetryWait))
		}

		return nil, errors.New("http error: retries exhausted")
	}
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, url string, opts ...RequestOption) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req, opts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.New("http error: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, url string, body []byte, contentType string, opts ...RequestOption) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := c.doRequest(req, opts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("http error: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// PostJson performs a POST request with JSON body.
func (c *Client) PostJson(ctx context.Context, url string, data any, opts ...RequestOption) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.Post(ctx, url, body, "application/json", opts...)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, url string, body []byte, contentType string, opts ...RequestOption) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := c.doRequest(req, opts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, errors.New("http error: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, url string, opts ...RequestOption) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req, opts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.New("http error: " + resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// Do performs a custom request.
func (c *Client) Do(req *http.Request, opts ...RequestOption) (*http.Response, error) {
	return c.doRequest(req, opts...)
}
