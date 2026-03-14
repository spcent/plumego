package rest

import (
	"testing"

	"github.com/spcent/plumego/router"
)

func TestRegisterResourceRoutesUsesCanonicalRouteSurface(t *testing.T) {
	r := router.NewRouter()

	RegisterResourceRoutes(r, "/users/", NewBaseResourceController("users"), DefaultRouteOptions())

	assertRouteSet(t, r.Routes(), []routeKey{
		{"DELETE", "/users/:id"},
		{"DELETE", "/users/batch"},
		{"GET", "/users"},
		{"GET", "/users/:id"},
		{"HEAD", "/users"},
		{"HEAD", "/users/:id"},
		{"OPTIONS", "/users"},
		{"OPTIONS", "/users/:id"},
		{"PATCH", "/users/:id"},
		{"POST", "/users"},
		{"POST", "/users/batch"},
		{"PUT", "/users/:id"},
	})
}

func TestRegisterResourceRoutesRespectsRouteOptions(t *testing.T) {
	r := router.NewRouter()

	RegisterResourceRoutes(r, "users", NewBaseResourceController("users"), RouteOptions{})

	assertRouteSet(t, r.Routes(), []routeKey{
		{"DELETE", "/users/:id"},
		{"GET", "/users"},
		{"GET", "/users/:id"},
		{"PATCH", "/users/:id"},
		{"POST", "/users"},
		{"PUT", "/users/:id"},
	})
}

func TestRegisterContextResourceRoutesUsesCanonicalRouteSurface(t *testing.T) {
	r := router.NewRouter()

	RegisterContextResourceRoutes(r, "users/", NewBaseContextResourceController("users"))

	assertRouteSet(t, r.Routes(), []routeKey{
		{"DELETE", "/users/:id"},
		{"DELETE", "/users/batch"},
		{"GET", "/users"},
		{"GET", "/users/:id"},
		{"PATCH", "/users/:id"},
		{"POST", "/users"},
		{"POST", "/users/batch"},
		{"PUT", "/users/:id"},
	})
}

type routeKey struct {
	method string
	path   string
}

func assertRouteSet(t *testing.T, routes []router.RouteInfo, want []routeKey) {
	t.Helper()

	if len(routes) != len(want) {
		t.Fatalf("expected %d routes, got %d: %#v", len(want), len(routes), routes)
	}

	for i, route := range routes {
		if route.Method != want[i].method || route.Path != want[i].path {
			t.Fatalf("route %d: expected %s %s, got %s %s", i, want[i].method, want[i].path, route.Method, route.Path)
		}
	}
}
