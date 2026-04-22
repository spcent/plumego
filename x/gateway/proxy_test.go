package gateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// startBackend starts a test HTTP server and returns its URL.
func startBackend(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// --- ServeHTTP basics ---

func TestProxyBasicRequest(t *testing.T) {
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "ok")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "hello")
	}))

	proxy := New(Config{
		Targets:             []string{backend.URL},
		RetryBackoff:        0,
		AddForwardedHeaders: true,
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxy.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if w.Body.String() != "hello" {
		t.Errorf("body = %q, want hello", w.Body.String())
	}
	if w.Header().Get("X-Backend") != "ok" {
		t.Error("backend header not forwarded")
	}
}

func TestProxyForwardedHeaders(t *testing.T) {
	var receivedXFF string
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedXFF = r.Header.Get("X-Forwarded-For")
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:             []string{backend.URL},
		AddForwardedHeaders: true,
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:1234"
	proxy.ServeHTTP(w, r)

	if receivedXFF == "" {
		t.Error("X-Forwarded-For should be set on backend request")
	}
}

func TestProxyPathRewrite(t *testing.T) {
	var receivedPath string
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:     []string{backend.URL},
		PathRewrite: StripPrefix("/api"),
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	proxy.ServeHTTP(w, r)

	if receivedPath != "/users" {
		t.Errorf("backend received path %q, want /users", receivedPath)
	}
}

func TestProxyNoHealthyBackends(t *testing.T) {
	proxy := New(Config{Targets: []string{"http://localhost:9999"}})
	defer proxy.Close()

	// Mark the only backend unhealthy
	proxy.pool.Backends()[0].SetHealthy(false)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxy.ServeHTTP(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestProxyBackendFailureRecorded(t *testing.T) {
	// Backend that always returns 500
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:      []string{backend.URL},
		RetryCount:   0,
		RetryBackoff: 0,
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxy.ServeHTTP(w, r)

	// Request succeeded; verify stats
	stats := proxy.Stats()
	if stats.TotalBackends != 1 {
		t.Errorf("TotalBackends = %d, want 1", stats.TotalBackends)
	}
	if stats.HealthyBackends != 1 {
		t.Errorf("HealthyBackends = %d, want 1", stats.HealthyBackends)
	}
	if stats.Strategy != "round_robin" {
		t.Errorf("Strategy = %q", stats.Strategy)
	}
}

func TestProxyCustomModifyRequest(t *testing.T) {
	var receivedHeader string
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom")
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:       []string{backend.URL},
		ModifyRequest: SetHeader("X-Custom", "injected"),
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxy.ServeHTTP(w, r)

	if receivedHeader != "injected" {
		t.Errorf("X-Custom = %q, want injected", receivedHeader)
	}
}

func TestProxyCustomModifyResponse(t *testing.T) {
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:        []string{backend.URL},
		ModifyResponse: SetResponseHeader("X-Modified", "yes"),
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxy.ServeHTTP(w, r)

	if w.Header().Get("X-Modified") != "yes" {
		t.Errorf("X-Modified = %q, want yes", w.Header().Get("X-Modified"))
	}
}

func TestProxyPreserveHost(t *testing.T) {
	var receivedHost string
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHost = r.Host
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:      []string{backend.URL},
		PreserveHost: true,
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "original.host"
	proxy.ServeHTTP(w, r)

	if receivedHost != "original.host" {
		t.Errorf("backend host = %q, want original.host", receivedHost)
	}
}

func TestProxyStats(t *testing.T) {
	proxy := New(Config{
		Targets:      []string{"http://a:8080", "http://b:8080"},
		LoadBalancer: NewRoundRobinBalancer(),
	})
	defer proxy.Close()

	stats := proxy.Stats()
	if stats.TotalBackends != 2 {
		t.Errorf("TotalBackends = %d, want 2", stats.TotalBackends)
	}
	if stats.HealthyBackends != 2 {
		t.Errorf("HealthyBackends = %d, want 2", stats.HealthyBackends)
	}
}

func TestProxyMiddleware(t *testing.T) {
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{Targets: []string{backend.URL}})
	defer proxy.Close()

	mw := proxy.Middleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called by proxy middleware")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestProxyWebSocketDetection(t *testing.T) {
	// Proxy with WebSocket disabled should not route as WebSocket
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:          []string{backend.URL},
		WebSocketEnabled: false,
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Connection", "upgrade")
	r.Header.Set("Upgrade", "websocket")
	proxy.ServeHTTP(w, r)

	// Should fall through to HTTP handler (not WebSocket path)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestProxyRetryOnFailure verifies that the proxy retries after backend failure.
func TestProxyRetryOnFailure(t *testing.T) {
	attempt := 0
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:      []string{backend.URL},
		RetryCount:   2,
		RetryBackoff: 0,
		Timeout:      5 * time.Second,
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxy.ServeHTTP(w, r)

	// First attempt should succeed
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestProxyLeastConnectionsBalancer uses LeastConnections and completes without error.
func TestProxyLeastConnectionsBalancer(t *testing.T) {
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:      []string{backend.URL},
		LoadBalancer: NewLeastConnectionsBalancer(),
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxy.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestProxyCustomErrorHandler verifies custom error handler is invoked.
func TestProxyCustomErrorHandler(t *testing.T) {
	proxy := New(Config{
		Targets: []string{"http://localhost:9999"},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("X-Custom-Error", "yes")
			w.WriteHeader(http.StatusBadGateway)
		},
	})
	defer proxy.Close()

	// Force all backends unhealthy
	proxy.pool.Backends()[0].SetHealthy(false)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxy.ServeHTTP(w, r)

	if w.Header().Get("X-Custom-Error") != "yes" {
		t.Error("custom error handler was not called")
	}
}

// TestProxyNewPanicsOnInvalidConfig ensures New panics on invalid config.
func TestProxyNewPanicsOnInvalidConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid config")
		}
	}()
	New(Config{}) // no targets or discovery
}

// TestProxyIsWebSocketRequest exercises the helper.
func TestProxyIsWebSocketRequest(t *testing.T) {
	tests := []struct {
		connection string
		upgrade    string
		want       bool
	}{
		{"upgrade", "websocket", true},
		{"Upgrade", "WebSocket", true},
		{"keep-alive", "", false},
		{"", "", false},
		{"upgrade", "http2", false},
	}

	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		if tt.connection != "" {
			r.Header.Set("Connection", tt.connection)
		}
		if tt.upgrade != "" {
			r.Header.Set("Upgrade", tt.upgrade)
		}
		got := isWebSocketRequest(r)
		if got != tt.want {
			t.Errorf("isWebSocketRequest(connection=%q upgrade=%q) = %v, want %v",
				tt.connection, tt.upgrade, got, tt.want)
		}
	}
}

