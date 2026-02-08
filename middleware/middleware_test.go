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

func TestChainOrder(t *testing.T) {
	var order []string

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

	chain := NewChain(mw1, mw2)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	chain.Apply(handler).ServeHTTP(rr, req)

	expected := []string{"mw1", "mw2", "handler"}
	if !equalStringSlice(order, expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}
}

func TestChainShortCircuit(t *testing.T) {
	var order []string

	shortCircuit := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "short")
			w.WriteHeader(http.StatusTeapot)
		})
	}

	downstream := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "downstream")
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	NewChain(shortCircuit, downstream).Apply(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Fatalf("expected status 418, got %d", rr.Code)
	}

	expected := []string{"short"}
	if !equalStringSlice(order, expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
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
	// All nil-receiver methods must not panic.
	reg.Use(func(next http.Handler) http.Handler { return next })
	reg.Prepend(func(next http.Handler) http.Handler { return next })
	if reg.Len() != 0 {
		t.Fatal("nil registry Len should be 0")
	}
	if reg.Middlewares() != nil {
		t.Fatal("nil registry Middlewares should be nil")
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

func TestFromFuncMiddleware(t *testing.T) {
	fm := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Func", "yes")
			next(w, r)
		}
	}

	mw := FromFuncMiddleware(fm)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	mw(handler).ServeHTTP(rr, req)

	if rr.Header().Get("X-Func") != "yes" {
		t.Error("expected X-Func header from FromFuncMiddleware")
	}
}

// --- Benchmarks ---

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
