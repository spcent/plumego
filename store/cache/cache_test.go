package cache

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryCacheSetAndGet(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Set(context.Background(), "foo", []byte("bar"), time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	val, err := cache.Get(context.Background(), "foo")
	if err != nil {
		t.Fatalf("expected value, got error %v", err)
	}

	if string(val) != "bar" {
		t.Fatalf("expected 'bar', got %q", val)
	}

	val[0] = 'B'
	val2, err := cache.Get(context.Background(), "foo")
	if err != nil {
		t.Fatalf("expected value, got error %v", err)
	}
	if string(val2) != "bar" {
		t.Fatalf("expected original value to remain unchanged, got %q", val2)
	}
}

func TestMemoryCacheExpiration(t *testing.T) {
	cache := NewMemoryCache()

	if err := cache.Set(context.Background(), "temp", []byte("value"), 10*time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	if _, err := cache.Get(context.Background(), "temp"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	exists, err := cache.Exists(context.Background(), "temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatalf("expected key to be expired")
	}
}

func TestMemoryCacheClear(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set(context.Background(), "a", []byte("1"), 0)
	cache.Set(context.Background(), "b", []byte("2"), 0)

	if err := cache.Clear(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := cache.Get(context.Background(), "a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	exists, _ := cache.Exists(context.Background(), "b")
	if exists {
		t.Fatalf("expected cache to be cleared")
	}
}

func TestCachedMiddlewareCachesSuccessfulResponse(t *testing.T) {
	cache := NewMemoryCache()
	calls := int32(0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "hello"}`))
	})

	cachedHandler := Cached(cache, time.Minute, func(r *http.Request) string {
		return "key"
	})(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	resp1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(resp1, req)

	if resp1.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp1.Result().StatusCode)
	}
	if resp1.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("expected X-Cache MISS, got %q", resp1.Header().Get("X-Cache"))
	}
	if ct := resp1.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	resp2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(resp2, req)

	if resp2.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 on cache hit, got %d", resp2.Result().StatusCode)
	}
	if resp2.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("expected X-Cache HIT, got %q", resp2.Header().Get("X-Cache"))
	}
	if ct := resp2.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected cached Content-Type application/json, got %q", ct)
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected handler to be called once, got %d", calls)
	}
}

func TestCachedMiddlewareDoesNotCacheFailures(t *testing.T) {
	cache := NewMemoryCache()
	calls := int32(0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	})

	cachedHandler := Cached(cache, time.Minute, func(r *http.Request) string {
		return "error-key"
	})(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	resp1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(resp1, req)
	if resp1.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("expected MISS on first failure response")
	}

	resp2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(resp2, req)
	if resp2.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("expected MISS on second failure response")
	}

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected handler to be called for each request, got %d", calls)
	}
}

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "negative max key length",
			config:  Config{MaxKeyLength: -1},
			wantErr: true,
		},
		{
			name:    "negative cleanup interval",
			config:  Config{CleanupInterval: -1 * time.Second},
			wantErr: true,
		},
		{
			name:    "valid config",
			config:  Config{MaxKeyLength: 100, CleanupInterval: 1 * time.Second},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		if err := tc.config.Validate(); (err != nil) != tc.wantErr {
			t.Fatalf("%s: Validate() error = %v", tc.name, err)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxKeyLength != DefaultMaxKeyLength {
		t.Fatalf("expected MaxKeyLength %d, got %d", DefaultMaxKeyLength, config.MaxKeyLength)
	}
	if config.MaxMemoryUsage != DefaultMaxMemoryUsage {
		t.Fatalf("expected MaxMemoryUsage %d, got %d", DefaultMaxMemoryUsage, config.MaxMemoryUsage)
	}
	if config.CleanupInterval != DefaultCleanupInterval {
		t.Fatalf("expected CleanupInterval %v, got %v", DefaultCleanupInterval, config.CleanupInterval)
	}
	if !config.EnableMetrics {
		t.Fatal("expected EnableMetrics to be true")
	}
	if config.DefaultTTL != 10*time.Minute {
		t.Fatalf("expected DefaultTTL 10m, got %v", config.DefaultTTL)
	}
}

func TestMemoryCacheWithConfig(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  1024,
		CleanupInterval: 1 * time.Second,
		EnableMetrics:   true,
		DefaultTTL:      5 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// Test Set with default TTL
	err := cache.Set(context.Background(), "test", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test Get
	val, err := cache.Get(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value" {
		t.Fatalf("expected 'value', got %q", val)
	}
}

func TestMemoryCacheMetrics(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Set some values
	cache.Set(context.Background(), "key1", []byte("value1"), time.Minute)
	cache.Set(context.Background(), "key2", []byte("value2"), time.Minute)

	// Get existing key (hit)
	_, _ = cache.Get(context.Background(), "key1")

	// Get non-existing key (miss)
	_, _ = cache.Get(context.Background(), "nonexistent")

	// Delete a key
	_ = cache.Delete(context.Background(), "key2")

	// Get metrics
	metrics := cache.GetMetrics()

	if metrics.Sets != 2 {
		t.Fatalf("expected 2 sets, got %d", metrics.Sets)
	}
	if metrics.Hits != 1 {
		t.Fatalf("expected 1 hit, got %d", metrics.Hits)
	}
	if metrics.Misses != 1 {
		t.Fatalf("expected 1 miss, got %d", metrics.Misses)
	}
	if metrics.Deletes != 1 {
		t.Fatalf("expected 1 delete, got %d", metrics.Deletes)
	}
	if metrics.CurrentSize != 1 {
		t.Fatalf("expected 1 current size, got %d", metrics.CurrentSize)
	}
}

func TestMemoryCacheMemoryLimit(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  10, // Very small limit
		CleanupInterval: 0,
		EnableMetrics:   false,
		DefaultTTL:      1 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// First set should succeed
	err := cache.Set(context.Background(), "key1", []byte("small"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second set should fail due to memory limit
	err = cache.Set(context.Background(), "key2", []byte("larger value that exceeds limit"), 0)
	if err == nil {
		t.Fatal("expected error when exceeding memory limit")
	}
	if !errors.Is(err, ErrCacheFull) {
		t.Fatalf("expected ErrCacheFull, got %v", err)
	}
}

func TestMemoryCacheKeyValidation(t *testing.T) {
	config := Config{
		MaxKeyLength:    10,
		MaxMemoryUsage:  0,
		CleanupInterval: 0,
		EnableMetrics:   false,
		DefaultTTL:      1 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// Test empty key
	err := cache.Set(context.Background(), "", []byte("value"), 0)
	if err == nil {
		t.Fatal("expected error for empty key")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}

	// Test key too long
	longKey := "this_is_a_very_long_key_that_exceeds_the_limit"
	err = cache.Set(context.Background(), longKey, []byte("value"), 0)
	if err == nil {
		t.Fatal("expected error for key too long")
	}
	if !errors.Is(err, ErrKeyTooLong) {
		t.Fatalf("expected ErrKeyTooLong, got %v", err)
	}
}

func TestMemoryCacheDefaultTTL(t *testing.T) {
	config := DefaultConfig()
	config.DefaultTTL = 50 * time.Millisecond

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// Set with zero TTL should use default TTL
	err := cache.Set(context.Background(), "test", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be available immediately
	val, err := cache.Get(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value" {
		t.Fatalf("expected 'value', got %q", val)
	}

	// Wait for default TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	_, err = cache.Get(context.Background(), "test")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after TTL expiration, got %v", err)
	}
}

func TestMemoryCacheCleanup(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  0,
		CleanupInterval: 50 * time.Millisecond,
		EnableMetrics:   true,
		DefaultTTL:      1 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// Set some items with short TTL
	cache.Set(context.Background(), "exp1", []byte("value1"), 10*time.Millisecond)
	cache.Set(context.Background(), "exp2", []byte("value2"), 10*time.Millisecond)
	cache.Set(context.Background(), "keep", []byte("value3"), 1*time.Minute)

	// Wait for items to expire
	time.Sleep(20 * time.Millisecond)

	// Wait for cleanup to run
	time.Sleep(100 * time.Millisecond)

	// Check metrics
	metrics := cache.GetMetrics()
	if metrics.Expired != 2 {
		t.Fatalf("expected 2 expired items, got %d", metrics.Expired)
	}
	if metrics.CurrentSize != 1 {
		t.Fatalf("expected 1 remaining item, got %d", metrics.CurrentSize)
	}
}

func TestKeyFromRequest(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	key1 := KeyFromRequest(req1)
	expected1 := "GET:/api/users"
	if key1 != expected1 {
		t.Fatalf("expected key %q, got %q", expected1, key1)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/users?page=1&limit=10", nil)
	key2 := KeyFromRequest(req2)
	expected2 := "GET:/api/users?page=1&limit=10"
	if key2 != expected2 {
		t.Fatalf("expected key %q, got %q", expected2, key2)
	}

	req3 := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	key3 := KeyFromRequest(req3)
	expected3 := "POST:/api/users"
	if key3 != expected3 {
		t.Fatalf("expected key %q, got %q", expected3, key3)
	}
}

func TestKeyFromRequestWithHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Content-Type", "application/json")

	key := KeyFromRequestWithHeaders(req, "Authorization", "Content-Type")
	expected := "GET:/api/users:Authorization=Bearer token123:Content-Type=application/json"
	if key != expected {
		t.Fatalf("expected key %q, got %q", expected, key)
	}
}

func TestCachedWithConfig(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	calls := int32(0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "hello"}`))
	})

	config := DefaultConfig()
	config.DefaultTTL = 1 * time.Minute

	cachedHandler := CachedWithConfig(cache, config, func(r *http.Request) string {
		return "config-key"
	})(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// First request - cache miss
	resp1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(resp1, req)

	if resp1.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("expected X-Cache MISS, got %q", resp1.Header().Get("X-Cache"))
	}

	// Second request - cache hit
	resp2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(resp2, req)

	if resp2.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("expected X-Cache HIT, got %q", resp2.Header().Get("X-Cache"))
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected handler to be called once, got %d", calls)
	}
}

