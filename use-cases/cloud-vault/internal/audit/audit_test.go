package audit

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"cloud-vault/internal/database"
)

func openTestDB(t *testing.T) *database.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestLogger_LogAndList(t *testing.T) {
	db := openTestDB(t)
	logger := NewLogger(db.DB)
	ctx := context.Background()

	logger.Log(ctx, "user-1", "127.0.0.1", ActionCreate, ResourceDocument, "doc-1", map[string]any{"title": "hello"})
	logger.Log(ctx, "user-1", "127.0.0.1", ActionUpdate, ResourceDocument, "doc-1", nil)
	logger.Log(ctx, "user-2", "10.0.0.1", ActionCreate, ResourceTag, "tag-1", nil)

	events, total, err := logger.List(ctx, "", "", 50, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(events) != 3 {
		t.Fatalf("len(events) = %d, want 3", len(events))
	}

	filtered, filteredTotal, err := logger.List(ctx, ResourceDocument, "doc-1", 50, 0)
	if err != nil {
		t.Fatalf("List filtered: %v", err)
	}
	if filteredTotal != 2 {
		t.Fatalf("filteredTotal = %d, want 2", filteredTotal)
	}
	for _, e := range filtered {
		if e.ResourceType != ResourceDocument || e.ResourceID != "doc-1" {
			t.Errorf("unexpected event in filtered results: %+v", e)
		}
	}
}

func TestLogger_NilSafe(t *testing.T) {
	var logger *Logger
	logger.Log(context.Background(), "u", "1.2.3.4", ActionDelete, ResourceDocument, "d", nil)
}

func TestClientIP(t *testing.T) {
	req := &http.Request{Header: http.Header{}, RemoteAddr: "127.0.0.1:1234"}
	req.Header.Set("X-Forwarded-For", "9.9.9.9, 10.0.0.1")
	if ip := ClientIP(req); ip != "9.9.9.9" {
		t.Errorf("ClientIP() = %q, want 9.9.9.9", ip)
	}

	req2 := &http.Request{Header: http.Header{}, RemoteAddr: "192.168.1.5:8080"}
	if ip := ClientIP(req2); ip != "192.168.1.5" {
		t.Errorf("ClientIP() = %q, want 192.168.1.5", ip)
	}
}
