package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteValidationGroupPrefix(t *testing.T) {
	r := NewRouter()
	api := r.Group("/api")

	validation := NewRouteValidation().AddParam("id", PositiveIntValidator)
	api.AddValidation(http.MethodGet, "/users/:id", validation)

	api.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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

	validation := NewRouteValidation().AddParam("path", NewLengthValidator(5, 100))
	r.AddValidation(http.MethodGet, "/files/*path", validation)

	r.GetFunc("/files/*path", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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
