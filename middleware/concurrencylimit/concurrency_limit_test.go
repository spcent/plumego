package concurrencylimit

import (
	"context"
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

func TestMiddlewareWithConfigMatchesPositionalConstructor(t *testing.T) {
	for _, tt := range []struct {
		name string
		mw   func() func(http.Handler) http.Handler
	}{
		{name: "positional", mw: func() func(http.Handler) http.Handler {
			return Middleware(1, 0, 50*time.Millisecond)
		}},
		{name: "config", mw: func() func(http.Handler) http.Handler {
			return MiddlewareWithConfig(Config{MaxConcurrent: 1, QueueDepth: 0, QueueTimeout: 50 * time.Millisecond})
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			blocker := make(chan struct{})
			release := make(chan struct{})

			handler := tt.mw()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		})
	}
}

func TestDefaultConfigUsesDefaultQueueTimeout(t *testing.T) {
	cfg := DefaultConfig(3)
	if cfg.MaxConcurrent != 3 {
		t.Fatalf("MaxConcurrent = %d, want 3", cfg.MaxConcurrent)
	}
	if cfg.QueueTimeout != defaultQueueTimeout {
		t.Fatalf("QueueTimeout = %s, want %s", cfg.QueueTimeout, defaultQueueTimeout)
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

func TestMiddlewareQueuedRequestReturnsOnContextCancel(t *testing.T) {
	mw := Middleware(1, 1, time.Second)
	blocker := make(chan struct{}, 1)
	release := make(chan struct{})
	handlerCalls := make(chan struct{}, 2)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalls <- struct{}{}
		select {
		case blocker <- struct{}{}:
		default:
		}
		<-release
		w.WriteHeader(http.StatusOK)
	}))

	firstDone := serveAsync(handler)
	<-blocker

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	queuedDone := serveAsyncRequest(handler, req)
	assertNotDone(t, queuedDone, 20*time.Millisecond)

	cancel()
	select {
	case rec := <-queuedDone:
		if rec.Body.Len() != 0 {
			t.Fatalf("canceled queued request wrote body %q", rec.Body.String())
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("queued request did not return after context cancellation")
	}

	select {
	case <-handlerCalls:
		// First request.
	default:
		t.Fatal("expected first handler call")
	}
	select {
	case <-handlerCalls:
		t.Fatal("canceled queued request invoked downstream handler")
	default:
	}

	close(release)
	if first := <-firstDone; first.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d", first.Code)
	}
}

func serveAsync(handler http.Handler) <-chan *httptest.ResponseRecorder {
	return serveAsyncRequest(handler, httptest.NewRequest(http.MethodGet, "/", nil))
}

func serveAsyncRequest(handler http.Handler, req *http.Request) <-chan *httptest.ResponseRecorder {
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
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
