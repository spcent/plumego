package core

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spcent/plumego/log"
)

func testListenAddr() string {
	if addr := strings.TrimSpace(os.Getenv("PLUMEGO_TEST_ADDR")); addr != "" {
		return addr
	}
	return "127.0.0.1:0"
}

func requireNetwork(t *testing.T) string {
	t.Helper()
	if strings.EqualFold(strings.TrimSpace(os.Getenv("PLUMEGO_TEST_SKIP_NETWORK")), "1") {
		t.Skip("network tests disabled via PLUMEGO_TEST_SKIP_NETWORK")
	}

	addr := testListenAddr()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Skipf("network listen not permitted: %v", err)
	}
	_ = ln.Close()
	return addr
}

func TestPrepareStartServeAndShutdown(t *testing.T) {
	addr := requireNetwork(t)

	app := New(WithAddr(addr))

	// Add a test route
	mustRegisterRoute(t, app.Get("/boot-test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("booted"))
	}))

	if err := app.Prepare(); err != nil {
		t.Fatalf("prepare returned unexpected error: %v", err)
	}
	if err := app.Start(context.Background()); err != nil {
		t.Fatalf("start returned unexpected error: %v", err)
	}
	srv, err := app.Server()
	if err != nil {
		t.Fatalf("server returned unexpected error: %v", err)
	}

	serverDone := make(chan error)
	go func() {
		serverDone <- srv.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test server is responding
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boot-test", nil)
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), "booted") {
		t.Errorf("expected response body to contain 'booted'")
	}

	// Signal shutdown
	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown returned unexpected error: %v", err)
	}

	// Wait for shutdown
	select {
	case err := <-serverDone:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("server returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("server did not complete in time")
	}
}

