package middleware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestApplyExecutesMiddlewareOnceInOrder(t *testing.T) {
	var steps []string

	handler := func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, "handler")
	}

	mw1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			steps = append(steps, "mw1-before")
			next(w, r)
			steps = append(steps, "mw1-after")
		}
	}

	mw2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			steps = append(steps, "mw2-before")
			next(w, r)
			steps = append(steps, "mw2-after")
		}
	}

	// Use ApplyFuncMiddleware which accepts FuncMiddleware type
	wrapped := ApplyFuncMiddleware(handler, mw1, mw2)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	wrapped(w, req)

	expected := []string{
		"mw1-before",
		"mw2-before",
		"handler",
		"mw2-after",
		"mw1-after",
	}

	if !reflect.DeepEqual(steps, expected) {
		t.Fatalf("unexpected execution order: %v", steps)
	}
}

func TestChainAppliesMiddlewareInOrder(t *testing.T) {
	var steps []string

	handler := func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, "handler")
	}

	// Convert FuncMiddleware to Middleware
	mw1 := FromFuncMiddleware(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			steps = append(steps, "mw1-before")
			next(w, r)
			steps = append(steps, "mw1-after")
		}
	})

	mw2 := FromFuncMiddleware(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			steps = append(steps, "mw2-before")
			next(w, r)
			steps = append(steps, "mw2-after")
		}
	})

	// Create a chain and apply it
	chain := NewChain(mw1, mw2)
	wrapped := chain.ApplyFunc(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	wrapped(w, req)

	expected := []string{
		"mw1-before",
		"mw2-before",
		"handler",
		"mw2-after",
		"mw1-after",
	}

	if !reflect.DeepEqual(steps, expected) {
		t.Fatalf("unexpected execution order: %v", steps)
	}
}
