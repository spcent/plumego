// Package proxy provides reverse proxy middleware for plumego
//
// This package implements a production-grade HTTP reverse proxy with features including:
//   - Load balancing (round-robin, weighted, least-connections, IP-hash)
//   - Health checking (active and passive)
//   - Service discovery integration
//   - Request/response modification
//   - Path rewriting
//   - Connection pooling and reuse
//   - Automatic retries
//   - WebSocket support
//
// Example usage:
//
//	import (
//		"github.com/spcent/plumego/middleware/proxy"
//		"github.com/spcent/plumego/core"
//	)
//
//	app := core.New()
//
//	// Simple static proxy
//	app.Use("/api/*", proxy.New(proxy.Config{
//		Targets: []string{
//			"http://backend-1:8080",
//			"http://backend-2:8080",
//		},
//	}))
//
//	app.Boot()
package proxy

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Proxy is the reverse proxy middleware
type Proxy struct {
	config    *Config
	pool      *BackendPool
	transport *TransportPool
	balancer  LoadBalancer
	health    *HealthChecker

	// For least-connections balancer
	lcBalancer *LeastConnectionsBalancer

	mu sync.RWMutex
}

// New creates a new reverse proxy middleware
func New(config Config) func(http.Handler) http.Handler {
	// Apply defaults
	cfg := config.WithDefaults()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	// Create backend pool
	var pool *BackendPool
	var err error

	if len(cfg.Targets) > 0 {
		pool, err = NewBackendPool(cfg.Targets)
		if err != nil {
			panic(err)
		}
	} else {
		// Service discovery will populate backends dynamically
		pool = &BackendPool{backends: make([]*Backend, 0)}
	}

	// Create transport pool
	transportPool := NewTransportPool(cfg.Transport)

	// Create proxy instance
	proxy := &Proxy{
		config:    cfg,
		pool:      pool,
		transport: transportPool,
		balancer:  cfg.LoadBalancer,
	}

	// Check if using least-connections balancer
	if lcb, ok := cfg.LoadBalancer.(*LeastConnectionsBalancer); ok {
		proxy.lcBalancer = lcb
	}

	// Enable circuit breakers for backends if configured
	if cfg.CircuitBreakerEnabled {
		for _, backend := range pool.Backends() {
			cb := newBackendCircuitBreaker(backend.URL, cfg.CircuitBreakerConfig)
			backend.SetCircuitBreaker(cb)
		}
	}

	// Start health checker if configured
	if cfg.HealthCheck != nil {
		proxy.health = NewHealthChecker(pool, cfg.HealthCheck)
		proxy.health.Start()
	}

	// Start service discovery watcher if configured
	if cfg.Discovery != nil {
		go proxy.watchServiceDiscovery()
	}

	return proxy.middleware
}

// middleware is the actual middleware function
func (p *Proxy) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a WebSocket upgrade request
		if p.config.WebSocketEnabled && isWebSocketRequest(r) {
			p.handleWebSocket(w, r)
			return
		}

		// Handle HTTP request
		p.handleHTTP(w, r)
	})
}

// handleHTTP handles HTTP requests
func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	var lastErr error
	attempt := 0

	// Retry loop
	for attempt <= p.config.RetryCount {
		// Select backend
		backend, err := p.balancer.Next(p.pool)
		if err != nil {
			p.config.ErrorHandler(w, r, err)
			return
		}

		// Track connection for least-connections balancer
		if p.lcBalancer != nil {
			p.lcBalancer.Acquire(backend)
			defer p.lcBalancer.Release(backend)
		}

		// Proxy the request
		err = p.proxyRequest(w, r, backend)

		if err == nil {
			// Success!
			backend.RecordSuccess()
			return
		}

		// Record failure
		backend.RecordFailure(p.config.FailureThreshold)
		lastErr = NewProxyError(backend.URL, err, attempt)

		// Check if we should retry
		if attempt >= p.config.RetryCount {
			break
		}

		// Backoff before retry
		if p.config.RetryBackoff > 0 {
			time.Sleep(p.config.RetryBackoff * time.Duration(attempt+1))
		}

		attempt++
	}

	// All retries failed
	if lastErr != nil {
		p.config.ErrorHandler(w, r, lastErr)
	}
}

