package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestURLFromNestedGroupNamedRoute(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api/")
	v1 := api.Group("v1")
	files := v1.Group("/files")

	err := files.AddRouteWithOptions(http.MethodGet, "/:tenant/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tenant, _ := contract.Param(req, "tenant")
		path, _ := contract.Param(req, "path")
		w.Write([]byte(tenant + "|" + path))
	}), WithRouteName("files.show"))
	if err != nil {
		t.Fatalf("add named route failed: %v", err)
	}

	got := r.URL("files.show", "tenant", "acme corp", "path", "reports/2026 q1.pdf")
	want := "/api/v1/files/acme%20corp/reports/2026%20q1.pdf"
	if got != want {
		t.Fatalf("URL() = %q, want %q", got, want)
	}

	req := httptest.NewRequest(http.MethodGet, got, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "acme corp|reports/2026 q1.pdf" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestURLMissingParamsInNestedGroupRoute(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api")
	v1 := api.Group("/v1")
	err := v1.AddRouteWithOptions(http.MethodGet, "/users/:id/files/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), WithRouteName("users.file"))
	if err != nil {
		t.Fatalf("add named route failed: %v", err)
	}

	got := r.URL("users.file", "id", "u-1")
	want := "/api/v1/users/u-1/files/"
	if got != want {
		t.Fatalf("URL() with missing wildcard = %q, want %q", got, want)
	}

	got = r.URL("users.file", "path", "a/b")
	want = "/api/v1/users/:id/files/a/b"
	if got != want {
		t.Fatalf("URL() with missing param = %q, want %q", got, want)
	}
}

func TestNamedRouteCollisionAcrossGroupsLastRegistrationWins(t *testing.T) {
	r := NewRouter()

	v1 := r.Group("/api/v1")
	err := v1.AddRouteWithOptions(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), WithRouteName("users.show"))
	if err != nil {
		t.Fatalf("v1 add route failed: %v", err)
	}

	v2 := r.Group("/api/v2")
	err = v2.AddRouteWithOptions(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), WithRouteName("users.show"))
	if err != nil {
		t.Fatalf("v2 add route failed: %v", err)
	}

	got := r.URL("users.show", "id", "42")
	want := "/api/v2/users/42"
	if got != want {
		t.Fatalf("URL() after same-name override = %q, want %q", got, want)
	}

	named := r.NamedRoutes()
	route, ok := named["users.show"]
	if !ok {
		t.Fatalf("expected users.show route to exist")
	}
	if route.Pattern != "/api/v2/users/:id" {
		t.Fatalf("named route pattern = %q, want %q", route.Pattern, "/api/v2/users/:id")
	}
}

func TestURLMustPanicsForUnknownRoute(t *testing.T) {
	r := NewRouter()

	defer func() {
		if rec := recover(); rec == nil {
			t.Fatalf("expected URLMust to panic for unknown route")
		}
	}()

	_ = r.URLMust("does.not.exist", "id", "1")
}

func TestGroupRootNamedRouteUsesNormalizedPrefix(t *testing.T) {
	r := NewRouter()
	api := r.Group("/api/")
	v1 := api.Group("/v1/")

	err := v1.AddRouteWithOptions(http.MethodGet, "", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), WithRouteName("api.v1.root"))
	if err != nil {
		t.Fatalf("add group root route failed: %v", err)
	}

	if got := r.URL("api.v1.root"); got != "/api/v1" {
		t.Fatalf("URL() = %q, want %q", got, "/api/v1")
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
