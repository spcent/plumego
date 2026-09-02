package singleflight

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkCoalescerMiddlewareWithSequential(b *testing.B) {
	c := New(Config{})
	count := atomic.Int32{}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	mw := c.Middleware()
	handler := mw(inner)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkCoalescerMiddlewareWithParallel(b *testing.B) {
	c := New(Config{
		Methods: []string{"GET"},
		Timeout: 5 * time.Second,
	})
	count := atomic.Int32{}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	mw := c.Middleware()
	handler := mw(inner)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		for pb.Next() {
			w.Body.Reset()
			handler.ServeHTTP(w, req)
		}
	})
}

func BenchmarkDefaultKeyFunc(b *testing.B) {
	req := httptest.NewRequest("GET", "http://example.com/api/v1/users/123", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultKeyFunc(req)
	}
}

func BenchmarkKeyFuncWithHeaders(b *testing.B) {
	fn := HeaderAwareKeyFunc([]string{"Accept", "Authorization"})
	req := httptest.NewRequest("GET", "http://example.com/api/v1/users/123", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(req)
	}
}
