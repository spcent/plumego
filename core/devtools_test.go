package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/core/components/devtools"
)

func TestDevToolsRoutesEndpoint(t *testing.T) {
	app := New(WithDebug(), WithDevTools())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devtools.DevToolsRoutesPath, nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	if body := resp.Body.String(); body == "" || !containsRoute(body, "/ping") {
		t.Fatalf("expected routes output to include /ping, got %q", body)
	}
}

func TestDevToolsRoutesJSONEndpoint(t *testing.T) {
	app := New(WithDebug(), WithDevTools())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devtools.DevToolsRoutesJSONPath, nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", payload["data"])
	}
	routes, ok := data["routes"].([]any)
	if !ok || len(routes) == 0 {
		t.Fatalf("expected routes array, got %#v", data["routes"])
	}
}

func TestDevToolsConfigEndpoint(t *testing.T) {
	app := New(WithDebug(), WithDevTools())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devtools.DevToolsConfigPath, nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode config payload: %v", err)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", payload["data"])
	}
	if debug, ok := data["debug"].(bool); !ok || !debug {
		t.Fatalf("expected debug true, got %#v", data["debug"])
	}
}

func TestDevToolsMetricsEndpoints(t *testing.T) {
	app := New(WithDebug(), WithDevTools())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Generate one request metric
	pingReq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	pingResp := httptest.NewRecorder()
	app.ServeHTTP(pingResp, pingReq)

	req := httptest.NewRequest(http.MethodGet, devtools.DevToolsMetricsPath, nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode metrics payload: %v", err)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", payload["data"])
	}
	if enabled, ok := data["enabled"].(bool); !ok || !enabled {
		t.Fatalf("expected enabled true, got %#v", data["enabled"])
	}
	if _, ok := data["db"]; !ok {
		t.Fatalf("expected db metrics in payload")
	}
}

func TestDevToolsMetricsClearEndpoint(t *testing.T) {
	app := New(WithDebug(), WithDevTools())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	pingReq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	pingResp := httptest.NewRecorder()
	app.ServeHTTP(pingResp, pingReq)

	clearReq := httptest.NewRequest(http.MethodPost, devtools.DevToolsMetricsClear, nil)
	clearResp := httptest.NewRecorder()
	app.ServeHTTP(clearResp, clearReq)

	if clearResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", clearResp.Code)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, devtools.DevToolsMetricsPath, nil)
	metricsResp := httptest.NewRecorder()
	app.ServeHTTP(metricsResp, metricsReq)

	if metricsResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", metricsResp.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(metricsResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode metrics payload: %v", err)
	}
	if _, ok := payload["data"]; !ok {
		t.Fatalf("expected data key in payload")
	}
}

func TestDevToolsPprofEndpoint(t *testing.T) {
	app := New(WithDebug(), WithDevTools())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devtools.DevToolsPprofBasePath, nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestDevToolsPprofPaths(t *testing.T) {
	app := New(WithDebug(), WithDevTools())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test pprof index path
	req := httptest.NewRequest(http.MethodGet, devtools.DevToolsPprofIndexPath, nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for pprof index, got %d", resp.Code)
	}

	// Test pprof cmdline path
	req = httptest.NewRequest(http.MethodGet, devtools.DevToolsPprofCmdline, nil)
	resp = httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for pprof cmdline, got %d", resp.Code)
	}

	// Test pprof profile path (use short duration to avoid long test timeouts)
	req = httptest.NewRequest(http.MethodGet, devtools.DevToolsPprofProfile+"?seconds=1", nil)
	resp = httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for pprof profile, got %d", resp.Code)
	}

	// Test pprof symbol path
	req = httptest.NewRequest(http.MethodGet, devtools.DevToolsPprofSymbol, nil)
	resp = httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for pprof symbol, got %d", resp.Code)
	}

	// Test pprof trace path (use short duration to avoid long test timeouts)
	req = httptest.NewRequest(http.MethodGet, devtools.DevToolsPprofTrace+"?seconds=1", nil)
	resp = httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for pprof trace, got %d", resp.Code)
	}
}

func TestDevToolsDisabledInNonDebug(t *testing.T) {
	app := New()
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devtools.DevToolsRoutesPath, nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func containsRoute(output, path string) bool {
	return len(output) > 0 && strings.Contains(output, path)
}
