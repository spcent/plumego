package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
	"standard-service/internal/config"
)

func TestRegisterRoutesCanonicalShape(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	// router.Routes() returns entries sorted by Method then Path, so the want
	// slice must follow the same order: DELETE before GET before POST,
	// and paths alphabetically within each method.
	got := a.Core.Routes()
	// Sorted by Method then Path — DELETE < GET < POST < PUT.
	want := []router.RouteInfo{
		{Method: http.MethodDelete, Path: "/api/v1/items/:id"},
		{Method: http.MethodGet, Path: "/"},
		{Method: http.MethodGet, Path: "/api/hello"},
		{Method: http.MethodGet, Path: "/api/info"},
		{Method: http.MethodGet, Path: "/api/v1/greet"},
		{Method: http.MethodGet, Path: "/api/v1/items"},
		{Method: http.MethodGet, Path: "/api/v1/items/:id"},
		{Method: http.MethodGet, Path: "/healthz"},
		{Method: http.MethodGet, Path: "/readyz"},
		{Method: http.MethodPatch, Path: "/api/v1/items/:id"},
		{Method: http.MethodPost, Path: "/api/v1/items"},
		{Method: http.MethodPut, Path: "/api/v1/items/:id"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("routes = %#v, want %#v", got, want)
	}
}

// TestHelloEndpointListMatchesRegisteredRoutes is a drift-detection test.
// It validates that the explicit endpoint list in handler.APIHandler.Hello
// contains exactly the same Method+Path pairs as the routes registered in
// RegisterRoutes. When you add a route in routes.go you must also add it to
// the Endpoints slice in handler/api.go Hello — this test will fail if you forget.
func TestHelloEndpointListMatchesRegisteredRoutes(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/hello", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/hello: status = %d, want %d", rec.Code, http.StatusOK)
	}

	var env struct {
		Data struct {
			Endpoints []struct {
				Method string `json:"method"`
				Path   string `json:"path"`
			} `json:"endpoints"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode /api/hello response: %v", err)
	}

	// Build a sorted (method, path) set from the Hello response.
	type methodPath struct{ method, path string }
	helloSet := make(map[methodPath]bool, len(env.Data.Endpoints))
	for _, ep := range env.Data.Endpoints {
		helloSet[methodPath{ep.Method, ep.Path}] = true
	}

	// Every registered route must appear in the Hello endpoint list.
	registered := a.Core.Routes()
	for _, ri := range registered {
		key := methodPath{ri.Method, ri.Path}
		if !helloSet[key] {
			t.Errorf("registered route %s %s is missing from GET /api/hello endpoint list", ri.Method, ri.Path)
		}
	}

	// The Hello list must not contain paths that are not registered.
	registeredSet := make(map[methodPath]bool, len(registered))
	for _, ri := range registered {
		registeredSet[methodPath{ri.Method, ri.Path}] = true
	}
	for _, ep := range env.Data.Endpoints {
		key := methodPath{ep.Method, ep.Path}
		if !registeredSet[key] {
			t.Errorf("GET /api/hello lists %s %s which is not a registered route", ep.Method, ep.Path)
		}
	}
}

func TestNewWiresCanonicalMiddlewareShape(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	req.Header.Set(contract.RequestIDHeader, "req-test-1")
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(contract.RequestIDHeader); got != "req-test-1" {
		t.Fatalf("%s = %q, want %q", contract.RequestIDHeader, got, "req-test-1")
	}
}

// TestMiddlewareStackPanicRecovery verifies that the recovery middleware converts
// a handler panic into a 500 response instead of crashing the server.
func TestMiddlewareStackPanicRecovery(t *testing.T) {
	a, err := New(config.Defaults())
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	// Register a panicking handler to exercise the recovery middleware.
	// Must be done before Prepare() freezes the route table.
	if err := a.Core.Get("/api/v1/test-panic", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("intentional test panic: recovery middleware must convert this to 500")
	})); err != nil {
		t.Fatalf("register panic route: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/test-panic", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("panic recovery: status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

// TestMiddlewareStackBodyLimitRejection verifies that the bodylimit middleware
// rejects requests with oversized bodies with 413 before the handler runs.
func TestMiddlewareStackBodyLimitRejection(t *testing.T) {
	cfg := config.Defaults()
	cfg.App.MaxBodyBytes = 10 // tiny limit so the test body stays small
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare app: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("get server: %v", err)
	}

	body := strings.NewReader(strings.Repeat("x", 11))
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/items", body))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("body limit: status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}
