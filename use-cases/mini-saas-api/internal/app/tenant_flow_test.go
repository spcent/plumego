package app

import (
	"net/http"
	"testing"
)

// login returns an access token for the given workspace slug.
func login(t *testing.T, h http.Handler, email, slug string) string {
	t.Helper()
	rec, body := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email":          email,
		"password":       testPassword,
		"workspace_slug": slug,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("login(%s, %s): status = %d, body = %s", email, slug, rec.Code, rec.Body.String())
	}
	return body["data"].(map[string]any)["tokens"].(map[string]any)["access_token"].(string)
}

func TestAcceptanceTenantAdminFlow(t *testing.T) {
	h := newTestServer(t)

	ownerAccess, _ := signup(t, h, "owner@example.com", "acme")
	signup(t, h, "mate@example.com", "mate-own-space")

	// Owner reads the workspace with usage.
	rec, body := doJSON(t, h, http.MethodGet, "/api/v1/tenant", ownerAccess, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get tenant: %d %s", rec.Code, rec.Body.String())
	}
	if body["data"].(map[string]any)["tenant"].(map[string]any)["slug"] != "acme" {
		t.Fatalf("wrong tenant: %s", rec.Body.String())
	}

	// Owner renames the workspace.
	rec, _ = doJSON(t, h, http.MethodPatch, "/api/v1/tenant", ownerAccess, map[string]any{"name": "Acme Renamed"})
	if rec.Code != http.StatusOK {
		t.Fatalf("patch tenant: %d %s", rec.Code, rec.Body.String())
	}

	// Owner adds mate as member.
	rec, body = doJSON(t, h, http.MethodPost, "/api/v1/tenant/members", ownerAccess, map[string]any{
		"email": "mate@example.com",
		"role":  "member",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("add member: %d %s", rec.Code, rec.Body.String())
	}
	membershipID := body["data"].(map[string]any)["id"].(string)

	// Member list shows both.
	rec, body = doJSON(t, h, http.MethodGet, "/api/v1/tenant/members", ownerAccess, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list members: %d", rec.Code)
	}
	if total := body["data"].(map[string]any)["total"].(float64); total != 2 {
		t.Fatalf("members total = %v, want 2", total)
	}

	// Mate logs into acme (member role): read OK, admin actions 403.
	mateAccess := login(t, h, "mate@example.com", "acme")
	if rec, _ = doJSON(t, h, http.MethodGet, "/api/v1/tenant", mateAccess, nil); rec.Code != http.StatusOK {
		t.Fatalf("member get tenant: %d", rec.Code)
	}
	if rec, _ = doJSON(t, h, http.MethodPatch, "/api/v1/tenant", mateAccess, map[string]any{"name": "Hijack"}); rec.Code != http.StatusForbidden {
		t.Fatalf("member patch tenant: %d, want 403", rec.Code)
	}
	if rec, _ = doJSON(t, h, http.MethodPost, "/api/v1/tenant/members", mateAccess, map[string]any{"email": "owner@example.com", "role": "member"}); rec.Code != http.StatusForbidden {
		t.Fatalf("member add member: %d, want 403", rec.Code)
	}

	// Owner promotes mate to admin; mate's rights apply after re-login.
	rec, _ = doJSON(t, h, http.MethodPatch, "/api/v1/tenant/members/"+membershipID, ownerAccess, map[string]any{"role": "admin"})
	if rec.Code != http.StatusOK {
		t.Fatalf("promote member: %d %s", rec.Code, rec.Body.String())
	}
	mateAdmin := login(t, h, "mate@example.com", "acme")
	if rec, _ = doJSON(t, h, http.MethodPatch, "/api/v1/tenant", mateAdmin, map[string]any{"name": "Mate Was Here"}); rec.Code != http.StatusOK {
		t.Fatalf("admin patch tenant: %d", rec.Code)
	}

	// Removing the member works; removing the last owner is blocked.
	rec, _ = doJSON(t, h, http.MethodDelete, "/api/v1/tenant/members/"+membershipID, ownerAccess, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("remove member: %d", rec.Code)
	}
	rec, body = doJSON(t, h, http.MethodGet, "/api/v1/tenant/members", ownerAccess, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list members after removal: %d", rec.Code)
	}
	ownerMembershipID := body["data"].(map[string]any)["members"].([]any)[0].(map[string]any)["id"].(string)
	if rec, _ = doJSON(t, h, http.MethodDelete, "/api/v1/tenant/members/"+ownerMembershipID, ownerAccess, nil); rec.Code != http.StatusConflict {
		t.Fatalf("remove last owner: %d, want 409", rec.Code)
	}
}

func TestMemberAddUnknownEmail(t *testing.T) {
	h := newTestServer(t)
	ownerAccess, _ := signup(t, h, "solo@example.com", "solo-space")
	rec, _ := doJSON(t, h, http.MethodPost, "/api/v1/tenant/members", ownerAccess, map[string]any{
		"email": "ghost@example.com",
		"role":  "member",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("add unknown email: %d, want 404", rec.Code)
	}
}

func TestTenantRateLimitEnforced(t *testing.T) {
	cfg := testConfig(t)
	cfg.App.TenantRPS = 1
	cfg.App.TenantBurst = 2
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

	access, _ := signup(t, h, "rl@example.com", "rl-space")

	// Burst capacity is 2: two requests pass, the third is throttled.
	got429 := false
	for i := 0; i < 4; i++ {
		rec, _ := doJSON(t, h, http.MethodGet, "/api/v1/me", access, nil)
		if rec.Code == http.StatusTooManyRequests {
			got429 = true
			break
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: unexpected status %d", i, rec.Code)
		}
	}
	if !got429 {
		t.Fatal("rate limit never triggered within burst+2 requests")
	}
}

func TestTenantQuotaEnforced(t *testing.T) {
	cfg := testConfig(t)
	cfg.App.TenantQuotaPerMinute = 2
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

	access, _ := signup(t, h, "q@example.com", "q-space")
	for i := 0; i < 2; i++ {
		if rec, _ := doJSON(t, h, http.MethodGet, "/api/v1/me", access, nil); rec.Code != http.StatusOK {
			t.Fatalf("request %d within quota: %d", i, rec.Code)
		}
	}
	rec, _ := doJSON(t, h, http.MethodGet, "/api/v1/me", access, nil)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("over-quota request: %d, want 429 (%s)", rec.Code, rec.Body.String())
	}
}

// TestTenantLimitsArePerTenant verifies one tenant exhausting its quota does
// not affect another tenant.
func TestTenantLimitsArePerTenant(t *testing.T) {
	cfg := testConfig(t)
	cfg.App.TenantQuotaPerMinute = 2
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

	accessA, _ := signup(t, h, "a@example.com", "space-a")
	accessB, _ := signup(t, h, "b@example.com", "space-b")

	// Exhaust tenant A's quota.
	for i := 0; ; i++ {
		rec, _ := doJSON(t, h, http.MethodGet, "/api/v1/me", accessA, nil)
		if rec.Code == http.StatusTooManyRequests {
			break
		}
		if i > 5 {
			t.Fatal("tenant A quota never triggered")
		}
	}
	// Tenant B is unaffected.
	if rec, _ := doJSON(t, h, http.MethodGet, "/api/v1/me", accessB, nil); rec.Code != http.StatusOK {
		t.Fatalf("tenant B blocked by tenant A's quota: %d", rec.Code)
	}
}