func TestCachedWithConfigNon2xxResponse(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	calls := int32(0)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	})

	config := DefaultConfig()
	config.DefaultTTL = 1 * time.Minute

	cachedHandler := CachedWithConfig(cache, config, func(r *http.Request) string {
		return "error-key"
	})(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// First request - cache miss
	resp1 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(resp1, req)

	if resp1.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("expected X-Cache MISS, got %q", resp1.Header().Get("X-Cache"))
	}

	// Second request - should still be cache miss (404 not cached)
	resp2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(resp2, req)

	if resp2.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("expected X-Cache MISS for 404, got %q", resp2.Header().Get("X-Cache"))
	}

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected handler to be called twice, got %d", calls)
	}
}

func TestMemoryCacheClose(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  0,
		CleanupInterval: 50 * time.Millisecond,
		EnableMetrics:   false,
		DefaultTTL:      1 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)

	// Set some data
	cache.Set(context.Background(), "test", []byte("value"), 1*time.Minute)

	// Close the cache
	err := cache.Close()
	if err != nil {
		t.Fatalf("unexpected error closing cache: %v", err)
	}

	// Try to use cache after close (should still work for basic operations)
	// Note: In a real implementation, you might want to prevent usage after close
}

func TestMemoryCacheZeroTTL(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Set with zero TTL (should store indefinitely)
	err := cache.Set(context.Background(), "permanent", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Should still be available
	val, err := cache.Get(context.Background(), "permanent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value" {
		t.Fatalf("expected 'value', got %q", val)
	}
}

func TestMemoryCacheUpdateExistingKey(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Set initial value
	err := cache.Set(context.Background(), "key", []byte("old"), 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Update with new value
	err = cache.Set(context.Background(), "key", []byte("new"), 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get updated value
	val, err := cache.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "new" {
		t.Fatalf("expected 'new', got %q", val)
	}

	// Check metrics
	metrics := cache.GetMetrics()
	if metrics.Sets != 2 {
		t.Fatalf("expected 2 sets, got %d", metrics.Sets)
	}
	if metrics.CurrentSize != 1 {
		t.Fatalf("expected 1 current size, got %d", metrics.CurrentSize)
	}
}

func TestMemoryCacheConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Test concurrent Set operations
	for i := 0; i < 100; i++ {
		go func(i int) {
			key := fmt.Sprintf("key%d", i)
			value := []byte(fmt.Sprintf("value%d", i))
			cache.Set(context.Background(), key, value, 1*time.Minute)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	// Test concurrent Get operations
	for i := 0; i < 100; i++ {
		go func(i int) {
			key := fmt.Sprintf("key%d", i)
			cache.Get(context.Background(), key)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify metrics
	metrics := cache.GetMetrics()
	if metrics.Sets != 100 {
		t.Fatalf("expected 100 sets, got %d", metrics.Sets)
	}
}
