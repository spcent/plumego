package router

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestFreezeRoutesSnapshotSortedAndIncludesMetadata(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodPost, "/zeta", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	mustAddRoute(r, http.MethodGet, "/alpha/:id", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), WithRouteName("alpha.show"))
	mustAddRoute(r, http.MethodDelete, "/alpha/:id", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	routes := r.Routes()
	want := []RouteInfo{
		{Method: http.MethodDelete, Path: "/alpha/:id"},
		{Method: http.MethodGet, Path: "/alpha/:id", Meta: RouteMeta{Name: "alpha.show"}},
		{Method: http.MethodPost, Path: "/zeta"},
	}
	if len(routes) != len(want) {
		t.Fatalf("route count = %d, want %d: %#v", len(routes), len(want), routes)
	}
	for i := range want {
		if routes[i] != want[i] {
			t.Fatalf("route %d = %#v, want %#v", i, routes[i], want[i])
		}
	}
}

func TestFreezeNestedGroupDispatchContext(t *testing.T) {
	r := NewRouter()
	v1 := r.Group("/api/").Group("v1")
	users := v1.Group("/tenants/:tenant/users")

	if err := users.AddRoute(http.MethodGet, ":id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rc := contract.RequestContextFromContext(req.Context())
		if rc.RoutePattern != "/api/v1/tenants/:tenant/users/:id" {
			t.Fatalf("route pattern = %q", rc.RoutePattern)
		}
		if rc.RouteName != "tenant.users.show" {
			t.Fatalf("route name = %q", rc.RouteName)
		}
		if Param(req, "tenant") != "acme" || Param(req, "id") != "u-1" {
			t.Fatalf("params tenant=%q id=%q", Param(req, "tenant"), Param(req, "id"))
		}
		_, _ = w.Write([]byte("ok"))
	}), WithRouteName("tenant.users.show")); err != nil {
		t.Fatalf("add route: %v", err)
	}

	rec := serveRouter(r, http.MethodGet, "/api/v1/tenants/acme/users/u-1")
	assertResponseStatus(t, rec, http.StatusOK)
	assertResponseBody(t, rec, "ok")
	if got := r.URL("tenant.users.show", "tenant", "acme", "id", "u-1"); got != "/api/v1/tenants/acme/users/u-1" {
		t.Fatalf("URL() = %q", got)
	}
}

func TestFreezeMethodNotAllowedStructuredError(t *testing.T) {
	r := NewRouter(WithMethodNotAllowed(true))
	mustAddRoute(r, http.MethodGet, "/resource", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	mustAddRoute(r, http.MethodPost, "/resource", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	rec := serveRouter(r, http.MethodPatch, "/resource")
	assertResponseStatus(t, rec, http.StatusMethodNotAllowed)
	assertResponseHeader(t, rec, "Allow", "GET, POST")

	var body contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error.Code != contract.CodeMethodNotAllowed {
		t.Fatalf("code = %q, want %q", body.Error.Code, contract.CodeMethodNotAllowed)
	}
	if body.Error.Type != contract.TypeMethodNotAllowed {
		t.Fatalf("type = %q, want %q", body.Error.Type, contract.TypeMethodNotAllowed)
	}
	if body.Error.Category != contract.CategoryClient {
		t.Fatalf("category = %q, want %q", body.Error.Category, contract.CategoryClient)
	}
}
