package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestMethodMismatchReturnsNotFound(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/only", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/only", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "404 page not found\n" {
		t.Fatalf("unexpected body: %q", body)
	}
	if allow := rec.Header().Get("Allow"); allow != "" {
		t.Fatalf("expected empty Allow header, got %q", allow)
	}
}

func TestAnyFallbackWhenMethodTreeMissing(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	r.Any("/fallback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("any"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/fallback", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "any" {
		t.Fatalf("expected %q, got %q", "any", body)
	}
}

func TestTrailingSlashNormalization(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := contract.Param(r, "id")
		w.Write([]byte(id))
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "123" {
		t.Fatalf("expected id %q, got %q", "123", body)
	}
}

func TestInternalDoubleSlashNotNormalized(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/users//123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestWildcardParamCapturesRemainder(t *testing.T) {
	r := NewRouter()
	r.GetFunc("/files/*path", func(w http.ResponseWriter, r *http.Request) {
		path, _ := contract.Param(r, "path")
		w.Write([]byte(path))
	})

	req := httptest.NewRequest(http.MethodGet, "/files/a/b/c.txt", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "a/b/c.txt" {
		t.Fatalf("expected %q, got %q", "a/b/c.txt", body)
	}
}
