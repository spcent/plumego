package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain(t *testing.T) {
	// Create test middleware that adds a header
	addHeader := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "chain")
			next.ServeHTTP(w, r)
		})
	}

	// Create chain with multiple middlewares
	chain := NewChain(addHeader)

	// Test Apply with http.Handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	wrapped := chain.Apply(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Test") != "chain" {
		t.Error("expected X-Test header to be set")
	}
}

func TestChainApplyFunc(t *testing.T) {
	addHeader := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "func")
			next.ServeHTTP(w, r)
		})
	}

	chain := NewChain(addHeader)

	// Test ApplyFunc
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}

	wrapped := chain.ApplyFunc(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Test") != "func" {
		t.Error("expected X-Test header to be set")
	}
}

func TestApply(t *testing.T) {
	// Test with http.HandlerFunc
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Apply", "true")
			next.ServeHTTP(w, r)
		})
	}

	wrapped := Apply(handler, mw)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-Apply") != "true" {
		t.Error("expected X-Apply header to be set")
	}
}

func TestApplyMultiple(t *testing.T) {
	// Test with multiple middlewares
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW1", "true")
			next.ServeHTTP(w, r)
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW2", "true")
			next.ServeHTTP(w, r)
		})
	}

	wrapped := Apply(handler, mw1, mw2)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-MW1") != "true" {
		t.Error("expected X-MW1 header to be set")
	}
	if rr.Header().Get("X-MW2") != "true" {
		t.Error("expected X-MW2 header to be set")
	}
}

func TestNewChain(t *testing.T) {
	// Test creating a chain with multiple middlewares
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW1", "true")
			next.ServeHTTP(w, r)
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW2", "true")
			next.ServeHTTP(w, r)
		})
	}

	chain := NewChain(mw1, mw2)

	// Test Apply
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := chain.Apply(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-MW1") != "true" {
		t.Error("expected X-MW1 header to be set")
	}
	if rr.Header().Get("X-MW2") != "true" {
		t.Error("expected X-MW2 header to be set")
	}
}