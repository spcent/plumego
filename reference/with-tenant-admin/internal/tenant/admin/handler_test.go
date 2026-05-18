package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/reference/with-tenant-admin/internal/auth"
)

func TestCreateTenantReturnsCreatedWithID(t *testing.T) {
	h := NewHandler(NewInMemoryStore())
	rec := serveTenantAdmin(t, h.CreateTenant, http.MethodPost, "/admin/tenants", `{"name":"Acme"}`, true)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	record := decodeData[TenantRecord](t, rec)
	if record.ID == "" {
		t.Fatal("expected generated tenant id")
	}
	if record.Name != "Acme" || record.Status != StatusActive {
		t.Fatalf("unexpected tenant record: %+v", record)
	}
}

func TestGetTenantExistingReturnsRecord(t *testing.T) {
	store := NewInMemoryStore()
	record, _ := store.Create(t.Context(), TenantRecord{ID: "tenant-1", Name: "Acme"})
	h := NewHandler(store)

	rec := serveTenantAdmin(t, h.GetTenant, http.MethodGet, "/admin/tenants/"+record.ID, "", true)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	got := decodeData[TenantRecord](t, rec)
	if got.ID != record.ID || got.Name != record.Name {
		t.Fatalf("record = %+v, want %+v", got, record)
	}
}

func TestGetTenantUnknownReturnsNotFound(t *testing.T) {
	h := NewHandler(NewInMemoryStore())
	rec := serveTenantAdmin(t, h.GetTenant, http.MethodGet, "/admin/tenants/missing", "", true)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestSuspendTenantTransitionsStatus(t *testing.T) {
	store := NewInMemoryStore()
	record, _ := store.Create(t.Context(), TenantRecord{ID: "tenant-1", Name: "Acme"})
	h := NewHandler(store)

	rec := serveTenantAdmin(t, h.SuspendTenant, http.MethodPost, "/admin/tenants/"+record.ID+"/suspend", "", true)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	got := decodeData[TenantRecord](t, rec)
	if got.Status != StatusSuspended || got.SuspendedAt == nil {
		t.Fatalf("record = %+v, want suspended with timestamp", got)
	}
}

func TestDeleteTenantRemovesRecord(t *testing.T) {
	store := NewInMemoryStore()
	record, _ := store.Create(t.Context(), TenantRecord{ID: "tenant-1", Name: "Acme"})
	h := NewHandler(store)

	rec := serveTenantAdmin(t, h.DeleteTenant, http.MethodDelete, "/admin/tenants/"+record.ID, "", true)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec = serveTenantAdmin(t, h.GetTenant, http.MethodGet, "/admin/tenants/"+record.ID, "", true)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status after delete = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUnauthenticatedTenantAdminReturnsUnauthorized(t *testing.T) {
	h := NewHandler(NewInMemoryStore())
	rec := serveTenantAdmin(t, h.CreateTenant, http.MethodPost, "/admin/tenants", `{"name":"Acme"}`, false)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func serveTenantAdmin(t *testing.T, handler http.HandlerFunc, method string, path string, body string, authenticated bool) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if authenticated {
		req.Header.Set(auth.HeaderAdminToken, "secret")
	}
	auth.RequireAdminToken("secret")(handler).ServeHTTP(rec, req)
	return rec
}

func decodeData[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	var data T
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	return data
}
