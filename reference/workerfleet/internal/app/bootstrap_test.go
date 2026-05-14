package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/core"
)

func TestBootstrapMemoryBackend(t *testing.T) {
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap memory: %v", err)
	}
	if runtime.Service == nil {
		t.Fatalf("service is nil")
	}
	if runtime.Metrics == nil {
		t.Fatalf("metrics collector is nil")
	}
	if err := runtime.Close(context.Background()); err != nil {
		t.Fatalf("close memory runtime: %v", err)
	}
}

func TestBootstrapMongoFailsClosedWhenMissingConfig(t *testing.T) {
	_, err := Bootstrap(context.Background(), Config{StoreBackend: StoreBackendMongo})
	if err == nil || !strings.Contains(err.Error(), "WORKERFLEET_MONGO_URI") {
		t.Fatalf("error = %v, want missing mongo uri", err)
	}
}

func TestNewAppWiresCoreMiddlewareAndRoutes(t *testing.T) {
	cfg := DefaultConfig()
	serverCfg := DefaultServerConfig()
	called := false

	app, err := New(context.Background(), cfg, serverCfg, func(coreApp *core.App, service *Service, ready func(context.Context) error, metrics http.Handler) error {
		called = true
		if service == nil {
			t.Fatalf("service is nil")
		}
		if ready == nil {
			t.Fatalf("ready is nil")
		}
		if metrics == nil {
			t.Fatalf("metrics is nil")
		}
		return coreApp.Get("/probe", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}))
	})
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	t.Cleanup(func() {
		if err := app.runtime.Close(context.Background()); err != nil {
			t.Fatalf("close runtime: %v", err)
		}
	})
	if !called {
		t.Fatalf("route registrar was not called")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	app.core.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestNewAppRejectsMissingRouteRegistrar(t *testing.T) {
	_, err := New(context.Background(), DefaultConfig(), DefaultServerConfig(), nil)
	if err == nil || !strings.Contains(err.Error(), "route registrar") {
		t.Fatalf("error = %v, want missing route registrar", err)
	}
}
