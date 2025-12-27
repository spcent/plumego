package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadinessHandler(t *testing.T) {
        t.Cleanup(func() { SetNotReady("starting") })
        SetNotReady("booting")
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()

	ReadinessHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when not ready, got %d", rr.Code)
	}

	var payload ReadinessStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if payload.Ready {
		t.Fatalf("expected ready=false, got true")
	}

	SetReady()
	rr = httptest.NewRecorder()
	ReadinessHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when ready, got %d", rr.Code)
	}
}

func TestBuildInfoHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/build", nil)
	rr := httptest.NewRecorder()

	BuildInfoHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	var info BuildInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &info); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if info.Version == "" {
		t.Fatalf("version should be populated from defaults")
	}
}
