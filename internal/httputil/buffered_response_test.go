package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBufferedResponseRecorderDoesNotWriteThrough(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := NewBufferedResponseRecorder(underlying)

	rec.Header().Set("X-Test", "recorded")
	rec.WriteHeader(http.StatusAccepted)
	n, err := rec.Write([]byte("buffered"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != len("buffered") {
		t.Fatalf("expected %d bytes written, got %d", len("buffered"), n)
	}

	if rec.StatusCode() != http.StatusAccepted {
		t.Fatalf("expected recorded status %d, got %d", http.StatusAccepted, rec.StatusCode())
	}
	if string(rec.Body()) != "buffered" {
		t.Fatalf("expected recorded body %q, got %q", "buffered", string(rec.Body()))
	}
	if got := rec.Header().Get("X-Test"); got != "recorded" {
		t.Fatalf("expected recorded header, got %q", got)
	}

	if underlying.Body.Len() != 0 {
		t.Fatalf("expected no write-through body, got %q", underlying.Body.String())
	}
	if got := underlying.Header().Get("X-Test"); got != "" {
		t.Fatalf("expected no write-through header, got %q", got)
	}
}

func TestBufferedResponseRecorderDefaultsStatusOK(t *testing.T) {
	rec := NewBufferedResponseRecorder(nil)

	if rec.StatusCode() != http.StatusOK {
		t.Fatalf("expected default status %d, got %d", http.StatusOK, rec.StatusCode())
	}

	_, err := rec.Write([]byte("body"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if rec.StatusCode() != http.StatusOK {
		t.Fatalf("expected write to record status %d, got %d", http.StatusOK, rec.StatusCode())
	}
}
