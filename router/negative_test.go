package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFreezeBlocksNewRoutes verifies that adding a route after Freeze panics.
func TestFreezeBlocksNewRoutes(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/ping", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
	r.Freeze()

	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("expected panic when adding route to frozen router, got none")
		}
	}()
	mustAddRoute(r, http.MethodGet, "/pong", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
}

// TestDuplicateRoutePanics verifies that registering the same method+path twice panics.
func TestDuplicateRoutePanics(t *testing.T) {
	r := NewRouter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
	mustAddRoute(r, http.MethodGet, "/dup", handler)

	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("expected panic on duplicate route, got none")
		}
	}()
	mustAddRoute(r, http.MethodGet, "/dup", handler)
}

// TestUnknownPathReturns404 verifies that a request for an unregistered path
// returns 404.
func TestUnknownPathReturns404(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/known", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestEmptySegmentInPathReturns404 verifies that double slashes in the request
// path (resulting in empty segments) fail to match a clean route.
func TestEmptySegmentInPathReturns404(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/a/:id/b", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/a//b", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for path with empty segment, got %d", w.Code)
	}
}
