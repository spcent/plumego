package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/utils"
)

// TracingMiddleware adds distributed tracing support
func TracingMiddleware(cfg TracingConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if we should sample this request
			if rand.Float64() > cfg.SampleRate {
				next.ServeHTTP(w, r)
				return
			}

			// Extract or generate trace ID
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = generateTraceID()
			}

			// Extract or generate span ID
			spanID := generateSpanID()

			// Add trace headers to request
			r.Header.Set("X-Trace-ID", traceID)
			r.Header.Set("X-Span-ID", spanID)
			r.Header.Set("X-Parent-Span-ID", r.Header.Get("X-Span-ID"))

			// Add trace context to request context
			ctx := context.WithValue(r.Context(), "trace_id", traceID)
			ctx = context.WithValue(ctx, "span_id", spanID)

			// Record span start time
			startTime := time.Now()

			// Wrap response writer to capture status
			rw := &tracingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(rw, r.WithContext(ctx))

			// Record span duration
			duration := time.Since(startTime)

			// Add trace headers to response
			w.Header().Set("X-Trace-ID", traceID)

			// TODO: Send span to tracing backend (Jaeger/Zipkin/OTLP)
			// For now, we just add trace IDs to headers for downstream services
			_ = duration // Use duration for span export
		})
	}
}

// tracingResponseWriter wraps http.ResponseWriter for tracing
type tracingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *tracingResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// CacheMiddleware implements response caching
func CacheMiddleware(cfg CacheConfig) func(http.Handler) http.Handler {
	cache := newResponseCache(cfg.MaxSize * 1024 * 1024) // Convert MB to bytes

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache configured methods
			if !contains(cfg.OnlyMethods, r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			// Check if path should be excluded
			for _, excludePath := range cfg.ExcludePaths {
				if matchPath(r.URL.Path, excludePath) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Generate cache key
			cacheKey := generateCacheKey(r, cfg.VaryHeaders)

			// Try to get from cache
			if cached, found := cache.Get(cacheKey); found {
				// Cache hit
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", cached.ContentType)
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
				return
			}

			// Cache miss - record response
			w.Header().Set("X-Cache", "MISS")
			recorder := &cacheRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(recorder, r)

			// Store in cache if successful response
			if recorder.statusCode >= 200 && recorder.statusCode < 300 {
				cache.Set(cacheKey, &cachedResponse{
					StatusCode:  recorder.statusCode,
					Body:        recorder.body.Bytes(),
					ContentType: recorder.Header().Get("Content-Type"),
					CachedAt:    time.Now(),
				}, cfg.TTL)
			}
		})
	}
}

// cacheRecorder records response for caching
type cacheRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (cr *cacheRecorder) WriteHeader(statusCode int) {
	cr.statusCode = statusCode
	cr.ResponseWriter.WriteHeader(statusCode)
}

func (cr *cacheRecorder) Write(b []byte) (int, error) {
	cr.body.Write(b)
	return utils.SafeWrite(cr.ResponseWriter, b)
}

// responseCache implements a simple in-memory cache
type responseCache struct {
	mu       sync.RWMutex
	items    map[string]*cacheEntry
	maxSize  int
	currSize int
}

type cacheEntry struct {
	response   *cachedResponse
	expiration time.Time
	size       int
}

type cachedResponse struct {
	StatusCode  int
	Body        []byte
	ContentType string
	CachedAt    time.Time
}

func newResponseCache(maxSize int) *responseCache {
	cache := &responseCache{
		items:   make(map[string]*cacheEntry),
		maxSize: maxSize,
	}
	// Start cleanup goroutine
	go cache.cleanup()
	return cache
}

func (c *responseCache) Get(key string) (*cachedResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.items[key]
	if !found {
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.response, true
}

func (c *responseCache) Set(key string, response *cachedResponse, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	size := len(response.Body)

	// Evict if needed to make room
	for c.currSize+size > c.maxSize && len(c.items) > 0 {
		c.evictOldest()
	}

	c.items[key] = &cacheEntry{
		response:   response,
		expiration: time.Now().Add(ttl),
		size:       size,
	}
	c.currSize += size
}

func (c *responseCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.items {
		if oldestKey == "" || entry.response.CachedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.response.CachedAt
		}
	}

	if oldestKey != "" {
		c.currSize -= c.items[oldestKey].size
		delete(c.items, oldestKey)
	}
}

func (c *responseCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.items {
			if now.After(entry.expiration) {
				c.currSize -= entry.size
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// CanaryMiddleware implements traffic splitting for canary deployments
func CanaryMiddleware(cfg CanaryConfig, serviceName string) func(http.Handler) http.Handler {
	// Find canary rule for this service
	var rule *CanaryRule
	for i := range cfg.Rules {
		if cfg.Rules[i].ServiceName == serviceName {
			rule = &cfg.Rules[i]
			break
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rule == nil {
				// No canary rule for this service
				next.ServeHTTP(w, r)
				return
			}

			// Check header-based routing
			if len(rule.HeaderMatch) > 0 {
				shouldRoute := true
				for key, expectedValue := range rule.HeaderMatch {
					if r.Header.Get(key) != expectedValue {
						shouldRoute = false
						break
					}
				}
				if shouldRoute {
					// Route to canary
					r.Header.Set("X-Canary-Target", "canary")
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check cookie-based routing
			if rule.CookieMatch != "" {
				cookie, err := r.Cookie(rule.CookieMatch)
				if err == nil && cookie.Value == "canary" {
					r.Header.Set("X-Canary-Target", "canary")
					next.ServeHTTP(w, r)
					return
				}
			}

			// Weight-based routing
			if rand.Intn(100) < rule.CanaryWeight {
				r.Header.Set("X-Canary-Target", "canary")
			} else {
				r.Header.Set("X-Canary-Target", "primary")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdvancedRoutingMiddleware implements advanced routing rules
func AdvancedRoutingMiddleware(cfg AdvancedConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sort rules by priority (higher priority first)
			// In production, this should be done once during initialization

			// Check each routing rule
			for _, rule := range cfg.RoutingRules {
				if matchRoutingRule(r, rule) {
					// Add routing info to request
					r.Header.Set("X-Target-Service", rule.TargetService)
					if rule.TargetURL != "" {
						r.Header.Set("X-Target-URL", rule.TargetURL)
					}
					break
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions

func generateTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSpanID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateCacheKey(r *http.Request, varyHeaders []string) string {
	// Include method, path, and query
	key := fmt.Sprintf("%s:%s?%s", r.Method, r.URL.Path, r.URL.RawQuery)

	// Include vary headers in cache key
	for _, header := range varyHeaders {
		value := r.Header.Get(header)
		if value != "" {
			key += fmt.Sprintf(":%s=%s", header, value)
		}
	}

	// Hash the key to keep it shorter
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func matchPath(path, pattern string) bool {
	// Simple wildcard matching
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return path == pattern
}

func matchRoutingRule(r *http.Request, rule RoutingRule) bool {
	// Check path pattern
	if rule.Path != "" && !matchPath(r.URL.Path, rule.Path) {
		return false
	}

	// Check header conditions
	for key, expectedValue := range rule.HeaderMatch {
		if r.Header.Get(key) != expectedValue {
			return false
		}
	}

	// Check query parameter conditions
	for key, expectedValue := range rule.QueryMatch {
		if r.URL.Query().Get(key) != expectedValue {
			return false
		}
	}

	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