// proxyRequest proxies a single HTTP request to the backend
func (p *Proxy) proxyRequest(w http.ResponseWriter, r *http.Request, backend *Backend) error {
	// Create a new request for the backend
	backendReq := p.createBackendRequest(r, backend)

	// Apply request modifications
	if err := p.applyRequestModifications(backendReq); err != nil {
		return err
	}

	// Get transport for this backend
	transport := p.transport.Get(backend.URL)
	client := &http.Client{
		Transport: transport,
		Timeout:   p.config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects - return them to the client
			return http.ErrUseLastResponse
		},
	}

	// Execute request with circuit breaker protection
	var resp *http.Response
	err := backend.Execute(func() error {
		var execErr error
		resp, execErr = client.Do(backendReq)
		return execErr
	})

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Apply response modifications
	if p.config.ModifyResponse != nil {
		if err := p.config.ModifyResponse(resp); err != nil {
			return err
		}
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	if resp.Body != nil {
		_, err = p.copyResponse(w, resp.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

// createBackendRequest creates a new request for the backend
func (p *Proxy) createBackendRequest(r *http.Request, backend *Backend) *http.Request {
	// Clone the request
	backendReq := r.Clone(r.Context())

	// Set the backend URL
	backendReq.URL.Scheme = backend.ParsedURL.Scheme
	backendReq.URL.Host = backend.ParsedURL.Host

	// Apply path rewriting if configured
	if p.config.PathRewrite != nil {
		applyPathRewrite(backendReq, p.config.PathRewrite)
	}

	// Set or preserve host header
	if p.config.PreserveHost {
		backendReq.Host = r.Host
	} else {
		backendReq.Host = backend.ParsedURL.Host
	}

	return backendReq
}

// applyRequestModifications applies request modifications
func (p *Proxy) applyRequestModifications(r *http.Request) error {
	// Add forwarded headers
	if p.config.AddForwardedHeaders {
		if err := AddForwardedHeaders()(r); err != nil {
			return err
		}
	}

	// Remove hop-by-hop headers
	if p.config.RemoveHopByHop {
		if err := RemoveHopByHopHeaders()(r); err != nil {
			return err
		}
	}

	// Apply custom modifications
	if p.config.ModifyRequest != nil {
		if err := p.config.ModifyRequest(r); err != nil {
			return err
		}
	}

	return nil
}

// copyResponse copies the response body
func (p *Proxy) copyResponse(w io.Writer, src io.Reader) (int64, error) {
	if p.config.BufferPool != nil {
		// Use buffer pool
		buf := p.config.BufferPool.Get()
		defer p.config.BufferPool.Put(buf)
		return io.CopyBuffer(w, src, buf)
	}

	// Standard copy
	return io.Copy(w, src)
}

// isWebSocketRequest checks if the request is a WebSocket upgrade request
func isWebSocketRequest(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}

// handleWebSocket handles WebSocket proxying
func (p *Proxy) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	var lastErr error
	attempt := 0

	// Retry loop
	for attempt <= p.config.RetryCount {
		// Select backend
		backend, err := p.balancer.Next(p.pool)
		if err != nil {
			p.config.ErrorHandler(w, r, err)
			return
		}

		// Proxy WebSocket connection
		err = p.handleWebSocketProxy(w, r, backend)

		if err == nil {
			// Success!
			backend.RecordSuccess()
			return
		}

		// Record failure
		backend.RecordFailure(p.config.FailureThreshold)
		lastErr = NewProxyError(backend.URL, err, attempt)

		// Check if we should retry
		if attempt >= p.config.RetryCount {
			break
		}

		// Backoff before retry
		if p.config.RetryBackoff > 0 {
			time.Sleep(p.config.RetryBackoff * time.Duration(attempt+1))
		}

		attempt++
	}

	// All retries failed
	if lastErr != nil {
		p.config.ErrorHandler(w, r, lastErr)
	}
}

// watchServiceDiscovery watches for backend updates from service discovery
func (p *Proxy) watchServiceDiscovery() {
	if p.config.Discovery == nil {
		return
	}

	ctx := context.Background()
	updateChan, err := p.config.Discovery.Watch(ctx, p.config.ServiceName)
	if err != nil {
		return
	}

	for backends := range updateChan {
		if len(backends) > 0 {
			if err := p.pool.UpdateBackends(backends); err == nil {
				// Reset load balancer state on backend changes
				p.balancer.Reset()
			}
		}
	}
}

// Close closes the proxy and cleans up resources
func (p *Proxy) Close() {
	// Stop health checker
	if p.health != nil {
		p.health.Stop()
	}

	// Close transport pool
	p.transport.Close()
}

// Stats returns proxy statistics
func (p *Proxy) Stats() ProxyStats {
	return ProxyStats{
		TotalBackends:   p.pool.Count(),
		HealthyBackends: p.pool.HealthyCount(),
		Strategy:        p.balancer.Name(),
	}
}

// ProxyStats holds proxy statistics
type ProxyStats struct {
	TotalBackends   int
	HealthyBackends int
	Strategy        string
}
