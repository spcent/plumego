package router

import (
	"errors"
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

// TestValidationFailureReturns400 verifies that route param validation errors
// produce a 400 response with a structured body.
func TestValidationFailureReturns400(t *testing.T) {
	r := NewRouter()

	validation := NewRouteValidation()
	validation.AddParam("id", &numericValidator{})
	r.AddValidation("GET", "/orders/:id", validation)

	mustAddRoute(r, http.MethodGet, "/orders/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/orders/not-a-number", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid param, got %d", w.Code)
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

// numericValidator rejects any value that is not numeric.
type numericValidator struct{}

func (numericValidator) Validate(name, value string) error {
	for _, c := range value {
		if c < '0' || c > '9' {
			return errors.New("must be numeric")
		}
	}
	return nil
}
