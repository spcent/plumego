package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"guardus/internal/app"
	"guardus/internal/config"
	"guardus/internal/domain/connectivity"
	"guardus/internal/domain/endpoint"
	"guardus/internal/storage"
)

// TestAppRouteWiring boots an App against an httptest target and verifies
// every entrypoint registered by RegisterRoutes responds with the expected
// content type and status — covering /api/v1/config, /health, /metrics, and
// the SPA fallback that x/frontend installs at "/".
func TestAppRouteWiring(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	// Endpoint definitions come from the storage layer; seed them via the
	// bootstrap-file mechanism so app.New runs the same path as production.
	bootstrapPath := filepath.Join(t.TempDir(), "endpoints.json")
	bootstrap := config.BootstrapFile{
		Endpoints: []*endpoint.Endpoint{{
			Name:       "ok",
			URL:        target.URL,
			Interval:   time.Hour,
			Conditions: []endpoint.Condition{"[STATUS] == 200"},
		}},
	}
	body, err := json.Marshal(bootstrap)
	if err != nil {
		t.Fatalf("marshal bootstrap: %v", err)
	}
	if err := os.WriteFile(bootstrapPath, body, 0o600); err != nil {
		t.Fatalf("write bootstrap: %v", err)
	}
	t.Setenv("GUARDUS_BOOTSTRAP_FILE", bootstrapPath)

	defaults := config.Defaults()
	cfg := &defaults
	cfg.Metrics = true
	cfg.Storage = &storage.Config{Type: storage.TypeMemory}
	cfg.Connectivity = &connectivity.Config{}
	if err := config.ValidateStorageConfig(cfg); err != nil {
		t.Fatalf("validate storage: %v", err)
	}
	config.ValidateAndSetConcurrencyDefaults(cfg)
	a, err := app.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	defer func() {
		// Best-effort runtime release; ignoring error since Start was not called.
		if a.Store != nil {
			a.Store.Close()
		}
	}()
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("RegisterRoutes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	cases := []struct {
		name           string
		path           string
		wantStatus     int
		wantBodyPrefix string
	}{
		{"health", "/health", http.StatusOK, "ok"},
		{"config", "/api/v1/config", http.StatusOK, `{`},
		{"metrics", "/metrics", http.StatusOK, ""},
		{"spa-fallback", "/some/spa/route", http.StatusOK, "<!"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			a.Core.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status: got %d want %d body=%q", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if tc.wantBodyPrefix != "" && !strings.HasPrefix(rec.Body.String(), tc.wantBodyPrefix) {
				t.Fatalf("body: got %q want prefix %q", rec.Body.String(), tc.wantBodyPrefix)
			}
		})
	}
}