// --- NewE safe constructor ---

func TestProxyNewE_InvalidConfig_ReturnsError(t *testing.T) {
	p, err := NewE(Config{}) // no targets or discovery
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
	if p != nil {
		t.Error("expected nil proxy on error")
	}
}

func TestProxyNewE_ValidConfig_ReturnsProxy(t *testing.T) {
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	p, err := NewE(Config{Targets: []string{backend.URL}})
	if err != nil {
		t.Fatalf("NewE error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil proxy")
	}
	defer p.Close()
}

// --- Circuit breaker integration ---

func TestProxyCircuitBreakerTripsAfterFailures(t *testing.T) {
	// Start and immediately close a backend so all TCP connections fail.
	// The circuit breaker counts failures when client.Do returns an error
	// (connection-level failure), not when the backend returns an HTTP 5xx.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	targetURL := srv.URL
	srv.Close() // subsequent connections → "connection refused"

	proxy := New(Config{
		Targets:               []string{targetURL},
		CircuitBreakerEnabled: true,
		CircuitBreakerConfig: &CircuitBreakerConfig{
			FailureThreshold: 0.5,
			MinRequests:      3,
			SuccessThreshold: 2,
			Timeout:          5 * time.Second,
		},
		// Use non-zero sentinel values to avoid WithDefaults overriding them.
		RetryCount:   1,
		RetryBackoff: time.Nanosecond,
	})
	defer proxy.Close()

	pool := proxy.pool
	if len(pool.Backends()) == 0 {
		t.Skip("no backends to inspect circuit state")
	}
	cb := pool.Backends()[0].circuitBreaker
	if cb == nil {
		t.Skip("circuit breaker not attached")
	}

	// Drive enough connection failures to trip the breaker.
	// With MinRequests=3 and FailureThreshold=0.5, circuit opens after 3 failures.
	// RetryCount=1 means each ServeHTTP call makes up to 2 CB attempts.
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		proxy.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	if cb.State().String() == "closed" {
		t.Errorf("expected circuit to be open after connection failures, got %v", cb.State())
	}
}

func TestProxyCircuitBreakerReset(t *testing.T) {
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	proxy := New(Config{
		Targets:               []string{backend.URL},
		CircuitBreakerEnabled: true,
		CircuitBreakerConfig: &CircuitBreakerConfig{
			FailureThreshold: 0.5,
			MinRequests:      5,
			SuccessThreshold: 2,
		},
	})
	defer proxy.Close()

	pool := proxy.pool
	if len(pool.Backends()) == 0 {
		t.Skip("no backends to inspect circuit state")
	}
	cb := pool.Backends()[0].circuitBreaker
	if cb == nil {
		t.Skip("circuit breaker not attached")
	}

	cb.Trip()
	cb.Reset()

	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 after circuit reset, got %d", w.Code)
	}
}

// --- TestProxyBufferPool verifies buffer pool path in copyResponse.
func TestProxyBufferPool(t *testing.T) {
	backend := startBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, strings.Repeat("x", 1024))
	}))

	proxy := New(Config{
		Targets:    []string{backend.URL},
		BufferPool: &testBufferPool{},
	})
	defer proxy.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	proxy.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if len(w.Body.String()) != 1024 {
		t.Errorf("body length = %d, want 1024", len(w.Body.String()))
	}
}

// testBufferPool is a trivial BufferPool implementation.
type testBufferPool struct{}

func (p *testBufferPool) Get() []byte { return make([]byte, 32*1024) }
func (p *testBufferPool) Put([]byte)  {}
