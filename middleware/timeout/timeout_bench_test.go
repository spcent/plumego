package timeout

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkTimeoutOverheadZero(b *testing.B) {
	// Timeout <= 0 is transparent pass-through
	mw := Middleware(Config{Timeout: 0})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
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

func BenchmarkTimeoutOverheadShort(b *testing.B) {
	mw := Middleware(Config{
		Timeout:        5 * time.Second,
		MaxBufferBytes: 1 << 20,
		MaxReplayBytes: 256 << 10,
	})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
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

func BenchmarkTimeoutWithBodyWrite(b *testing.B) {
	payload := []byte("hello world")
	mw := Middleware(Config{
		Timeout:        5 * time.Second,
		MaxBufferBytes: 1 << 20,
		MaxReplayBytes: 256 << 10,
	})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	})
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
