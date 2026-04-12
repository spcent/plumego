package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFreezeBlocksNewRoutes verifies that adding a route after Freeze returns
// an error through the public AddRoute contract.
func TestFreezeBlocksNewRoutes(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/ping", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
	r.Freeze()

	if err := r.AddRoute(http.MethodGet, "/pong", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})); err == nil {
		t.Fatal("expected error when adding route to frozen router, got nil")
	}
}

// TestDuplicateRouteReturnsError verifies that registering the same method+path
// twice returns an error through the public AddRoute contract.
func TestDuplicateRouteReturnsError(t *testing.T) {
	r := NewRouter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
	mustAddRoute(r, http.MethodGet, "/dup", handler)

	if err := r.AddRoute(http.MethodGet, "/dup", handler); err == nil {
		t.Fatal("expected error on duplicate route, got nil")
	}
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
