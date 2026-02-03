package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDevToolsRoutesEndpoint(t *testing.T) {
	app := New(WithDebug())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devToolsRoutesPath, nil)
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
	app := New(WithDebug())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devToolsRoutesJSONPath, nil)
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
	app := New(WithDebug())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devToolsConfigPath, nil)
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
	app := New(WithDebug(), WithLogging())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Generate one request metric
	pingReq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	pingResp := httptest.NewRecorder()
	app.ServeHTTP(pingResp, pingReq)

	req := httptest.NewRequest(http.MethodGet, devToolsMetricsPath, nil)
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
	app := New(WithDebug(), WithLogging())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	pingReq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	pingResp := httptest.NewRecorder()
	app.ServeHTTP(pingResp, pingReq)

	clearReq := httptest.NewRequest(http.MethodPost, devToolsMetricsClear, nil)
	clearResp := httptest.NewRecorder()
	app.ServeHTTP(clearResp, clearReq)

	if clearResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", clearResp.Code)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, devToolsMetricsPath, nil)
	metricsResp := httptest.NewRecorder()
	app.ServeHTTP(metricsResp, metricsReq)

	if metricsResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", metricsResp.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(metricsResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode metrics payload: %v", err)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", payload["data"])
	}
	httpMetrics, ok := data["http"].(map[string]any)
	if !ok {
		t.Fatalf("expected http metrics, got %#v", data["http"])
	}
	total, ok := httpMetrics["total"].(map[string]any)
	if !ok {
		t.Fatalf("expected total metrics, got %#v", httpMetrics["total"])
	}
	if count, ok := total["count"].(float64); !ok || count != 0 {
		t.Fatalf("expected count 0 after clear, got %#v", total["count"])
	}
}

func TestDevToolsPprofEndpoint(t *testing.T) {
	app := New(WithDebug())
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devToolsPprofBasePath, nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestDevToolsDisabledInNonDebug(t *testing.T) {
	app := New()
	app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, devToolsRoutesPath, nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func containsRoute(output, path string) bool {
	return len(output) > 0 && strings.Contains(output, path)
}
