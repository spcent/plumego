package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/requestid"
)

func TestAppDependenciesLogger(t *testing.T) {
	logger := log.NewGLogger()
	app := New(DefaultConfig(), AppDependencies{Logger: logger})
	if app.logger != logger {
		t.Errorf("expected logger to be set")
	}
}

func TestAppDependenciesLoggerDoesNotMirrorIntoDefaultRouter(t *testing.T) {
	logger := log.NewGLogger()
	app := New(DefaultConfig(), AppDependencies{Logger: logger})

	if app.Logger() != logger {
		t.Fatal("expected App.Logger to return the configured logger")
	}
	if app.router.Logger() != nil {
		t.Fatal("expected default router logger to remain unset")
	}
}

func TestNewWithNilLoggerFallsBackToNoOpLogger(t *testing.T) {
	var logger log.StructuredLogger
	app := New(DefaultConfig(), AppDependencies{Logger: logger})
	if app.logger == nil {
		t.Fatal("expected logger to be initialized")
	}
	if _, ok := app.logger.(*log.NoOpLogger); !ok {
		t.Fatalf("expected NoOpLogger when nil logger is provided, got %T", app.logger)
	}
}

func TestNewDefaultsToNoOpLogger(t *testing.T) {
	app := newTestApp()
	if app.logger == nil {
		t.Fatal("expected logger to be initialized")
	}
	if _, ok := app.logger.(*log.NoOpLogger); !ok {
		t.Fatalf("expected NoOpLogger by default, got %T", app.logger)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	app := newTestApp()
	app.Use(requestid.Middleware())
	mustRegisterRoute(t, app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	app.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatalf("expected X-Request-ID to be set")
	}
}

func TestMethodNotAllowedConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Router.MethodNotAllowed = true
	app := New(cfg, AppDependencies{})
	mustRegisterRoute(t, app.Get("/only", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("expected Allow header to include GET")
	}
}

func TestRouterConfiguresOwnedMethodNotAllowedPolicy(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Router.MethodNotAllowed = true
	app := New(cfg, AppDependencies{})
	mustRegisterRoute(t, app.Get("/only", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	if !app.router.MethodNotAllowedEnabled() {
		t.Fatal("expected owned router to have method-not-allowed enabled")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("expected Allow header to include GET")
	}
}

func TestRoutesAndLookupDoNotResyncRouterPolicy(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Router.MethodNotAllowed = true
	app := New(cfg, AppDependencies{})

	mustRegisterRoute(t, app.Get("/only", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

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
