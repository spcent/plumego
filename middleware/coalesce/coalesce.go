// Package coalesce provides request coalescing middleware
//
// This package deduplicates identical in-flight requests to reduce backend load.
// When multiple concurrent requests for the same resource are received, only one
// request is forwarded to the backend, and all clients receive the same response.
//
// This is particularly useful for:
//   - High-traffic endpoints
//   - Expensive backend operations
//   - Cache warming scenarios
//   - Reducing thundering herd problems
//
// Example usage:
//
//	import (
//		"github.com/spcent/plumego/middleware/coalesce"
//		"github.com/spcent/plumego/core"
//	)
//
//	app := core.New()
//
//	// Simple coalescing with defaults
//	app.Use(coalesce.Middleware(coalesce.Config{}))
//
//	// Advanced configuration
//	app.Use(coalesce.Middleware(coalesce.Config{
//		KeyFunc: coalesce.DefaultKeyFunc,
//		Methods: []string{"GET", "HEAD"},
//		OnCoalesced: func(key string, count int) {
//			log.Printf("Coalesced %d requests for key %s", count, key)
//		},
//	}))
package coalesce

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	nethttp "github.com/spcent/plumego/net/http"
)

// KeyFunc generates a unique key for request deduplication
type KeyFunc func(r *http.Request) string

// Config holds coalescing middleware configuration
type Config struct {
	// KeyFunc generates cache keys from requests
	// Default: DefaultKeyFunc (Method + URL)
	KeyFunc KeyFunc

	// Methods is the list of HTTP methods to coalesce
	// Only safe, idempotent methods should be coalesced
	// Default: ["GET", "HEAD"]
	Methods []string

	// Timeout is the maximum time to wait for an in-flight request
	// If the upstream request takes longer, subsequent requests will
	// create their own requests instead of waiting
	// Default: 30 seconds
	Timeout time.Duration

	// OnCoalesced is called when requests are coalesced (optional)
	// Parameters: key (request key), count (number of coalesced requests)
	OnCoalesced func(key string, count int)

	// OnError is called when an error occurs during coalescing (optional)
	OnError func(key string, err error)
}

// WithDefaults returns a Config with default values applied
func (c *Config) WithDefaults() *Config {
	config := *c

	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}

	if len(config.Methods) == 0 {
		config.Methods = []string{"GET", "HEAD"}
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &config
}

// Coalescer manages in-flight request deduplication
type Coalescer struct {
	mu       sync.RWMutex
	inFlight map[string]*inFlightRequest
	config   *Config
}

// inFlightRequest represents an in-flight request being processed
type inFlightRequest struct {
	key       string
	response  *capturedResponse
	err       error
	done      chan struct{}
	waiters   int
	startTime time.Time
}

// capturedResponse represents a captured HTTP response
type capturedResponse struct {
	statusCode int
	header     http.Header
	body       []byte
}

// New creates a new Coalescer
func New(config Config) *Coalescer {
	cfg := config.WithDefaults()

	return &Coalescer{
		inFlight: make(map[string]*inFlightRequest),
		config:   cfg,
	}
}

// Middleware creates a coalescing middleware
func Middleware(config Config) func(http.Handler) http.Handler {
	coalescer := New(config)
	return coalescer.Middleware()
}

// Middleware returns the middleware function
func (c *Coalescer) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only coalesce safe methods
			if !c.isSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			// Generate request key
			key := c.config.KeyFunc(r)

			// Check if request is already in-flight (single critical section)
			c.mu.Lock()
			inflight, exists := c.inFlight[key]
			if exists {
				inflight.waiters++
				c.mu.Unlock()

				// Wait for in-flight request to complete
				c.waitForInFlight(w, r, key, inflight)
				return
			}

			// Start new request
			inflight = &inFlightRequest{
				key:       key,
				done:      make(chan struct{}),
				waiters:   0,
				startTime: time.Now(),
			}
			c.inFlight[key] = inflight
			c.mu.Unlock()

			c.executeRequest(w, r, key, inflight, next)
		})
	}
}

// waitForInFlight waits for an in-flight request to complete
func (c *Coalescer) waitForInFlight(w http.ResponseWriter, r *http.Request, key string, inflight *inFlightRequest) {
	// Wait for completion or timeout
	select {
	case <-inflight.done:
		// Request completed - write cached response
		if inflight.err != nil {
			if c.config.OnError != nil {
				c.config.OnError(key, inflight.err)
			}
			http.Error(w, "Upstream request failed", http.StatusBadGateway)
			return
		}

		writeResponse(w, inflight.response)

		// Call hook
		if c.config.OnCoalesced != nil {
			c.config.OnCoalesced(key, inflight.waiters)
		}

	case <-time.After(c.config.Timeout):
		// Timeout - execute own request
		// Don't wait for slow upstream
		c.mu.Lock()
		inflight.waiters--
		c.mu.Unlock()

		http.Error(w, "Upstream request timeout", http.StatusGatewayTimeout)
	}
}

// executeRequest executes a new request and broadcasts to waiters
func (c *Coalescer) executeRequest(w http.ResponseWriter, r *http.Request, key string, inflight *inFlightRequest, next http.Handler) {
	// Create response recorder
	recorder := nethttp.NewResponseRecorder(w)

	// Execute request
	next.ServeHTTP(recorder, r)

	// Capture response
	inflight.response = &capturedResponse{
		statusCode: recorder.StatusCode(),
		header:     recorder.Header().Clone(),
		body:       recorder.Body(),
	}

	// Cleanup and broadcast
	c.mu.Lock()
	delete(c.inFlight, key)
	c.mu.Unlock()

	close(inflight.done)
}

// isSafeMethod checks if the HTTP method is safe to coalesce
func (c *Coalescer) isSafeMethod(method string) bool {
	for _, m := range c.config.Methods {
		if method == m {
			return true
		}
	}
	return false
}

// writeResponse writes a captured response to the client
func writeResponse(w http.ResponseWriter, resp *capturedResponse) {
	// Write headers
	for key, values := range resp.header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Add coalesced indicator
	w.Header().Set("X-Coalesced", "true")

	// Write status code
	w.WriteHeader(resp.statusCode)

	// Write body
	if resp.body != nil {
		w.Write(resp.body)
	}
}

// DefaultKeyFunc generates a key from method and URL
func DefaultKeyFunc(r *http.Request) string {
	h := fnv.New64a()
	h.Write([]byte(r.Method))
	h.Write([]byte("|"))
	h.Write([]byte(r.URL.String()))
	return fmt.Sprintf("%x", h.Sum64())
}

// HeaderAwareKeyFunc generates a key including specific headers
func HeaderAwareKeyFunc(headers []string) KeyFunc {
	return func(r *http.Request) string {
		h := fnv.New64a()

		// Method
		h.Write([]byte(r.Method))
		h.Write([]byte("|"))

		// URL
		h.Write([]byte(r.URL.String()))
		h.Write([]byte("|"))

		// Headers
		for _, header := range headers {
			value := r.Header.Get(header)
			if value != "" {
				h.Write([]byte(header))
				h.Write([]byte(":"))
				h.Write([]byte(value))
				h.Write([]byte("|"))
			}
		}

		return fmt.Sprintf("%x", h.Sum64())
	}
}

// Stats returns coalescing statistics
type Stats struct {
	InFlight int
}

// Stats returns current coalescing statistics
func (c *Coalescer) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		InFlight: len(c.inFlight),
	}
}
