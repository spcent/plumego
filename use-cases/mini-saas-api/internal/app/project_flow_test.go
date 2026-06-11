package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// doJSONIdem is doJSON plus an Idempotency-Key header.
func doJSONIdem(t *testing.T, h http.Handler, method, path, bearer, idemKey string, body any) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	req := buildJSONRequest(t, method, path, bearer, body)
	req.Header.Set("Idempotency-Key", idemKey)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, decodeBody(t, rec)
}

func TestAcceptanceProjectCRUD(t *testing.T) {
	h := newTestServer(t)
	access, _ := signup(t, h, "pm@example.com", "pm-space")

	// Create.
	rec, body := doJSON(t, h, http.MethodPost, "/api/v1/projects", access, map[string]any{
		"name": "Apollo", "description": "first project",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	id := body["data"].(map[string]any)["id"].(string)

	// List and show.
	rec, body = doJSON(t, h, http.MethodGet, "/api/v1/projects", access, nil)
	if rec.Code != http.StatusOK || body["data"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("list: %d %s", rec.Code, rec.Body.String())
	}
	rec, body = doJSON(t, h, http.MethodGet, "/api/v1/projects/"+id, access, nil)
	if rec.Code != http.StatusOK || body["data"].(map[string]any)["name"] != "Apollo" {
		t.Fatalf("show: %d %s", rec.Code, rec.Body.String())
	}

	// Update.
	rec, body = doJSON(t, h, http.MethodPut, "/api/v1/projects/"+id, access, map[string]any{
		"name": "Apollo 11", "status": "paused",
	})
	if rec.Code != http.StatusOK || body["data"].(map[string]any)["status"] != "paused" {
		t.Fatalf("update: %d %s", rec.Code, rec.Body.String())
	}

	// Delete (owner has admin rights).
	if rec, _ = doJSON(t, h, http.MethodDelete, "/api/v1/projects/"+id, access, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", rec.Code)
	}
	if rec, _ = doJSON(t, h, http.MethodGet, "/api/v1/projects/"+id, access, nil); rec.Code != http.StatusNotFound {
		t.Fatalf("show after delete: %d, want 404", rec.Code)
	}
}

func TestProjectCrossTenantIsolation(t *testing.T) {
	h := newTestServer(t)
	accessA, _ := signup(t, h, "ta@example.com", "iso-a")
	accessB, _ := signup(t, h, "tb@example.com", "iso-b")

	rec, body := doJSON(t, h, http.MethodPost, "/api/v1/projects", accessA, map[string]any{"name": "Secret"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d", rec.Code)
	}
	id := body["data"].(map[string]any)["id"].(string)

	// Tenant B sees 404 on read/update/delete of tenant A's project.
	if rec, _ = doJSON(t, h, http.MethodGet, "/api/v1/projects/"+id, accessB, nil); rec.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant show: %d, want 404", rec.Code)
	}
	if rec, _ = doJSON(t, h, http.MethodPut, "/api/v1/projects/"+id, accessB, map[string]any{"name": "Stolen"}); rec.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant update: %d, want 404", rec.Code)
	}
	if rec, _ = doJSON(t, h, http.MethodDelete, "/api/v1/projects/"+id, accessB, nil); rec.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant delete: %d, want 404", rec.Code)
	}
	if rec, body = doJSON(t, h, http.MethodGet, "/api/v1/projects", accessB, nil); body["data"].(map[string]any)["total"].(float64) != 0 {
		t.Fatalf("tenant B list leaked projects: %s", rec.Body.String())
	}
}

