package bodylimit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimit(t *testing.T) {
	mw := BodyLimit(10, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("this body is definitely too long"))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestBodyLimitAllowsExactLimit(t *testing.T) {
	mw := BodyLimit(5, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "hello" {
		t.Fatalf("expected exact body to pass through, got %q", got)
	}
}

func TestBodyLimitDetectsSingleReadOverrun(t *testing.T) {
	mw := BodyLimit(5, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 16)
		n, err := r.Body.Read(buf)
		if n != 5 {
			t.Fatalf("first read bytes = %d, want 5", n)
		}
		if err != errRequestTooLarge {
			t.Fatalf("first read error = %v, want %v", err, errRequestTooLarge)
		}
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("toolong"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "REQUEST_BODY_TOO_LARGE") {
		t.Fatalf("expected structured body limit error, got %q", rec.Body.String())
	}
}

func TestBodyLimitDisabledPassesThrough(t *testing.T) {
	mw := BodyLimit(0, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("unlimited"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "unlimited" {
		t.Fatalf("expected disabled limit to pass body through, got %q", got)
	}
}
