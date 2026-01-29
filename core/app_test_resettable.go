package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResettableOnceForTesting(t *testing.T) {
	app := New()

	// Test that handlerOnce can be reset
	app.ensureHandler()
	if !app.handlerOnce.IsDone() {
		t.Fatalf("handlerOnce should be done after ensureHandler")
	}

	// Reset for testing
	app.ResetForTesting()
	if app.handlerOnce.IsDone() {
		t.Fatalf("handlerOnce should not be done after reset")
	}

	// Test that handler can be built again
	app.ensureHandler()
	if !app.handlerOnce.IsDone() {
		t.Fatalf("handlerOnce should be done after second ensureHandler")
	}

	// Test that stop methods can be reset
	ctx := t.Context()
	app.stopComponents(ctx)
	app.stopRunners(ctx)
	app.runShutdownHooks(ctx)

	// Reset for testing
	app.ResetForTesting()

	// Test that stop methods can be called again
	app.stopComponents(ctx)
	app.stopRunners(ctx)
	app.runShutdownHooks(ctx)
}

func TestHandlerRebuildAfterReset(t *testing.T) {
	app := New()

	// Add a route
	app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Build handler
	app.ensureHandler()

	// Test that handler works
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "test" {
		t.Fatalf("expected body 'test', got %q", rr.Body.String())
	}

	// Reset for testing
	app.ResetForTesting()

	// Add another route
	app.Get("/test2", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test2"))
	})

	// Build handler again
	app.ensureHandler()

	// Test that both routes work
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	rr = httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /test, got %d", rr.Code)
	}
	if rr.Body.String() != "test" {
		t.Fatalf("expected body 'test' for /test, got %q", rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/test2", nil)
	rr = httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /test2, got %d", rr.Code)
	}
	if rr.Body.String() != "test2" {
		t.Fatalf("expected body 'test2' for /test2, got %q", rr.Body.String())
	}
}
