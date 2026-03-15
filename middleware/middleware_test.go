package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// trace returns a Middleware that appends "name:before" before calling next and
// "name:after" after it returns. Tests use the collected log to assert the exact
// two-way execution order of a middleware pipeline.
func trace(name string, log *[]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*log = append(*log, name+":before")
			next.ServeHTTP(w, r)
			*log = append(*log, name+":after")
		})
	}
}

func assertOrder(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("execution order mismatch\n  got:  %v\n  want: %v", got, want)
	}
}

// TestChainRegistrationOrder verifies that the first registered middleware is
// outermost: its "before" fires first on the request path, its "after" fires
// last on the response path.
func TestChainRegistrationOrder(t *testing.T) {
	var log []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
		w.WriteHeader(http.StatusOK)
	})

	NewChain(trace("mw1", &log), trace("mw2", &log), trace("mw3", &log)).
		Apply(handler).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	assertOrder(t, log, []string{
		"mw1:before", "mw2:before", "mw3:before",
		"handler",
		"mw3:after", "mw2:after", "mw1:after",
	})
}

// TestChainUseAppends verifies that Use adds to the end of the chain, so
// later-added middlewares wrap closer to the handler.
func TestChainUseAppends(t *testing.T) {
	var log []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
	})

	chain := NewChain(trace("a", &log))
	chain.Use(trace("b", &log))
	chain.Use(trace("c", &log))
	chain.Apply(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	// a is outermost (registered first), c is innermost (registered last).
	assertOrder(t, log, []string{
		"a:before", "b:before", "c:before",
		"handler",
		"c:after", "b:after", "a:after",
	})
}

// TestChainUseVariadic verifies that Use accepts multiple middlewares at once,
// registering them in the order provided.
func TestChainUseVariadic(t *testing.T) {
	var log []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
	})

	NewChain().
		Use(trace("a", &log), trace("b", &log), trace("c", &log)).
		Apply(handler).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	assertOrder(t, log, []string{
		"a:before", "b:before", "c:before",
		"handler",
		"c:after", "b:after", "a:after",
	})
}

// TestChainPrependRunsFirst verifies that Prepend inserts at the front of the
// chain, so prepended middlewares execute before any Use middlewares.
func TestChainPrependRunsFirst(t *testing.T) {
	var log []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
	})

	chain := NewChain(trace("b", &log), trace("c", &log))
	chain.Prepend(trace("a", &log))
	chain.Apply(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	assertOrder(t, log, []string{
		"a:before", "b:before", "c:before",
		"handler",
		"c:after", "b:after", "a:after",
	})
}

// TestChainPrependMultiple verifies that multiple middlewares prepended at once
// keep their relative order at the front.
func TestChainPrependMultiple(t *testing.T) {
	var log []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
	})

	chain := NewChain(trace("c", &log))
	chain.Prepend(trace("a", &log), trace("b", &log))
	chain.Apply(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	assertOrder(t, log, []string{
		"a:before", "b:before", "c:before",
		"handler",
		"c:after", "b:after", "a:after",
	})
}

// TestChainApplyIsImmutable verifies that a handler built by Apply is not
// affected by subsequent Use or Prepend calls on the same Chain.
func TestChainApplyIsImmutable(t *testing.T) {
	var log []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
	})

	chain := NewChain(trace("a", &log))
	built := chain.Apply(handler) // snapshot at this point

	chain.Use(trace("b", &log))       // must not affect built
	chain.Prepend(trace("pre", &log)) // must not affect built

	log = nil
	built.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	// Only "a" was registered before Apply; "b" and "pre" must not appear.
	assertOrder(t, log, []string{"a:before", "handler", "a:after"})
}

// TestChainShortCircuit verifies that a middleware that does not call next
// stops execution of all subsequent middlewares and the handler.
func TestChainShortCircuit(t *testing.T) {
	var log []string

	gate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log = append(log, "gate:reject")
			w.WriteHeader(http.StatusTeapot)
			// does not call next
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	NewChain(gate, trace("never", &log)).
		Apply(handler).
		ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	if rr.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rr.Code)
	}
	assertOrder(t, log, []string{"gate:reject"})
}

// TestChainLen verifies Len returns the correct count after Use and Prepend.
func TestChainLen(t *testing.T) {
	noop := func(next http.Handler) http.Handler { return next }

	chain := NewChain()
	if chain.Len() != 0 {
		t.Fatalf("expected 0, got %d", chain.Len())
	}

	chain.Use(noop, noop, noop)
	if chain.Len() != 3 {
		t.Fatalf("expected 3, got %d", chain.Len())
	}

	chain.Prepend(noop)
	if chain.Len() != 4 {
		t.Fatalf("expected 4, got %d", chain.Len())
	}
}

// TestChainSnapshot verifies that Snapshot returns an independent copy:
// mutations to the snapshot do not affect the chain, and chain growth after
// the snapshot does not change the snapshot length.
func TestChainSnapshot(t *testing.T) {
	noop := func(next http.Handler) http.Handler { return next }

	chain := NewChain(noop, noop)
	snap := chain.Snapshot()

	if len(snap) != 2 {
		t.Fatalf("expected snapshot length 2, got %d", len(snap))
	}

	snap[0] = nil // mutate the snapshot
	if chain.Snapshot()[0] == nil {
		t.Fatal("snapshot mutation leaked into the chain")
	}

	chain.Use(noop) // grow the chain after snapshot
	if len(snap) != 2 {
		t.Fatalf("chain growth changed snapshot length to %d", len(snap))
	}
}

// TestApply verifies the package-level Apply function applies middlewares in
// the order provided (first = outermost).
func TestApply(t *testing.T) {
	var log []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
		w.WriteHeader(http.StatusOK)
	})

	Apply(handler, trace("mw1", &log), trace("mw2", &log)).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	assertOrder(t, log, []string{
		"mw1:before", "mw2:before",
		"handler",
		"mw2:after", "mw1:after",
	})
}

// TestApplyFunc verifies the package-level ApplyFunc function.
func TestApplyFunc(t *testing.T) {
	var log []string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
		w.WriteHeader(http.StatusOK)
	})

	ApplyFunc(handler, trace("mw1", &log), trace("mw2", &log)).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	assertOrder(t, log, []string{
		"mw1:before", "mw2:before",
		"handler",
		"mw2:after", "mw1:after",
	})
}

// TestFromFuncMiddleware verifies the adapter for func-style middleware.
func TestFromFuncMiddleware(t *testing.T) {
	var log []string

	fm := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log = append(log, "func-mw:before")
			next(w, r)
			log = append(log, "func-mw:after")
		}
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log = append(log, "handler")
		w.WriteHeader(http.StatusOK)
	})

	FromFuncMiddleware(fm)(handler).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	assertOrder(t, log, []string{"func-mw:before", "handler", "func-mw:after"})
}

// --- Benchmarks ---

func noopMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

var benchHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func BenchmarkChainApply(b *testing.B) {
	for _, n := range []int{1, 5, 10, 20} {
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
	req := httptest.NewRequest("GET", "/", nil)
	for _, n := range []int{1, 5, 10, 20} {
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
