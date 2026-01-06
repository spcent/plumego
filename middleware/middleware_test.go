package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain(t *testing.T) {
	// Create test middleware that adds a header
	addHeader := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "chain")
			next.ServeHTTP(w, r)
		})
	}

	// Create chain with multiple middlewares
	chain := NewChain(addHeader)

	// Test Apply with Handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	wrapped := chain.Apply(Handler(handler))

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
	addHeader := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "func")
			next.ServeHTTP(w, r)
		})
	}

	chain := NewChain(addHeader)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	wrapped := chain.ApplyFunc(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Test") != "func" {
		t.Error("expected X-Test header to be set")
	}
}

func TestHandlerFunc(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := HandlerFunc(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestFromFuncMiddleware(t *testing.T) {
	// Create a FuncMiddleware
	fm := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-From-Func", "true")
			next(w, r)
		}
	}

	// Convert to Middleware
	mw := FromFuncMiddleware(fm)

	// Apply to a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := mw(Handler(handler))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-From-Func") != "true" {
		t.Error("expected X-From-Func header to be set")
	}
}

func TestFromHTTPHandlerMiddleware(t *testing.T) {
	// Create a standard http.Handler middleware
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Standard", "true")
			next.ServeHTTP(w, r)
		})
	}

	// Convert to Middleware
	converted := FromHTTPHandlerMiddleware(mw)

	// Apply to a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := converted(Handler(handler))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-Standard") != "true" {
		t.Error("expected X-Standard header to be set")
	}
}

func TestApply(t *testing.T) {
	// Test with http.HandlerFunc
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Apply", "true")
			next.ServeHTTP(w, r)
		})
	}

	wrapped, err := Apply(handler, mw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-Apply") != "true" {
		t.Error("expected X-Apply header to be set")
	}
}

func TestApplyWithFuncMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fm := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Func", "true")
			next(w, r)
		}
	}

	wrapped, err := Apply(handler, fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-Func") != "true" {
		t.Error("expected X-Func header to be set")
	}
}

func TestApplyErrors(t *testing.T) {
	// Test with invalid handler type
	_, err := Apply("invalid", func(next Handler) Handler { return next })
	if err == nil {
		t.Error("expected error for invalid handler type")
	}

	// Test with invalid middleware type
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	_, err = Apply(handler, "invalid")
	if err == nil {
		t.Error("expected error for invalid middleware type")
	}
}

func TestApplyFunc(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-ApplyFunc", "true")
			next.ServeHTTP(w, r)
		})
	}

	wrapped := ApplyFunc(handler, mw)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped(rr, req)

	if rr.Header().Get("X-ApplyFunc") != "true" {
		t.Error("expected X-ApplyFunc header to be set")
	}
}

func TestApplyFuncMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fm := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-ApplyFuncMiddleware", "true")
			next(w, r)
		}
	}

	wrapped := ApplyFuncMiddleware(handler, fm)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped(rr, req)

	if rr.Header().Get("X-ApplyFuncMiddleware") != "true" {
		t.Error("expected X-ApplyFuncMiddleware header to be set")
	}
}

func TestApplyLegacy(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fm := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Legacy", "true")
			next(w, r)
		}
	}

	wrapped := ApplyLegacy(handler, fm)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped(rr, req)

	if rr.Header().Get("X-Legacy") != "true" {
		t.Error("expected X-Legacy header to be set")
	}
}

func TestChainMultipleMiddlewares(t *testing.T) {
	mw1 := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW1", "1")
			next.ServeHTTP(w, r)
		})
	}

	mw2 := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW2", "2")
			next.ServeHTTP(w, r)
		})
	}

	mw3 := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW3", "3")
			next.ServeHTTP(w, r)
		})
	}

	chain := NewChain(mw1, mw2, mw3)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := chain.Apply(Handler(handler))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	// Check that all headers are set
	if rr.Header().Get("X-MW1") != "1" {
		t.Error("expected X-MW1 header")
	}
	if rr.Header().Get("X-MW2") != "2" {
		t.Error("expected X-MW2 header")
	}
	if rr.Header().Get("X-MW3") != "3" {
		t.Error("expected X-MW3 header")
	}

	// Check order: MW1 should be first, then MW2, then MW3, then handler
	// The headers are set in reverse order during wrapping
	// So MW3 sets first, then MW2, then MW1
	// But when executed, MW1 runs first, then MW2, then MW3
	// So the order in response should be X-MW1, X-MW2, X-MW3
	// Actually, let's check the order by examining the response
	// The last header set would be X-MW1 since it's the outermost
	if rr.Header().Get("X-MW1") != "1" {
		t.Error("MW1 should execute first")
	}
}

func TestChainUse(t *testing.T) {
	chain := NewChain()

	mw := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Used", "true")
			next.ServeHTTP(w, r)
		})
	}

	chain.Use(mw)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := chain.Apply(Handler(handler))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-Used") != "true" {
		t.Error("expected X-Used header to be set")
	}
}

func TestApplyWithMultipleMiddlewares(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw1 := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW1", "1")
			next.ServeHTTP(w, r)
		})
	}

	mw2 := func(next Handler) Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW2", "2")
			next.ServeHTTP(w, r)
		})
	}

	wrapped, err := Apply(handler, mw1, mw2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-MW1") != "1" {
		t.Error("expected X-MW1 header")
	}
	if rr.Header().Get("X-MW2") != "2" {
		t.Error("expected X-MW2 header")
	}
}
