package conformance_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/middleware/recovery"
)

func TestMiddlewareShortCircuitErrorPathOrder(t *testing.T) {
	order := make([]string, 0, 4)
	handlerCalled := false

	outer := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "outer:before")
			next.ServeHTTP(w, r)
			order = append(order, "outer:after")
		})
	}

	blocker := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "blocker:before")
			middleware.WriteTransportError(
				w,
				r,
				http.StatusTooManyRequests,
				middleware.CodeRateLimited,
				"rate limited",
				contract.CategoryRateLimit,
				nil,
			)
			order = append(order, "blocker:return")
		})
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	middleware.Apply(final, outer, blocker).ServeHTTP(rec, req)

	assertCanonicalErrorEnvelope(t, rec, middleware.CodeRateLimited)
	if handlerCalled {
		t.Fatalf("final handler should not be called on short-circuit error path")
	}

	want := []string{"outer:before", "blocker:before", "blocker:return", "outer:after"}
	if len(order) != len(want) {
		t.Fatalf("order length mismatch: got %v want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("unexpected execution order: got %v want %v", order, want)
		}
	}
}

func TestRecoveryCatchesPanicFromDownstreamMiddlewareOrder(t *testing.T) {
	order := make([]string, 0, 4)
	handlerCalled := false

	outer := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "outer:before")
			next.ServeHTTP(w, r)
			order = append(order, "outer:after")
		})
	}

	panicMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "panic:before")
			panic("boom")
		})
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	middleware.Apply(final, outer, recovery.Recovery(log.NewNoOpLogger()), panicMw).ServeHTTP(rec, req)

	assertCanonicalErrorEnvelope(t, rec, middleware.CodeInternalError)
	if handlerCalled {
		t.Fatalf("final handler should not be called when downstream middleware panics")
	}

	want := []string{"outer:before", "panic:before", "outer:after"}
	if len(order) != len(want) {
		t.Fatalf("order length mismatch: got %v want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("unexpected execution order: got %v want %v", order, want)
		}
	}
}
