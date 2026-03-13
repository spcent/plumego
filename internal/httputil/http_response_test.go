package utils

import (
	"net/http/httptest"
	"testing"
)

func TestEnsureNoSniff(t *testing.T) {
	rec := httptest.NewRecorder()
	EnsureNoSniff(rec.Header())
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff, got %q", got)
	}
}

func TestEnsureNoSniffPreservesExistingValue(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("X-Content-Type-Options", "custom")
	EnsureNoSniff(rec.Header())
	if got := rec.Header().Get("X-Content-Type-Options"); got != "custom" {
		t.Fatalf("expected existing value to be preserved, got %q", got)
	}
}

func TestSafeWriteSetsNoSniffAndWritesBody(t *testing.T) {
	rec := httptest.NewRecorder()
	n, err := SafeWrite(rec, []byte("ok"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 bytes written, got %d", n)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff, got %q", got)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("expected body ok, got %q", got)
	}
}
