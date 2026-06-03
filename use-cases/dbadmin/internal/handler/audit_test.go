package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/security/authn"
	kvstore "github.com/spcent/plumego/store/kv"

	"dbadmin/internal/domain/audit"
)

func TestAuditExportNDJSON(t *testing.T) {
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create kv store: %v", err)
	}
	store := audit.NewStore(kv)
	if err := store.Add(audit.Event{
		ID:        "evt-1",
		User:      "admin",
		Action:    "POST /api/connections",
		Method:    http.MethodPost,
		Path:      "/api/connections",
		Status:    http.StatusOK,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("add audit event: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/audit/export?format=ndjson", nil)
	rec := httptest.NewRecorder()
	AuditHandler{Store: store, Logger: testLogger{}}.Export(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/x-ndjson") {
		t.Fatalf("Content-Type = %q, want ndjson", got)
	}
	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("lines len = %d, want 1", len(lines))
	}
	var event audit.Event
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("decode ndjson line: %v", err)
	}
	if event.ID != "evt-1" || event.User != "admin" {
		t.Fatalf("unexpected event: %#v", event)
	}
}

func TestRoleMiddlewareReadonlyDenialIsAudited(t *testing.T) {
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create kv store: %v", err)
	}
	store := audit.NewStore(kv)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := AuditMiddleware(store, "readonly", testLogger{})(RoleMiddleware("readonly", testLogger{})(inner))

	req := httptest.NewRequest(http.MethodPost, "/api/connections", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req = req.WithContext(authn.WithPrincipal(req.Context(), &authn.Principal{Subject: "alice"}))
	req = req.WithContext(contract.WithRequestID(req.Context(), "req-123"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "DBADMIN_RBAC_DENIED") {
		t.Fatalf("response body missing RBAC code: %s", rec.Body.String())
	}

	events, err := store.List()
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	event := events[0]
	if event.User != "alice" || event.Role != "readonly" || event.RequestID != "req-123" {
		t.Fatalf("unexpected event identity fields: %#v", event)
	}
	if event.Status != http.StatusForbidden || event.DeniedReason != "role_readonly" {
		t.Fatalf("unexpected denial fields: %#v", event)
	}
}
