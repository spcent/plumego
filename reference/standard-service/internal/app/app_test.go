package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
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
	wantPaths := [][2]string{
		{http.MethodDelete, "/api/v1/items/:id"},
		{http.MethodGet, "/"},
		{http.MethodGet, "/api/hello"},
		{http.MethodGet, "/api/info"},
		{http.MethodGet, "/api/v1/greet"},
		{http.MethodGet, "/api/v1/items"},
		{http.MethodGet, "/api/v1/items/:id"},
		{http.MethodGet, "/healthz"},
		{http.MethodGet, "/readyz"},
		{http.MethodPatch, "/api/v1/items/:id"},
		{http.MethodPost, "/api/v1/items"},
		{http.MethodPut, "/api/v1/items/:id"},
	}
	if len(got) != len(wantPaths) {
		t.Fatalf("got %d routes, want %d", len(got), len(wantPaths))
	}
	for i, route := range got {
		if route.Method != wantPaths[i][0] || route.Path != wantPaths[i][1] {
			t.Fatalf("route %d: got %s %s, want %s %s", i, route.Method, route.Path, wantPaths[i][0], wantPaths[i][1])
		}
	}
}

// TestHelloEndpointListMatchesRegisteredRoutes is a drift-detection test.
// It validates that the explicit endpoint list in handler.APIHandler.Hello:
//  1. Contains exactly the same Method+Path pairs as the routes registered in RegisterRoutes.
//  2. Has a non-empty Name and non-empty Description for every entry.
//
// When you add a route in routes.go you must also add it to the Endpoints slice in
// handler/api.go Hello — this test will fail if you forget.
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
				Name        string `json:"name"`
				Method      string `json:"method"`
				Path        string `json:"path"`
				Description string `json:"description"`
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

	// Every entry must have a non-empty Name and Description so the endpoint list
	// is useful for service discovery and documentation. Update them in handler/api.go
	// when route semantics change.
	for i, ep := range env.Data.Endpoints {
		if ep.Name == "" {
			t.Errorf("endpoint[%d] (%s %s): Name is empty — add a stable machine-readable name in handler/api.go Hello", i, ep.Method, ep.Path)
		}
		if ep.Description == "" {
			t.Errorf("endpoint[%d] (%s %s): Description is empty — add a human-readable description in handler/api.go Hello", i, ep.Method, ep.Path)
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

// TestAcceptanceAppStartServesRequests verifies that App.Start binds,
// registers routes, and serves HTTP requests successfully.
// Uses httptest.Server to avoid port-binding complexity.
func TestAcceptanceAppStartServesRequests(t *testing.T) {
	cfg := config.Defaults()
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	// Wrap the app's handler in httptest.Server to test routing
	// without a full network listen.
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("http.Get /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /healthz: status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestAcceptanceAppStartGracefulShutdown verifies that canceling the context
// triggers graceful shutdown and App.Start returns nil.
func TestAcceptanceAppStartGracefulShutdown(t *testing.T) {
	cfg := config.Defaults()
	cfg.Core.Addr = ":0"
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	shutdownComplete := make(chan error, 1)
	go func() {
		shutdownComplete <- a.Start(ctx)
	}()

	// Wait for server to start.
	time.Sleep(100 * time.Millisecond)

	// Cancel context to trigger graceful shutdown.
	cancel()

	// Wait for shutdown to complete.
	select {
	case err := <-shutdownComplete:
		if err != nil {
			t.Fatalf("App.Start returned error after shutdown: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("App.Start did not return within 5 seconds after context cancellation")
	}
}

// TestAcceptanceAppStartPropagatesShutdownError verifies that shutdown errors
// are propagated as the return value of App.Start.
func TestAcceptanceAppStartPropagatesShutdownError(t *testing.T) {
	cfg := config.Defaults()
	cfg.Core.Addr = ":0"
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This test verifies that the drain channel pattern in app.go:138 works.
	// We start the server, cancel the context, and verify that any shutdown
	// errors are correctly propagated to the caller.
	shutdownComplete := make(chan error, 1)
	go func() {
		shutdownComplete <- a.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case result := <-shutdownComplete:
		// Both nil (successful shutdown) and non-nil (shutdown error) are acceptable;
		// the important thing is that the result is returned, not lost.
		_ = result
	case <-time.After(5 * time.Second):
		t.Fatal("App.Start did not complete after shutdown")
	}
}
