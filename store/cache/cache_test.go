package cache

import (
	"context"
	"errors"
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
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
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
	if ct := resp1.Header().Get("Content-Type"); ct != "text/plain" {
		t.Fatalf("expected Content-Type text/plain, got %q", ct)
	}

	resp2 := httptest.NewRecorder()
	cachedHandler.ServeHTTP(resp2, req)

	if resp2.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 on cache hit, got %d", resp2.Result().StatusCode)
	}
	if resp2.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("expected X-Cache HIT, got %q", resp2.Header().Get("X-Cache"))
	}
	if ct := resp2.Header().Get("Content-Type"); ct != "text/plain" {
		t.Fatalf("expected cached Content-Type text/plain, got %q", ct)
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
