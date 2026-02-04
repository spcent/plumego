package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenBucket_BasicLimiting(t *testing.T) {
	callCount := int32(0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	})

	// Create rate limiter with capacity 2, refill 1 per second
	middleware := TokenBucket(Config{
		Capacity:   2,
		RefillRate: 1,
		KeyFunc:    func(r *http.Request) string { return "test" },
	})

	limited := middleware(handler)

	// First request - should succeed (2 tokens -> 1 token)
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	limited.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected first request to succeed, got %d", w1.Code)
	}

	// Second request - should succeed (1 token -> 0 tokens)
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	limited.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected second request to succeed, got %d", w2.Code)
	}

	// Third request - should be rate limited (0 tokens)
	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	limited.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("Expected third request to be rate limited, got %d", w3.Code)
	}

	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("Expected handler to be called 2 times, got %d", callCount)
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Capacity 1, refill 10 per second (fast for testing)
	middleware := TokenBucket(Config{
		Capacity:   1,
		RefillRate: 10,
		KeyFunc:    func(r *http.Request) string { return "test" },
	})

	limited := middleware(handler)

	// First request - consume token
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	limited.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Error("Expected first request to succeed")
	}

	// Immediate second request - should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	limited.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Error("Expected immediate second request to be rate limited")
	}

	// Wait for refill (10 tokens/sec = 100ms per token)
	time.Sleep(150 * time.Millisecond)

	// Third request - should succeed after refill
	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	limited.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Error("Expected request to succeed after refill")
	}
}

func TestTokenBucket_IPKeyFunc(t *testing.T) {
	callCount := int32(0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	})

	middleware := TokenBucket(Config{
		Capacity:   1,
		RefillRate: 1,
		KeyFunc:    IPKeyFunc,
	})

	limited := middleware(handler)

	// Request from IP 1
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "192.168.1.1")
	w1 := httptest.NewRecorder()
	limited.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Error("Expected request from IP 1 to succeed")
	}

	// Second request from IP 1 - should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.1")
	w2 := httptest.NewRecorder()
	limited.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Error("Expected second request from IP 1 to be rate limited")
	}

	// Request from IP 2 - should succeed (different bucket)
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-Forwarded-For", "192.168.1.2")
	w3 := httptest.NewRecorder()
	limited.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Error("Expected request from IP 2 to succeed")
	}

	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("Expected handler to be called 2 times, got %d", callCount)
	}
}

func TestTokenBucket_RateLimitHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := TokenBucket(Config{
		Capacity:   10,
		RefillRate: 5,
		KeyFunc:    func(r *http.Request) string { return "test" },
	})

	limited := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	limited.ServeHTTP(w, req)

	// Check rate limit headers
	if w.Header().Get("X-RateLimit-Limit") != "10" {
		t.Errorf("Expected X-RateLimit-Limit: 10, got %s", w.Header().Get("X-RateLimit-Limit"))
	}

	if w.Header().Get("X-RateLimit-Remaining") != "9" {
		t.Errorf("Expected X-RateLimit-Remaining: 9, got %s", w.Header().Get("X-RateLimit-Remaining"))
	}

	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("Expected X-RateLimit-Reset header")
	}
}

func TestTokenBucket_RetryAfterHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := TokenBucket(Config{
		Capacity:   1,
		RefillRate: 1,
		KeyFunc:    func(r *http.Request) string { return "test" },
	})

	limited := middleware(handler)

	// Consume token
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	limited.ServeHTTP(w1, req1)

	// Rate limited request
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	limited.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Error("Expected request to be rate limited")
	}

	// Check Retry-After header
	retryAfter := w2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("Expected Retry-After header on rate limited response")
	}
}

func TestTokenBucket_ConcurrentRequests(t *testing.T) {
	callCount := int32(0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	})

	middleware := TokenBucket(Config{
		Capacity:   5,
		RefillRate: 10,
		KeyFunc:    func(r *http.Request) string { return "test" },
	})

	limited := middleware(handler)

	// Launch 10 concurrent requests
	concurrency := 10
	var wg sync.WaitGroup
	var successCount int32
	var limitedCount int32

	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			limited.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				atomic.AddInt32(&successCount, 1)
			} else if w.Code == http.StatusTooManyRequests {
				atomic.AddInt32(&limitedCount, 1)
			}
		}()
	}

	wg.Wait()

	// Should have 5 successful (capacity) and 5 rate limited
	success := atomic.LoadInt32(&successCount)
	rateLimited := atomic.LoadInt32(&limitedCount)

	if success != 5 {
		t.Errorf("Expected 5 successful requests, got %d", success)
	}

	if rateLimited != 5 {
		t.Errorf("Expected 5 rate limited requests, got %d", rateLimited)
	}
}

func TestMemoryStore_Cleanup(t *testing.T) {
	store := NewMemoryStore()

	// Add buckets
	bucket := &Bucket{
		Tokens:     10,
		LastRefill: time.Now().Add(-2 * time.Hour),
		Capacity:   10,
		RefillRate: 1,
	}

	store.UpdateBucket("old1", bucket)
	store.UpdateBucket("old2", bucket)

	// Add recent bucket
	recentBucket := &Bucket{
		Tokens:     10,
		LastRefill: time.Now(),
		Capacity:   10,
		RefillRate: 1,
	}
	store.UpdateBucket("recent", recentBucket)

	if store.Count() != 3 {
		t.Errorf("Expected 3 buckets, got %d", store.Count())
	}

	// Cleanup old buckets (> 1 hour)
	removed := store.Cleanup(1 * time.Hour)

	if removed != 2 {
		t.Errorf("Expected 2 buckets to be removed, got %d", removed)
	}

	if store.Count() != 1 {
		t.Errorf("Expected 1 bucket remaining, got %d", store.Count())
	}
}

func TestUserKeyFunc(t *testing.T) {
	keyFunc := UserKeyFunc("X-User-ID")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "user123")

	key := keyFunc(req)

	if key != "user:user123" {
		t.Errorf("Expected key 'user:user123', got '%s'", key)
	}

	// Test anonymous
	req2 := httptest.NewRequest("GET", "/test", nil)
	key2 := keyFunc(req2)

	if key2 != "anonymous" {
		t.Errorf("Expected key 'anonymous', got '%s'", key2)
	}
}

func TestEndpointKeyFunc(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/users/123", nil)
	key := EndpointKeyFunc(req)

	if key != "endpoint:/api/users/123" {
		t.Errorf("Expected key 'endpoint:/api/users/123', got '%s'", key)
	}
}

func TestCompositeKeyFunc(t *testing.T) {
	keyFunc := CompositeKeyFunc(IPKeyFunc, EndpointKeyFunc)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	key := keyFunc(req)

	// Should generate a hash combining both keys
	if key == "" {
		t.Error("Expected composite key to be non-empty")
	}
}

func TestTokenBucket_Callbacks(t *testing.T) {
	allowCalled := false
	rateLimitCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := TokenBucket(Config{
		Capacity:   1,
		RefillRate: 1,
		KeyFunc:    func(r *http.Request) string { return "test" },
		OnAllow: func(key string, tokensRemaining int64) {
			allowCalled = true
		},
		OnRateLimit: func(key string, retryAfter time.Duration) {
			rateLimitCalled = true
		},
	})

	limited := middleware(handler)

	// First request - should call OnAllow
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	limited.ServeHTTP(w1, req1)

	if !allowCalled {
		t.Error("Expected OnAllow to be called")
	}

	// Second request - should call OnRateLimit
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	limited.ServeHTTP(w2, req2)

	if !rateLimitCalled {
		t.Error("Expected OnRateLimit to be called")
	}
}
