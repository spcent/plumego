package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAcceptanceMetricsEndpoint(t *testing.T) {
	h := newTestServer(t)

	// Generate some traffic so the collector has data.
	doJSON(t, h, http.MethodGet, "/healthz", "", nil)
	doJSON(t, h, http.MethodGet, "/healthz", "", nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics: status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "mini_saas_api") {
		t.Fatalf("metrics output missing namespace; got:\n%.500s", rec.Body.String())
	}
}

func TestMetricsTokenGuard(t *testing.T) {
	cfg := testConfig(t)
	cfg.App.MetricsToken = "scrape-secret"
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	if err := a.RegisterRoutes(); err != nil {
		t.Fatalf("register routes: %v", err)
	}
	if err := a.Core.Prepare(); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		t.Fatalf("server: %v", err)
	}
	h := srv.Handler

	// No token → 401.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("metrics without token: %d, want 401", rec.Code)
	}
	// Wrong token → 401.
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("metrics with wrong token: %d, want 401", rec.Code)
	}
	// Correct token → 200.
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer scrape-secret")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics with correct token: %d, want 200", rec.Code)
	}
}

func TestAcceptanceAuditTrail(t *testing.T) {
	h := newTestServer(t)
	ownerAccess, _ := signup(t, h, "aud@example.com", "aud-space")
	signup(t, h, "aud2@example.com", "aud2-own")

	// Generate auditable activity: add member, create + delete project.
	if rec, _ := doJSON(t, h, http.MethodPost, "/api/v1/tenant/members", ownerAccess, map[string]any{"email": "aud2@example.com", "role": "member"}); rec.Code != http.StatusCreated {
		t.Fatalf("add member: %d", rec.Code)
	}
	rec, body := doJSON(t, h, http.MethodPost, "/api/v1/projects", ownerAccess, map[string]any{"name": "Audited"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create project: %d", rec.Code)
	}
	projectID := body["data"].(map[string]any)["id"].(string)
	if rec, _ = doJSON(t, h, http.MethodDelete, "/api/v1/projects/"+projectID, ownerAccess, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete project: %d", rec.Code)
	}

	// Owner reads the audit log, newest first.
	rec, body = doJSON(t, h, http.MethodGet, "/api/v1/tenant/audit", ownerAccess, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("audit list: %d %s", rec.Code, rec.Body.String())
	}
	entries := body["data"].(map[string]any)["entries"].([]any)
	if len(entries) < 4 { // tenant.created, member.added, project.created, project.deleted
		t.Fatalf("audit entries = %d, want >= 4", len(entries))
	}
	if entries[0].(map[string]any)["action"] != "project.deleted" {
		t.Fatalf("newest entry = %v, want project.deleted", entries[0].(map[string]any)["action"])
	}

	// Limit is honored.
	rec, body = doJSON(t, h, http.MethodGet, "/api/v1/tenant/audit?limit=2", ownerAccess, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("audit limit: %d", rec.Code)
	}
	if n := len(body["data"].(map[string]any)["entries"].([]any)); n != 2 {
		t.Fatalf("limited entries = %d, want 2", n)
	}

	// Members cannot read the audit log.
	memberAccess := login(t, h, "aud2@example.com", "aud-space")
	if rec, _ = doJSON(t, h, http.MethodGet, "/api/v1/tenant/audit", memberAccess, nil); rec.Code != http.StatusForbidden {
		t.Fatalf("member audit access: %d, want 403", rec.Code)
	}

	// Invalid limit → 400.
	if rec, _ = doJSON(t, h, http.MethodGet, "/api/v1/tenant/audit?limit=zero", ownerAccess, nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid limit: %d, want 400", rec.Code)
	}
}

// TestAuditNeverStoresSecrets samples the audit content after auth activity.
func TestAuditNeverStoresSecrets(t *testing.T) {
	h := newTestServer(t)
	access, refresh := signup(t, h, "sec@example.com", "sec-space")

	rec, _ := doJSON(t, h, http.MethodGet, "/api/v1/tenant/audit", access, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("audit list: %d", rec.Code)
	}
	raw := rec.Body.String()
	if strings.Contains(raw, testPassword) {
		t.Fatal("audit log contains a plaintext password")
	}
	if strings.Contains(raw, refresh) {
		t.Fatal("audit log contains a refresh token")
	}
}
