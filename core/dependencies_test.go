package core

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/requestid"
)

func TestAppDependenciesLogger(t *testing.T) {
	logger := log.NewLogger()
	app := New(DefaultConfig(), AppDependencies{Logger: logger})
	if app.logger != logger {
		t.Errorf("expected logger to be set")
	}
	if app.Logger() != logger {
		t.Fatal("expected App.Logger to return the configured logger")
	}
	if app.router == nil {
		t.Fatal("expected app to own a router instance")
	}
}

func TestNewFallsBackToDiscardLogger(t *testing.T) {
	tests := []struct {
		name         string
		dependencies AppDependencies
	}{
		{name: "default dependencies"},
		{name: "explicit nil logger", dependencies: AppDependencies{Logger: nil}},
	}

	wantType := reflect.TypeOf(log.NewLogger(log.LoggerConfig{Format: log.LoggerFormatDiscard}))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(DefaultConfig(), tt.dependencies)
			if app.logger == nil {
				t.Fatal("expected logger to be initialized")
			}
			if reflect.TypeOf(app.logger) != wantType {
				t.Fatalf("expected discard logger, got %T", app.logger)
			}
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	app := newTestApp()
	app.Use(requestid.Middleware())
	mustRegisterRoute(t, app.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	app.ServeHTTP(rec, req)

	if rec.Header().Get(contract.RequestIDHeader) == "" {
		t.Fatalf("expected %s to be set", contract.RequestIDHeader)
	}
}

func TestMethodNotAllowedConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Router.MethodNotAllowed = true
	app := New(cfg, AppDependencies{})
	mustRegisterRoute(t, app.Get("/only", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("expected Allow header to include GET and HEAD")
	}
}

func TestRouterConfiguresOwnedMethodNotAllowedPolicy(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Router.MethodNotAllowed = true
	app := New(cfg, AppDependencies{})
	mustRegisterRoute(t, app.Get("/only", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	if !app.router.MethodNotAllowedEnabled() {
		t.Fatal("expected owned router to have method-not-allowed enabled")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("expected Allow header to include GET and HEAD")
	}
}

func TestRoutesAndLookupDoNotResyncRouterPolicy(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Router.MethodNotAllowed = true
	app := New(cfg, AppDependencies{})

	mustRegisterRoute(t, app.Get("/only", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	if !app.router.MethodNotAllowedEnabled() {
		t.Fatal("expected owned router to have method-not-allowed enabled from constructor ownership")
	}

	app.config.Router.MethodNotAllowed = false
	_ = app.Routes()
	_ = app.URL("missing")

	if !app.router.MethodNotAllowedEnabled() {
		t.Fatal("expected read paths to not re-sync router policy")
	}
}
