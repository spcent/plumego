package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestURLFromNestedGroupNamedRoute(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api/")
	v1 := api.Group("/v1")
	files := v1.Group("/files")

	err := files.AddRoute(http.MethodGet, "/:tenant/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tenant := Param(req, "tenant")
		path := Param(req, "path")
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
	err := v1.AddRoute(http.MethodGet, "/users/:id/files/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), WithRouteName("users.file"))
	if err != nil {
		t.Fatalf("add named route failed: %v", err)
	}

	got := r.URL("users.file", "id", "u-1")
	want := ""
	if got != want {
		t.Fatalf("URL() with missing wildcard = %q, want %q", got, want)
	}

	got = r.URL("users.file", "path", "a/b")
	want = ""
	if got != want {
		t.Fatalf("URL() with missing param = %q, want %q", got, want)
	}
}

func TestURLEmptyParamsReturnEmpty(t *testing.T) {
	r := NewRouter()

	err := r.AddRoute(http.MethodGet, "/users/:id/files/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), WithRouteName("users.file"))
	if err != nil {
		t.Fatalf("add named route failed: %v", err)
	}

	tests := []struct {
		name   string
		params []string
	}{
		{name: "empty segment param", params: []string{"id", "", "path", "a/b"}},
		{name: "empty wildcard param", params: []string{"id", "u-1", "path", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.URL("users.file", tt.params...); got != "" {
				t.Fatalf("URL() = %q, want empty string", got)
			}
		})
	}
}

func TestURLMalformedParamPairsReturnEmpty(t *testing.T) {
	r := NewRouter()

	err := r.AddRoute(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), WithRouteName("users.show"))
	if err != nil {
		t.Fatalf("add named route failed: %v", err)
	}

	if got := r.URL("users.show", "id", "42", "extra"); got != "" {
		t.Fatalf("URL() with unpaired param = %q, want empty string", got)
	}
	if got := r.URL("users.show", "id", "42", "tenant", "acme"); got != "" {
		t.Fatalf("URL() with unknown param = %q, want empty string", got)
	}
	if got := r.URL("users.show", "id", "42", "id", "43"); got != "" {
		t.Fatalf("URL() with duplicate param = %q, want empty string", got)
	}
}

func TestNamedRouteCollisionAcrossGroupsReturnsError(t *testing.T) {
	r := NewRouter()

	v1 := r.Group("/api/v1")
	err := v1.AddRoute(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), WithRouteName("users.show"))
	if err != nil {
		t.Fatalf("v1 add route failed: %v", err)
	}

	v2 := r.Group("/api/v2")
	err = v2.AddRoute(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), WithRouteName("users.show"))
	if err == nil {
		t.Fatal("expected duplicate named route registration to fail")
	}

	got := r.URL("users.show", "id", "42")
	want := "/api/v1/users/42"
	if got != want {
		t.Fatalf("URL() after failed same-name registration = %q, want %q", got, want)
	}

	named := r.NamedRoutes()
	route, ok := named["users.show"]
	if !ok {
		t.Fatalf("expected users.show route to exist")
	}
	if route.Pattern != "/api/v1/users/:id" {
		t.Fatalf("named route pattern = %q, want %q", route.Pattern, "/api/v1/users/:id")
	}
}

func TestNamedRoutesReturnsDeepCopy(t *testing.T) {
	r := NewRouter()
	err := r.AddRoute(http.MethodGet, "/users/:id/files/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), WithRouteName("users.file"))
	if err != nil {
		t.Fatalf("add named route failed: %v", err)
	}

	named := r.NamedRoutes()
	named["users.file"].Pattern = "/mutated"
	named["users.file"].ParamPos["id"] = 99
	delete(named, "users.file")

	next := r.NamedRoutes()
	route, ok := next["users.file"]
	if !ok {
		t.Fatalf("expected users.file route to remain registered")
	}
	if route.Pattern != "/users/:id/files/*path" {
		t.Fatalf("named route pattern = %q, want %q", route.Pattern, "/users/:id/files/*path")
	}
	if route.ParamPos["id"] != 0 {
		t.Fatalf("named route param position = %d, want 0", route.ParamPos["id"])
	}
}

func TestURLMustPanicsForUnknownRoute(t *testing.T) {
	r := NewRouter()

	defer func() {
		if rec := recover(); rec == nil {
			t.Fatalf("expected URLMust to panic for unknown route")
		} else if !strings.Contains(rec.(string), `named route "does.not.exist" not found`) {
			t.Fatalf("panic = %v, want not found reason", rec)
		}
	}()

	_ = r.URLMust("does.not.exist", "id", "1")
}

func TestURLMustPanicsWithParamFailureReason(t *testing.T) {
	r := NewRouter()
	err := r.AddRoute(http.MethodGet, "/users/:id/files/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), WithRouteName("users.file"))
	if err != nil {
		t.Fatalf("add named route failed: %v", err)
	}

	tests := []struct {
		name       string
		params     []string
		wantReason string
	}{
		{name: "missing param", params: []string{"id", "u-1"}, wantReason: `missing required param "path"`},
		{name: "empty param", params: []string{"id", "", "path", "a/b"}, wantReason: `empty required param "id"`},
		{name: "unpaired key", params: []string{"id", "u-1", "path"}, wantReason: `unpaired URL param key "path"`},
		{name: "unknown key", params: []string{"id", "u-1", "path", "a/b", "tenant", "acme"}, wantReason: `unknown URL param key "tenant"`},
		{name: "duplicate key", params: []string{"id", "u-1", "path", "a/b", "id", "u-2"}, wantReason: `duplicate URL param key "id"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				rec := recover()
				if rec == nil {
					t.Fatalf("expected URLMust to panic")
				}
				if !strings.Contains(rec.(string), tt.wantReason) {
					t.Fatalf("panic = %v, want reason containing %q", rec, tt.wantReason)
				}
			}()

			_ = r.URLMust("users.file", tt.params...)
		})
	}
}

func TestGroupRootNamedRouteUsesNormalizedPrefix(t *testing.T) {
	r := NewRouter()
	api := r.Group("/api/")
	v1 := api.Group("/v1/")

	err := v1.AddRoute(http.MethodGet, "", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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

func TestNamedMethodHelpersOnGroups(t *testing.T) {
	r := NewRouter()

	api := r.Group("/api")
	v1 := api.Group("/v1")

	if err := v1.AddRoute(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := Param(req, "id")
		_, _ = w.Write([]byte(id))
	}), WithRouteName("users.show")); err != nil {
		t.Fatal(err)
	}
	if err := v1.AddRoute(http.MethodPost, "/users", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), WithRouteName("users.create")); err != nil {
		t.Fatal(err)
	}

	showURL := r.URL("users.show", "id", "42")
	if showURL != "/api/v1/users/42" {
		t.Fatalf("show URL = %q, want %q", showURL, "/api/v1/users/42")
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, showURL, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "42" {
		t.Fatalf("expected body 42, got %q", rec.Body.String())
	}

	createURL := r.URL("users.create")
	if createURL != "/api/v1/users" {
		t.Fatalf("create URL = %q, want %q", createURL, "/api/v1/users")
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, createURL, nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}
