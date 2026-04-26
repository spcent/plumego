package core

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/spcent/plumego/log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func writeTestTLSCertFiles(t *testing.T) (string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "127.0.0.1",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certFile, err := os.CreateTemp("", "plumego-test-cert-*.pem")
	if err != nil {
		t.Fatalf("create cert temp file: %v", err)
	}
	keyFile, err := os.CreateTemp("", "plumego-test-key-*.pem")
	if err != nil {
		t.Fatalf("create key temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(certFile.Name())
		_ = os.Remove(keyFile.Name())
	})

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		t.Fatalf("write cert pem: %v", err)
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		t.Fatalf("write key pem: %v", err)
	}
	if err := certFile.Close(); err != nil {
		t.Fatalf("close cert file: %v", err)
	}
	if err := keyFile.Close(); err != nil {
		t.Fatalf("close key file: %v", err)
	}

	return certFile.Name(), keyFile.Name()
}

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

func assertWrappedCoreError(t *testing.T, err error, operation string, message string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error")
	}
	want := "core " + operation + ": " + message
	if err.Error() != want {
		t.Fatalf("error message = %q, want %q", err.Error(), want)
	}
}

func TestPrepareServeAndShutdown(t *testing.T) {
	addr := requireNetwork(t)
	cfg := DefaultConfig()
	cfg.Addr = addr

	app := New(cfg, AppDependencies{})

	// Add a test route
	mustRegisterRoute(t, app.Get("/boot-test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("booted"))
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("prepare returned unexpected error: %v", err)
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
	if err := app.Shutdown(t.Context()); err != nil {
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

func TestPrepareMarksAppStarted(t *testing.T) {
	app := newTestApp()
	mustRegisterRoute(t, app.Get("/boot-order", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("prepare returned unexpected error: %v", err)
	}

	app.mu.RLock()
	state := app.preparationState
	app.mu.RUnlock()
	if state != PreparationStateServerPrepared {
		t.Fatalf("expected app to be server prepared, got %q", state)
	}
}

// TestPrepareConfiguresHTTPServer tests server preparation via the canonical public API.
func TestPrepareConfiguresHTTPServer(t *testing.T) {
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
				Addr:         ":8083",
				HTTP2Enabled: false,
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
			cfg := DefaultConfig()
			cfg.Addr = tt.config.Addr
			cfg.ReadTimeout = tt.config.ReadTimeout
			cfg.ReadHeaderTimeout = tt.config.ReadHeaderTimeout
			cfg.WriteTimeout = tt.config.WriteTimeout
			cfg.IdleTimeout = tt.config.IdleTimeout
			cfg.MaxHeaderBytes = tt.config.MaxHeaderBytes
			cfg.HTTP2Enabled = tt.config.HTTP2Enabled
			cfg.DrainInterval = tt.config.DrainInterval
			app := New(cfg, AppDependencies{})

			// Add a route to ensure handler is created
			mustRegisterRoute(t, app.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})))

			err := app.Prepare()

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
				if app.preparationState != PreparationStateServerPrepared {
					t.Errorf("preparation_state = %q, want %q", app.preparationState, PreparationStateServerPrepared)
				}
			}
		})
	}
}

func TestPrepareRejectsMissingTLSFiles(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TLS = TLSConfig{Enabled: true}
	app := New(cfg, AppDependencies{})

	mustRegisterRoute(t, app.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	assertWrappedCoreError(t, app.Prepare(), "prepare_server", "TLS enabled but certificate or key file not provided")
}

func TestServerReturnsWrappedErrorWhenNotPrepared(t *testing.T) {
	app := newTestApp()

	_, err := app.Server()

	assertWrappedCoreError(t, err, "get_server", "server not prepared")
}

func TestNilAppLifecycleEntrypointsReturnErrors(t *testing.T) {
	var app *App

	if err := app.Prepare(); err == nil || err.Error() != "core prepare_server: app is nil" {
		t.Fatalf("expected nil app prepare error, got %v", err)
	}
	if _, err := app.Server(); err == nil || err.Error() != "core get_server: app is nil" {
		t.Fatalf("expected nil app server error, got %v", err)
	}
	if err := app.Shutdown(nil); err == nil || err.Error() != "core shutdown_app: app is nil" {
		t.Fatalf("expected nil app shutdown error, got %v", err)
	}
}

func TestPrepareIsIdempotentAfterActivation(t *testing.T) {
	app := newTestApp()
	mustRegisterRoute(t, app.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("prepare returned unexpected error: %v", err)
	}
	if err := app.Prepare(); err != nil {
		t.Fatalf("second prepare returned unexpected error: %v", err)
	}
}

func TestPreparedServerCanServeTLSViaPublicPath(t *testing.T) {
	addr := requireNetwork(t)
	certFile, keyFile := writeTestTLSCertFiles(t)
	cfg := DefaultConfig()
	cfg.Addr = addr
	cfg.TLS = TLSConfig{Enabled: true, CertFile: certFile, KeyFile: keyFile}

	app := New(cfg, AppDependencies{})

	mustRegisterRoute(t, app.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("prepare returned unexpected error: %v", err)
	}
	srv, err := app.Server()
	if err != nil {
		t.Fatalf("server returned unexpected error: %v", err)
	}
	if srv.TLSConfig == nil || len(srv.TLSConfig.Certificates) != 1 {
		t.Fatalf("expected prepared server to have loaded TLS certificate")
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer ln.Close()

	serverDone := make(chan error)
	go func() {
		serverDone <- srv.ServeTLS(ln, "", "")
	}()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get("https://" + ln.Addr().String() + "/test")
	if err != nil {
		t.Fatalf("https get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if err := app.Shutdown(t.Context()); err != nil {
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

	t.Run("terminal states do not decrement below zero", func(t *testing.T) {
		ct := newConnectionTracker(nil, 100*time.Millisecond)

		ct.track(nil, http.StateClosed)
		ct.track(nil, http.StateHijacked)
		if ct.active.Load() != 0 {
			t.Fatalf("expected active connection count to stay at 0, got %d", ct.active.Load())
		}

		ct.track(nil, http.StateNew)
		ct.track(nil, http.StateClosed)
		ct.track(nil, http.StateClosed)
		if ct.active.Load() != 0 {
			t.Fatalf("expected duplicate terminal state to stay at 0, got %d", ct.active.Load())
		}
	})

	t.Run("drain with no active connections", func(t *testing.T) {
		ct := newConnectionTracker(nil, 100*time.Millisecond)
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		ct.drain(ctx)
		// Should return immediately
	})

	t.Run("drain with active connections", func(t *testing.T) {
		ct := newConnectionTracker(nil, 50*time.Millisecond)
		ct.active.Store(1)

		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
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

		ctx, cancel := context.WithCancel(t.Context())

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

func TestPrepareAndShutdownDoNotDriveLoggerLifecycle(t *testing.T) {
	addr := requireNetwork(t)
	logger := &testLifecycleLogger{}
	cfg := DefaultConfig()
	cfg.Addr = addr
	app := New(cfg, AppDependencies{Logger: logger})

	mustRegisterRoute(t, app.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("prepare returned unexpected error: %v", err)
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

	if logger.startCalled.Load() {
		t.Error("logger Start should not be called by core")
	}

	if err := app.Shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown returned unexpected error: %v", err)
	}

	select {
	case <-done:
		if logger.stopCalled.Load() {
			t.Error("logger Stop should not be called by core")
		}
	case <-time.After(1 * time.Second):
		t.Error("server did not complete")
	}
}
