package concurrencylimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddleware(t *testing.T) {
	mw := Middleware(1, 1, 50*time.Millisecond)
	blocker := make(chan struct{})
	release := make(chan struct{})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case blocker <- struct{}{}:
		default:
		}
		<-release
		w.WriteHeader(http.StatusOK)
	}))

	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		firstDone <- rec
	}()
	<-blocker

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	close(release)
	first := <-firstDone
	if first.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", first.Code)
	}
}
