package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFreezeNewCopiesConfigByValue(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = ":9091"
	cfg.ReadTimeout = 2 * time.Second

	app := New(cfg, AppDependencies{})
	cfg.Addr = ":9092"
	cfg.ReadTimeout = 5 * time.Second

	mustRegisterRoute(t, app.Get("/copy", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))
	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	server, err := app.Server()
	if err != nil {
		t.Fatalf("Server returned error: %v", err)
	}
	if server.Addr != ":9091" {
		t.Fatalf("server addr = %q, want %q", server.Addr, ":9091")
	}
	if server.ReadTimeout != 2*time.Second {
		t.Fatalf("server read timeout = %v, want %v", server.ReadTimeout, 2*time.Second)
	}
}

func TestFreezePrepareIsIdempotentAndReturnsSameServer(t *testing.T) {
	app := newTestApp()
	mustRegisterRoute(t, app.Get("/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("first Prepare returned error: %v", err)
	}
	first, err := app.Server()
	if err != nil {
		t.Fatalf("first Server returned error: %v", err)
	}
	if err := app.Prepare(); err != nil {
		t.Fatalf("second Prepare returned error: %v", err)
	}
	second, err := app.Server()
	if err != nil {
		t.Fatalf("second Server returned error: %v", err)
	}
	if first != second {
		t.Fatalf("Prepare should keep the same server pointer, got %p then %p", first, second)
	}
}

func TestFreezeServeHTTPBlocksLaterRouteRegistration(t *testing.T) {
	app := newTestApp()
	mustRegisterRoute(t, app.Get("/before", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("before"))
	})))

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/before", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("ServeHTTP status = %d, want %d", rec.Code, http.StatusOK)
	}

	err := app.Get("/after", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	if err == nil {
		t.Fatal("expected route registration after ServeHTTP to fail")
	}
	if !strings.Contains(err.Error(), "cannot register route after app has been prepared") {
		t.Fatalf("registration error = %q", err.Error())
	}

	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/after", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("late route should not be registered, status = %d", rec.Code)
	}
}
