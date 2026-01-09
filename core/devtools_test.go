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

	routes, ok := payload["routes"].([]any)
	if !ok || len(routes) == 0 {
		t.Fatalf("expected routes array, got %#v", payload["routes"])
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
	if debug, ok := payload["debug"].(bool); !ok || !debug {
		t.Fatalf("expected debug true, got %#v", payload["debug"])
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