func TestProjectMemberCannotDelete(t *testing.T) {
	h := newTestServer(t)
	ownerAccess, _ := signup(t, h, "po@example.com", "perm-space")
	signup(t, h, "pm2@example.com", "pm2-own")

	rec, body := doJSON(t, h, http.MethodPost, "/api/v1/projects", ownerAccess, map[string]any{"name": "Keep"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d", rec.Code)
	}
	id := body["data"].(map[string]any)["id"].(string)

	if rec, _ = doJSON(t, h, http.MethodPost, "/api/v1/tenant/members", ownerAccess, map[string]any{"email": "pm2@example.com", "role": "member"}); rec.Code != http.StatusCreated {
		t.Fatalf("add member: %d", rec.Code)
	}
	memberAccess := login(t, h, "pm2@example.com", "perm-space")

	// Member can create and update but not delete.
	if rec, _ = doJSON(t, h, http.MethodPost, "/api/v1/projects", memberAccess, map[string]any{"name": "Mine"}); rec.Code != http.StatusCreated {
		t.Fatalf("member create: %d", rec.Code)
	}
	if rec, _ = doJSON(t, h, http.MethodDelete, "/api/v1/projects/"+id, memberAccess, nil); rec.Code != http.StatusForbidden {
		t.Fatalf("member delete: %d, want 403", rec.Code)
	}
}

func TestProjectPlanLimit(t *testing.T) {
	h := newTestServer(t)
	access, _ := signup(t, h, "lim@example.com", "lim-space")

	// Free plan allows 5 projects.
	for i := 0; i < 5; i++ {
		if rec, _ := doJSON(t, h, http.MethodPost, "/api/v1/projects", access, map[string]any{"name": "P"}); rec.Code != http.StatusCreated {
			t.Fatalf("create %d: %d", i, rec.Code)
		}
	}
	rec, _ := doJSON(t, h, http.MethodPost, "/api/v1/projects", access, map[string]any{"name": "P6"})
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("over plan limit: %d, want 429", rec.Code)
	}
}

func TestAcceptanceIdempotentCreate(t *testing.T) {
	h := newTestServer(t)
	access, _ := signup(t, h, "idem@example.com", "idem-space")

	payload := map[string]any{"name": "Once", "description": "exactly once"}
	rec1, body1 := doJSONIdem(t, h, http.MethodPost, "/api/v1/projects", access, "key-1", payload)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("first create: %d %s", rec1.Code, rec1.Body.String())
	}
	id1 := body1["data"].(map[string]any)["id"].(string)

	// Same key + same payload → replayed response, same project, no duplicate.
	rec2, body2 := doJSONIdem(t, h, http.MethodPost, "/api/v1/projects", access, "key-1", payload)
	if rec2.Code != http.StatusCreated {
		t.Fatalf("replay create: %d", rec2.Code)
	}
	if rec2.Header().Get("X-Idempotency-Replay") != "true" {
		t.Fatal("replay response missing X-Idempotency-Replay header")
	}
	if id2 := body2["data"].(map[string]any)["id"].(string); id2 != id1 {
		t.Fatalf("replay returned different project: %s != %s", id2, id1)
	}
	rec3, body3 := doJSON(t, h, http.MethodGet, "/api/v1/projects", access, nil)
	if total := body3["data"].(map[string]any)["total"].(float64); total != 1 {
		t.Fatalf("duplicate created: total = %v, want 1 (%s)", total, rec3.Body.String())
	}

	// Same key + different payload → 400 payload mismatch.
	rec4, _ := doJSONIdem(t, h, http.MethodPost, "/api/v1/projects", access, "key-1", map[string]any{"name": "Different"})
	if rec4.Code != http.StatusBadRequest {
		t.Fatalf("payload mismatch: %d, want 400", rec4.Code)
	}

	// Without the header, duplicates are not deduplicated.
	doJSON(t, h, http.MethodPost, "/api/v1/projects", access, map[string]any{"name": "Dup"})
	doJSON(t, h, http.MethodPost, "/api/v1/projects", access, map[string]any{"name": "Dup"})
	_, body5 := doJSON(t, h, http.MethodGet, "/api/v1/projects", access, nil)
	if total := body5["data"].(map[string]any)["total"].(float64); total != 3 {
		t.Fatalf("plain duplicates: total = %v, want 3", total)
	}
}

func TestIdempotencyKeysAreTenantScoped(t *testing.T) {
	h := newTestServer(t)
	accessA, _ := signup(t, h, "ika@example.com", "ik-a")
	accessB, _ := signup(t, h, "ikb@example.com", "ik-b")

	payload := map[string]any{"name": "Same Key Project"}
	if rec, _ := doJSONIdem(t, h, http.MethodPost, "/api/v1/projects", accessA, "shared-key", payload); rec.Code != http.StatusCreated {
		t.Fatalf("tenant A create: %d", rec.Code)
	}
	// Tenant B reusing the same client key must create its own project, not replay A's.
	rec, _ := doJSONIdem(t, h, http.MethodPost, "/api/v1/projects", accessB, "shared-key", payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("tenant B create: %d", rec.Code)
	}
	if rec.Header().Get("X-Idempotency-Replay") == "true" {
		t.Fatal("tenant B replayed tenant A's response — idempotency keys leak across tenants")
	}
}
