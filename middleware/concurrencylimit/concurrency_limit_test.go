package concurrencylimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareAllowsConfiguredQueuedWaiter(t *testing.T) {
	mw := Middleware(1, 1, 200*time.Millisecond)
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

	firstDone := serveAsync(handler)
	<-blocker

	secondDone := serveAsync(handler)
	assertNotDone(t, secondDone, 20*time.Millisecond)

	close(release)
	if first := <-firstDone; first.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", first.Code)
	}
	if second := <-secondDone; second.Code != http.StatusOK {
		t.Fatalf("expected queued request to succeed, got %d", second.Code)
	}
}

func TestMiddlewareFailFastWhenNoQueueDepth(t *testing.T) {
	mw := Middleware(1, 0, 50*time.Millisecond)
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

	firstDone := serveAsync(handler)
	<-blocker

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	close(release)
	if first := <-firstDone; first.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", first.Code)
	}
}

func TestMiddlewareQueuedRequestTimesOut(t *testing.T) {
	mw := Middleware(1, 1, 10*time.Millisecond)
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

	firstDone := serveAsync(handler)
	<-blocker

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	close(release)
	if first := <-firstDone; first.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", first.Code)
	}
}

func serveAsync(handler http.Handler) <-chan *httptest.ResponseRecorder {
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		done <- rec
	}()
	return done
}

func assertNotDone(t *testing.T, done <-chan *httptest.ResponseRecorder, d time.Duration) {
	t.Helper()

	select {
	case rec := <-done:
		t.Fatalf("request completed before release with status %d", rec.Code)
	case <-time.After(d):
	}
}
