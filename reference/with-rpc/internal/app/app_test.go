package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoutesShape(t *testing.T) {
	ctx, cancel := newTestTimeout()
	defer cancel()

	a, err := New(ctx)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	defer func() {
		stopCtx, stopCancel := newTestTimeout()
		defer stopCancel()
		_ = a.Stop(stopCtx)
	}()

	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	routes := a.Core.Routes()
	if len(routes) != 1 {
		t.Fatalf("routes count = %d, want 1; routes = %v", len(routes), routes)
	}
	if routes[0].Method != http.MethodGet || routes[0].Path != "/v1/hello" {
		t.Errorf("route = %s %s, want GET /v1/hello", routes[0].Method, routes[0].Path)
	}
}

func TestHelloEndpointTranscoding(t *testing.T) {
	ctx, cancel := newTestTimeout()
	defer cancel()

	a, err := New(ctx)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	defer func() {
		stopCtx, stopCancel := newTestTimeout()
		defer stopCancel()
		_ = a.Stop(stopCtx)
	}()

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

	req := httptest.NewRequest(http.MethodGet, "/v1/hello", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["message"] == "" {
		t.Error("message should not be empty")
	}
}
