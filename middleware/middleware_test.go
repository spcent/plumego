package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApply(t *testing.T) {
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

func TestApplyMultipleOrder(t *testing.T) {
	var order []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1")
			next.ServeHTTP(w, r)
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2")
			next.ServeHTTP(w, r)
		})
	}

	wrapped := Apply(handler, mw1, mw2)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	expected := []string{"mw1", "mw2", "handler"}
	if !equalStringSlice(order, expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}
}

func TestApplyWithoutMiddlewareKeepsHandlerSemantics(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	wrapped := Apply(handler)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestChainUseAfterConstruction(t *testing.T) {
	var order []string

	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	chain := NewChain(mw("a"))
	chain.Use(mw("b"))
	chain.Use(mw("c"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	req := httptest.NewRequest("GET", "/", nil)
	chain.Apply(handler).ServeHTTP(httptest.NewRecorder(), req)

	expected := []string{"a", "b", "c", "handler"}
	if !equalStringSlice(order, expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
}

func TestChainLen(t *testing.T) {
	chain := NewChain()
	if chain.Len() != 0 {
		t.Fatalf("expected 0, got %d", chain.Len())
	}

	noop := func(next http.Handler) http.Handler { return next }
	chain.Use(noop).Use(noop).Use(noop)
	if chain.Len() != 3 {
		t.Fatalf("expected 3, got %d", chain.Len())
	}
}

func TestRegistryLen(t *testing.T) {
	reg := NewRegistry()
	if reg.Len() != 0 {
		t.Fatalf("expected 0, got %d", reg.Len())
	}

	noop := func(next http.Handler) http.Handler { return next }
	reg.Use(noop, noop)
	if reg.Len() != 2 {
		t.Fatalf("expected 2, got %d", reg.Len())
	}

	reg.Prepend(noop)
	if reg.Len() != 3 {
		t.Fatalf("expected 3, got %d", reg.Len())
	}
}

func TestRegistryNilSafety(t *testing.T) {
	var reg *Registry
	reg.Use(func(next http.Handler) http.Handler { return next })
	reg.Prepend(func(next http.Handler) http.Handler { return next })
	if reg.Len() != 0 {
		t.Fatal("nil registry Len should be 0")
	}
	if reg.Middlewares() != nil {
		t.Fatal("nil registry Middlewares should be nil")
	}
}

func TestRegistryPrependOrder(t *testing.T) {
	var order []string

	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	reg := NewRegistry()
	reg.Use(mw("b"))
	reg.Use(mw("c"))
	reg.Prepend(mw("a"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	chain := NewChain(reg.Middlewares()...)
	req := httptest.NewRequest("GET", "/", nil)
	chain.Apply(handler).ServeHTTP(httptest.NewRecorder(), req)

	expected := []string{"a", "b", "c", "handler"}
	if !equalStringSlice(order, expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
}

func noopMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

var benchHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func BenchmarkChainApply(b *testing.B) {
	sizes := []int{1, 5, 10, 20}
	for _, n := range sizes {
		mws := make([]Middleware, n)
		for i := range mws {
			mws[i] = noopMiddleware
		}
		b.Run(benchName(n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				NewChain(mws...).Apply(benchHandler)
			}
		})
	}
}

func BenchmarkChainServeHTTP(b *testing.B) {
	sizes := []int{1, 5, 10, 20}
	req := httptest.NewRequest("GET", "/", nil)
	for _, n := range sizes {
		mws := make([]Middleware, n)
		for i := range mws {
			mws[i] = noopMiddleware
		}
		handler := NewChain(mws...).Apply(benchHandler)
		b.Run(benchName(n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				handler.ServeHTTP(httptest.NewRecorder(), req)
			}
		})
	}
}

func BenchmarkRegistryUse(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		reg := NewRegistry()
		for range 10 {
			reg.Use(noopMiddleware)
		}
	}
}

func benchName(n int) string {
	switch n {
	case 1:
		return "1mw"
	case 5:
		return "5mw"
	case 10:
		return "10mw"
	case 20:
		return "20mw"
	default:
		return "Nmw"
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