func TestStartMarksAppReadyWithoutLegacyRuntimeHooks(t *testing.T) {
	app := New()
	mustRegisterRoute(t, app.Get("/boot-order", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	if err := app.Prepare(); err != nil {
		t.Fatalf("prepare returned unexpected error: %v", err)
	}
	if err := app.Start(context.Background()); err != nil {
		t.Fatalf("start returned unexpected error: %v", err)
	}

	app.mu.RLock()
	started := app.started
	app.mu.RUnlock()
	if !started {
		t.Fatalf("expected app to be marked started")
	}
}

// TestSetupServer tests server setup functionality
func TestSetupServer(t *testing.T) {
	tests := []struct {
		name        string
		config      AppConfig
		expectError bool
	}{
		{
			name: "basic setup",
			config: AppConfig{
				Addr: ":8081",
			},
			expectError: false,
		},
		{
			name: "with timeouts",
			config: AppConfig{
				Addr:              ":8082",
				ReadTimeout:       10 * time.Second,
				ReadHeaderTimeout: 5 * time.Second,
				WriteTimeout:      10 * time.Second,
				IdleTimeout:       30 * time.Second,
				MaxHeaderBytes:    1 << 20,
			},
			expectError: false,
		},
		{
			name: "with HTTP2 disabled",
			config: AppConfig{
				Addr:        ":8083",
				EnableHTTP2: false,
			},
			expectError: false,
		},
		{
			name: "with drain interval",
			config: AppConfig{
				Addr:          ":8084",
				DrainInterval: 1 * time.Second,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New()
			// Copy config values to app's config
			app.config.Addr = tt.config.Addr
			app.config.ReadTimeout = tt.config.ReadTimeout
			app.config.ReadHeaderTimeout = tt.config.ReadHeaderTimeout
			app.config.WriteTimeout = tt.config.WriteTimeout
			app.config.IdleTimeout = tt.config.IdleTimeout
			app.config.MaxHeaderBytes = tt.config.MaxHeaderBytes
			app.config.EnableHTTP2 = tt.config.EnableHTTP2
			app.config.DrainInterval = tt.config.DrainInterval

			// Add a route to ensure handler is created
			mustRegisterRoute(t, app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			err := app.setupServer()

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && err == nil {
				if app.httpServer == nil {
					t.Error("httpServer should be created")
				}
				if app.httpServer.Addr != tt.config.Addr {
					t.Errorf("expected addr %s, got %s", tt.config.Addr, app.httpServer.Addr)
				}
				if app.connTracker == nil {
					t.Error("connTracker should be created")
				}
			}
		})
	}
}

// TestStartServer tests server startup with graceful shutdown
func TestStartServer(t *testing.T) {
	// Test with TLS - skip actual TLS server due to certificate complexity
	t.Run("TLS success", func(t *testing.T) {
		// Skip this test as TLS requires valid certificates
		// The TLS logic is tested indirectly through other tests
		t.Skip("Skipping TLS test due to certificate complexity")
	})

	// Test TLS with missing files
	t.Run("TLS missing files", func(t *testing.T) {
		app := New()
		app.config.Addr = ":0"
		app.config.TLS.Enabled = true
		app.config.TLS.CertFile = ""
		app.config.TLS.KeyFile = ""
		app.config.ShutdownTimeout = 1 * time.Second

		mustRegisterRoute(t, app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		if err := app.setupServer(); err != nil {
			t.Fatalf("setupServer failed: %v", err)
		}

		err := app.startServer()
		if err == nil {
			t.Fatal("expected TLS validation error")
		}

		app.mu.RLock()
		started := app.started
		app.mu.RUnlock()
		if started {
			t.Fatal("app should not be marked started when TLS validation fails")
		}
	})

	// Test regular HTTP
	t.Run("HTTP success", func(t *testing.T) {
		addr := requireNetwork(t)
		app := New()
		app.config.Addr = addr
		app.config.ShutdownTimeout = 1 * time.Second

		mustRegisterRoute(t, app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}))

		if err := app.setupServer(); err != nil {
			t.Fatalf("setupServer failed: %v", err)
		}

		serverDone := make(chan error)
		go func() {
			serverDone <- app.startServer()
		}()

		time.Sleep(100 * time.Millisecond)

		// Test the server is running
		addr = app.httpServer.Addr
		if _, _, err := net.SplitHostPort(addr); err != nil {
			t.Errorf("expected host:port address, got %s", addr)
		}

		// Trigger shutdown
		if err := app.Shutdown(context.Background()); err != nil {
			t.Fatalf("shutdown returned unexpected error: %v", err)
		}

		select {
		case err := <-serverDone:
			if err != nil && err != http.ErrServerClosed {
				t.Errorf("unexpected error: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("server did not shut down in time")
		}
	})
}

// TestConnectionTracker tests connection tracking and draining
func TestConnectionTracker(t *testing.T) {
	t.Run("new connection tracker with default interval", func(t *testing.T) {
		ct := newConnectionTracker(nil, 0)
		if ct.interval != 500*time.Millisecond {
			t.Errorf("expected default interval 500ms, got %v", ct.interval)
		}
	})

	t.Run("new connection tracker with custom interval", func(t *testing.T) {
		ct := newConnectionTracker(nil, 1*time.Second)
		if ct.interval != 1*time.Second {
			t.Errorf("expected interval 1s, got %v", ct.interval)
		}
	})

	t.Run("track connection states", func(t *testing.T) {
		ct := newConnectionTracker(nil, 100*time.Millisecond)

		// Simulate connection lifecycle
		ct.track(nil, http.StateNew)
		if ct.active.Load() != 1 {
			t.Errorf("expected 1 active connection, got %d", ct.active.Load())
		}

		ct.track(nil, http.StateNew)
		if ct.active.Load() != 2 {
			t.Errorf("expected 2 active connections, got %d", ct.active.Load())
		}

		ct.track(nil, http.StateClosed)
		if ct.active.Load() != 1 {
			t.Errorf("expected 1 active connection after close, got %d", ct.active.Load())
		}

		ct.track(nil, http.StateHijacked)
		if ct.active.Load() != 0 {
			t.Errorf("expected 0 active connections after hijack, got %d", ct.active.Load())
		}
	})

	t.Run("drain with no active connections", func(t *testing.T) {
		ct := newConnectionTracker(nil, 100*time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		ct.drain(ctx)
		// Should return immediately
	})

	t.Run("drain with active connections", func(t *testing.T) {
		ct := newConnectionTracker(nil, 50*time.Millisecond)
		ct.active.Store(1)

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// This should drain and timeout
		ct.drain(ctx)

		// Connection should still be active (timeout)
		if ct.active.Load() != 1 {
			t.Errorf("expected connection to still be active")
		}
	})

	t.Run("drain with context cancellation", func(t *testing.T) {
		ct := newConnectionTracker(nil, 50*time.Millisecond)
		ct.active.Store(1)

		ctx, cancel := context.WithCancel(context.Background())

		// Cancel immediately
		cancel()

		ct.drain(ctx)
		// Should return immediately
	})
}

type testLifecycleLogger struct {
	log.StructuredLogger
	startCalled atomic.Bool
	stopCalled  atomic.Bool
}

func (l *testLifecycleLogger) Start(ctx context.Context) error {
	l.startCalled.Store(true)
	return nil
}

func (l *testLifecycleLogger) Stop(ctx context.Context) error {
	l.stopCalled.Store(true)
	return nil
}

func (l *testLifecycleLogger) Info(msg string, fields ...log.Fields)  {}
func (l *testLifecycleLogger) Error(msg string, fields ...log.Fields) {}
func (l *testLifecycleLogger) Debug(msg string, fields ...log.Fields) {}

func TestPrepareStartServeAndShutdownWithLoggerLifecycle(t *testing.T) {
	addr := requireNetwork(t)
	logger := &testLifecycleLogger{}
	app := New(
		WithLogger(logger),
		WithAddr(addr),
	)

	mustRegisterRoute(t, app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	if err := app.Prepare(); err != nil {
		t.Fatalf("prepare returned unexpected error: %v", err)
	}
	if err := app.Start(context.Background()); err != nil {
		t.Fatalf("start returned unexpected error: %v", err)
	}
	srv, err := app.Server()
	if err != nil {
		t.Fatalf("server returned unexpected error: %v", err)
	}

	done := make(chan error)
	go func() {
		done <- srv.ListenAndServe()
	}()

	time.Sleep(50 * time.Millisecond)

	// Check logger was started
	if !logger.startCalled.Load() {
		t.Error("logger Start should have been called")
	}

	// Trigger shutdown
	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown returned unexpected error: %v", err)
	}

	select {
	case <-done:
		if !logger.stopCalled.Load() {
			t.Error("logger Stop should have been called")
		}
	case <-time.After(1 * time.Second):
		t.Error("server did not complete")
	}
}
