package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/internal/validator"
)

func TestRouteValidationGroupPrefix(t *testing.T) {
	r := NewRouter()
	api := r.Group("/api")

	validation := NewRouteValidation().AddParam("id", validator.RouteParamPositiveInt)
	api.AddValidation(http.MethodGet, "/users/:id", validation)

	api.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/users/abc", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid param, got %d", rec.Code)
	}
}

func TestRouteValidationWildcard(t *testing.T) {
	r := NewRouter()

	validation := NewRouteValidation().AddParam("path", validator.RouteParamNewLength(5, 100))
	r.AddValidation(http.MethodGet, "/files/*path", validation)

	r.Get("/files/*path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/files/a/b", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for short wildcard param, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/files/abcde", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid wildcard param, got %d", rec.Code)
	}
}

func TestRouteValidationCache(t *testing.T) {
	r := NewRouter()

	validation := NewRouteValidation().AddParam("id", validator.RouteParamPositiveInt)
	r.AddValidation(http.MethodGet, "/users/:id", validation)

	r.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid param, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on cached invalid param, got %d", rec.Code)
	}
}

func TestRouteValidationAnyRoute(t *testing.T) {
	r := NewRouter()

	validation := NewRouteValidation().AddParam("id", validator.RouteParamPositiveInt)
	r.AddValidation(ANY, "/users/:id", validation)

	r.Any("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/users/abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid param on ANY route, got %d", rec.Code)
	}
}

func TestGroupCanServeHTTPDirectly(t *testing.T) {
	r := NewRouter()
	api := r.Group("/api")

	api.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when serving through group router, got %d", rec.Code)
	}
}
