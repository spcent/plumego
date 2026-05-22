package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	if runtime.shell.ingest == nil {
		t.Fatalf("ingest runtime component is nil")
	}
	if runtime.shell.query == nil {
		t.Fatalf("query runtime component is nil")
	}
	if runtime.shell.loops == nil {
		t.Fatalf("loop runner is nil")
	}
	if runtime.shell.alerts == nil {
		t.Fatalf("alert runner is nil")
	}
	if err := runtime.Close(context.Background()); err != nil {
		t.Fatalf("close memory runtime: %v", err)
	}
}

func TestBootstrapBuildsSplitRuntimeShell(t *testing.T) {
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap memory: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})

	if runtime.Service != runtime.shell.query {
		t.Fatalf("runtime service should reuse the query service component")
	}
	if runtime.shell.ingest == nil || runtime.shell.query == nil || runtime.shell.loops == nil || runtime.shell.alerts == nil {
		t.Fatalf("runtime shell components not fully wired: %#v", runtime.shell)
	}
}

func TestRuntimeReadyChecksConfiguredStore(t *testing.T) {
	runtime, err := Bootstrap(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("bootstrap memory: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})
	if err := runtime.Ready(context.Background()); err != nil {
		t.Fatalf("ready check failed: %v", err)
	}
}

func TestBootstrapMongoFailsClosedWhenMissingConfig(t *testing.T) {
	_, err := Bootstrap(context.Background(), Config{StoreBackend: StoreBackendMongo})
	if err == nil || !strings.Contains(err.Error(), "WORKERFLEET_MONGO_URI") {
		t.Fatalf("error = %v, want missing mongo uri", err)
	}
}

func TestBootstrapUsesConfiguredPolicies(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Policy.Status.StaleAfter = 75 * time.Second
	cfg.Policy.Status.OfflineAfter = 150 * time.Second
	cfg.Policy.Status.StageStuckAfter = 12 * time.Minute
	cfg.Policy.Status.RestartBurstThreshold = 7
	cfg.Policy.Alert.StageStuckAfter = 20 * time.Minute
	cfg.Policy.Alert.RestartBurstThreshold = 9

	runtime, err := Bootstrap(context.Background(), cfg)
	if err != nil {
		t.Fatalf("bootstrap memory: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})

	if runtime.shell.loops.policy != cfg.Policy.Status {
		t.Fatalf("loop status policy = %#v, want %#v", runtime.shell.loops.policy, cfg.Policy.Status)
	}
	if runtime.shell.alerts.policy != cfg.Policy.Status {
		t.Fatalf("alert status policy = %#v, want %#v", runtime.shell.alerts.policy, cfg.Policy.Status)
	}
	if runtime.shell.alerts.alertPolicy != cfg.Policy.Alert {
		t.Fatalf("alert policy = %#v, want %#v", runtime.shell.alerts.alertPolicy, cfg.Policy.Alert)
	}
}

func TestBootstrapUsesConfiguredMetricOptions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Metrics.ExperimentalSeriesEnabled = false

	runtime, err := Bootstrap(context.Background(), cfg)
	if err != nil {
		t.Fatalf("bootstrap memory: %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Close(context.Background())
	})

	if runtime.shell.loops.metrics == nil {
		t.Fatalf("loop metrics observer is nil")
	}
	if runtime.shell.loops.metrics.ExperimentalMetricsEnabled() {
		t.Fatalf("experimental metrics should be disabled")
	}
}

func TestNewAppWiresCoreMiddlewareAndRoutes(t *testing.T) {
	cfg := DefaultConfig()
	serverCfg := DefaultServerConfig()
	called := false

	app, err := New(context.Background(), cfg, serverCfg, func(coreApp *core.App, service *Service, ready func(context.Context) error, metrics http.Handler, workerAuth WorkerIngressAuthConfig, adminAuth AdminAuthConfig) error {
		called = true
		if service == nil {
			t.Fatalf("service is nil")
		}
		if workerAuth.Token != "" {
			t.Fatalf("worker auth token = %q, want empty", workerAuth.Token)
		}
		if adminAuth.Required || adminAuth.Token != "" {
			t.Fatalf("admin auth = %#v, want disabled", adminAuth)
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
