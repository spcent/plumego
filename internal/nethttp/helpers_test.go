package nethttp

import "net/http"

// mockTimeoutError is a net.Error that reports Timeout() == true.
// Used by multiple test files to simulate network timeout conditions.
type mockTimeoutError struct{}

func (mockTimeoutError) Error() string   { return "mock timeout" }
func (mockTimeoutError) Timeout() bool   { return true }
func (mockTimeoutError) Temporary() bool { return false }

// mockRoundTripper adapts a plain function into an http.RoundTripper.
type mockRoundTripper func(req *http.Request) (*http.Response, error)

func (fn mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }
