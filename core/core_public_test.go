package core_test

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
)

func TestPublicServeHTTPWorkflow(t *testing.T) {
	app := core.New(core.DefaultConfig(), core.AppDependencies{})

	if err := app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Core-Test", "middleware")
			next.ServeHTTP(w, r)
		})
	}); err != nil {
		t.Fatalf("Use returned error: %v", err)
	}
	if err := app.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("hello"))
	})); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("ServeHTTP status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if got := rec.Header().Get("X-Core-Test"); got != "middleware" {
		t.Fatalf("middleware header = %q, want middleware", got)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "hello" {
		t.Fatalf("response body = %q, want hello", got)
	}
	if got := app.PreparationState(); got != core.PreparationStateHandlerPrepared {
		t.Fatalf("preparation state = %q, want %q", got, core.PreparationStateHandlerPrepared)
	}
	if _, err := app.Server(); err == nil {
		t.Fatal("Server succeeded before explicit Prepare")
	}
}

func TestPublicPrepareServerShutdownWorkflow(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer ln.Close()

	cfg := core.DefaultConfig()
	cfg.Addr = ln.Addr().String()
	app := core.New(cfg, core.AppDependencies{})
	if err := app.Get("/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if got := app.PreparationState(); got != core.PreparationStateServerPrepared {
		t.Fatalf("preparation state = %q, want %q", got, core.PreparationStateServerPrepared)
	}
	srv, err := app.Server()
	if err != nil {
		t.Fatalf("Server returned error: %v", err)
	}
	if srv.Addr != cfg.Addr {
		t.Fatalf("server addr = %q, want %q", srv.Addr, cfg.Addr)
	}
	if srv.Handler == nil {
		t.Fatal("prepared server handler is nil")
	}

	done := make(chan error, 1)
	go func() {
		done <- srv.Serve(ln)
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	waitForPublicHTTPStatus(t, client, "http://"+ln.Addr().String()+"/ready", http.StatusNoContent)

	if err := app.Shutdown(t.Context()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("server returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop")
	}
}

func TestPublicRouteRegistrationErrors(t *testing.T) {
	app := core.New(core.DefaultConfig(), core.AppDependencies{})

	err := app.AddRoute(http.MethodGet, "/nil", nil)
	if !errors.Is(err, contract.ErrHandlerNil) {
		t.Fatalf("AddRoute nil error = %v, want ErrHandlerNil", err)
	}

	if err := app.Get("/ok", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if err := app.Get("/after-prepare", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})); err == nil {
		t.Fatal("route registration succeeded after Prepare")
	}
}

func waitForPublicHTTPStatus(t *testing.T, client *http.Client, url string, status int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == status {
				return
			}
			lastErr = errors.New(resp.Status)
		} else {
			lastErr = err
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s from %s: %v", http.StatusText(status), url, lastErr)
}
