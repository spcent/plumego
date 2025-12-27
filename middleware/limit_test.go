package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
)

func TestBodyLimit(t *testing.T) {
	largeBody := strings.Repeat("a", 1024)
	req := httptest.NewRequest(http.MethodPost, "/limit", strings.NewReader(largeBody))
	w := httptest.NewRecorder()

	mw := BodyLimit(10, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", w.Code)
	}

	var resp contract.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error.Code != "request_body_too_large" {
		t.Errorf("unexpected error code: %s", resp.Error.Code)
	}
}

func TestConcurrencyLimit(t *testing.T) {
	mw := ConcurrencyLimit(1, 1, 50*time.Millisecond, nil)

	blocked := make(chan struct{})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		blocked <- struct{}{}
		time.Sleep(100 * time.Millisecond)
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)

	// First request should be allowed
	go handler.ServeHTTP(httptest.NewRecorder(), req.Clone(req.Context()))
	<-blocked

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req.Clone(req.Context()))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for queued request, got %d", w.Code)
	}

	var resp contract.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Code == "" {
		t.Errorf("expected structured error code")
	}
}
