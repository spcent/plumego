package core

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
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

func waitForHTTPStatus(t *testing.T, url string, status int) {
	t.Helper()

	waitForHTTPStatusWithClient(t, &http.Client{Timeout: 100 * time.Millisecond}, url, status)
}

func waitForHTTPStatusWithClient(t *testing.T, client *http.Client, url string, status int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == status {
				return
			}
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server did not become ready at %s with status %d: %v", url, status, lastErr)
}

func TestPrepareServeAndShutdown(t *testing.T) {
	addr := requireNetwork(t)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	cfg := DefaultConfig()
	cfg.Addr = ln.Addr().String()

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
		serverDone <- srv.Serve(ln)
	}()

	waitForHTTPStatus(t, "http://"+ln.Addr().String()+"/boot-test", http.StatusOK)

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

func TestPostShutdownContractKeepsPreparedAppState(t *testing.T) {
	addr := requireNetwork(t)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	cfg := DefaultConfig()
	cfg.Addr = ln.Addr().String()
	app := New(cfg, AppDependencies{})

	mustRegisterRoute(t, app.Get("/post-shutdown", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	srv, err := app.Server()
	if err != nil {
		t.Fatalf("Server returned error: %v", err)
	}

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- srv.Serve(ln)
	}()

	waitForHTTPStatus(t, "http://"+ln.Addr().String()+"/post-shutdown", http.StatusNoContent)

	if err := app.Shutdown(t.Context()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	select {
	case err := <-serverDone:
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("server returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not complete in time")
	}

	if got := app.PreparationState(); got != PreparationStateServerPrepared {
		t.Fatalf("post-shutdown preparation state = %q, want %q", got, PreparationStateServerPrepared)
	}
	if err := app.Shutdown(t.Context()); err != nil {
		t.Fatalf("second Shutdown returned error: %v", err)
	}
	if err := app.Prepare(); err != nil {
		t.Fatalf("post-shutdown Prepare returned error: %v", err)
	}
	postShutdownServer, err := app.Server()
	if err != nil {
		t.Fatalf("post-shutdown Server returned error: %v", err)
	}
	if postShutdownServer != srv {
		t.Fatalf("post-shutdown Server returned %p, want original %p", postShutdownServer, srv)
	}

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/post-shutdown", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("post-shutdown ServeHTTP status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	restartLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen restart tcp: %v", err)
	}
	if err := srv.Serve(restartLn); err != http.ErrServerClosed {
		t.Fatalf("post-shutdown Serve returned %v, want %v", err, http.ErrServerClosed)
	}
}

func TestPreparedServerHooksAndCallerOwnedOverrides(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HTTP2Enabled = false
	app := New(cfg, AppDependencies{})
	mustRegisterRoute(t, app.Get("/owned", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	srv, err := app.Server()
	if err != nil {
		t.Fatalf("Server returned error: %v", err)
	}

	if srv.Handler == nil {
		t.Fatal("expected prepared server handler")
	}
	if srv.ConnState == nil {
		t.Fatal("expected prepared server connection tracker hook")
	}
	if srv.TLSNextProto == nil {
		t.Fatal("expected disabled HTTP/2 policy to install TLSNextProto override")
	}
	if app.connTracker == nil {
		t.Fatal("expected app connection tracker")
	}

	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/owned", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("default server handler status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	srv.ConnState(nil, http.StateNew)
	if got := app.connTracker.open.Load(); got != 1 {
		t.Fatalf("default ConnState active count = %d, want 1", got)
	}
	srv.ConnState(nil, http.StateClosed)
	if got := app.connTracker.open.Load(); got != 0 {
		t.Fatalf("default ConnState active count after close = %d, want 0", got)
	}

	var customConnStateCalled atomic.Bool
	srv.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			customConnStateCalled.Store(true)
		}
	}
	srv.ConnState(nil, http.StateNew)
	if !customConnStateCalled.Load() {
		t.Fatal("expected caller-owned ConnState override to run")
	}
	if got := app.connTracker.open.Load(); got != 0 {
		t.Fatalf("caller-owned ConnState override changed core tracker count to %d, want 0", got)
	}

	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Override", "caller")
		w.WriteHeader(http.StatusTeapot)
	})
	overrideRec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(overrideRec, httptest.NewRequest(http.MethodGet, "/owned", nil))
	if overrideRec.Code != http.StatusTeapot {
		t.Fatalf("caller-owned handler status = %d, want %d", overrideRec.Code, http.StatusTeapot)
	}
	if overrideRec.Header().Get("X-Override") != "caller" {
		t.Fatal("expected caller-owned handler override response")
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

	assertCoreError(t, app.Prepare(), operationPrepareServer, "TLS enabled but certificate or key file not provided")
	if app.preparationState != PreparationStateMutable {
		t.Fatalf("expected failed TLS validation to leave app mutable, got %q", app.preparationState)
	}
	mustRegisterRoute(t, app.Get("/after-tls-config-error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))
	if _, err := app.Server(); err == nil {
		t.Fatal("expected Server to fail after rejected Prepare")
	}
}

func TestPrepareTLSLoadFailureDoesNotFreezeMutation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TLS = TLSConfig{
		Enabled:  true,
		CertFile: "testdata/missing-cert.pem",
		KeyFile:  "testdata/missing-key.pem",
	}
	app := New(cfg, AppDependencies{})

	mustRegisterRoute(t, app.Get("/before-tls-load-error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	err := app.Prepare()
	if err == nil {
		t.Fatal("expected TLS load error")
	}
	if !strings.Contains(err.Error(), "core prepare_server: load tls certificate") {
		t.Fatalf("expected wrapped TLS load error, got %v", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected TLS load error to wrap os.ErrNotExist, got %v", err)
	}
	if app.preparationState != PreparationStateMutable {
		t.Fatalf("expected TLS load failure to leave app mutable, got %q", app.preparationState)
	}
	mustRegisterRoute(t, app.Get("/after-tls-load-error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))
}

func TestPrepareRejectsInvalidServerConfigBeforeFreeze(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*AppConfig)
		message string
	}{
		{
			name: "empty address",
			mutate: func(cfg *AppConfig) {
				cfg.Addr = " "
			},
			message: "server address cannot be empty",
		},
		{
			name: "negative read timeout",
			mutate: func(cfg *AppConfig) {
				cfg.ReadTimeout = -time.Second
			},
			message: "read timeout cannot be negative",
		},
		{
			name: "negative read header timeout",
			mutate: func(cfg *AppConfig) {
				cfg.ReadHeaderTimeout = -time.Second
			},
			message: "read header timeout cannot be negative",
		},
		{
			name: "negative write timeout",
			mutate: func(cfg *AppConfig) {
				cfg.WriteTimeout = -time.Second
			},
			message: "write timeout cannot be negative",
		},
		{
			name: "negative idle timeout",
			mutate: func(cfg *AppConfig) {
				cfg.IdleTimeout = -time.Second
			},
			message: "idle timeout cannot be negative",
		},
		{
			name: "negative max header bytes",
			mutate: func(cfg *AppConfig) {
				cfg.MaxHeaderBytes = -1
			},
			message: "max header bytes cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.mutate(&cfg)
			app := New(cfg, AppDependencies{})
			mustRegisterRoute(t, app.Get("/before-config-error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})))

			assertCoreError(t, app.Prepare(), operationPrepareServer, tt.message)
			if app.preparationState != PreparationStateMutable {
				t.Fatalf("expected invalid config to leave app mutable, got %q", app.preparationState)
			}
			mustRegisterRoute(t, app.Get("/after-config-error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})))
		})
	}
}

func TestPrepareAcceptsZeroServerTimeouts(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ReadTimeout = 0
	cfg.ReadHeaderTimeout = 0
	cfg.WriteTimeout = 0
	cfg.IdleTimeout = 0
	cfg.MaxHeaderBytes = 0
	app := New(cfg, AppDependencies{})

	mustRegisterRoute(t, app.Get("/zero-timeouts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error for zero server timeouts: %v", err)
	}
}

func TestServerReturnsWrappedErrorWhenNotPrepared(t *testing.T) {
	app := newTestApp()

	_, err := app.Server()

	assertCoreError(t, err, operationGetServer, "server not prepared")
}

func TestPrepareUsesLoggerFallbackForConnectionTracker(t *testing.T) {
	app := newTestApp()
	app.logger = nil
	mustRegisterRoute(t, app.Get("/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if app.connTracker == nil {
		t.Fatal("expected connection tracker")
	}
	if app.connTracker.logger == nil {
		t.Fatal("expected connection tracker to use logger fallback")
	}
}

func TestNilAppLifecycleEntrypointsReturnErrors(t *testing.T) {
	var app *App

	assertCoreError(t, app.Prepare(), operationPrepareServer, "app is nil")
	_, err := app.Server()
	assertCoreError(t, err, operationGetServer, "app is nil")
	assertCoreError(t, app.Shutdown(nil), operationShutdownApp, "app is nil")
}

func TestShutdownBeforePrepareReturnsError(t *testing.T) {
	app := newTestApp()

	err := app.Shutdown(nil)

	assertCoreError(t, err, operationShutdownApp, "server not prepared")
}

func TestShutdownUsesLoggerFallbackOnError(t *testing.T) {
	app := newTestApp()
	app.logger = nil
	app.httpServer = &http.Server{}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_ = app.Shutdown(ctx)
}

func TestShutdownStartsDrainOnce(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DrainInterval = time.Hour
	app := New(cfg, AppDependencies{})
	mustRegisterRoute(t, app.Get("/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if app.connTracker == nil {
		t.Fatal("expected connection tracker")
	}
	app.connTracker.open.Store(1)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		t.Fatalf("first Shutdown returned error: %v", err)
	}
	if !app.connTracker.drainStarted.Load() {
		t.Fatal("expected first Shutdown to start drain logging")
	}
	if err := app.Shutdown(ctx); err != nil {
		t.Fatalf("second Shutdown returned error: %v", err)
	}
	if app.connTracker.startDrain(ctx) {
		t.Fatal("expected drain logging to start at most once")
	}
}

func TestShutdownCanceledContextDoesNotConsumeDrainStart(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DrainInterval = time.Hour
	app := New(cfg, AppDependencies{})
	mustRegisterRoute(t, app.Get("/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	if err := app.Prepare(); err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if app.connTracker == nil {
		t.Fatal("expected connection tracker")
	}
	app.connTracker.open.Store(1)

	canceledCtx, cancelCanceled := context.WithCancel(t.Context())
	cancelCanceled()

	if err := app.Shutdown(canceledCtx); err != nil {
		assertCoreError(t, err, operationShutdownApp, "context canceled")
	}
	if app.connTracker.drainStarted.Load() {
		t.Fatal("expected canceled shutdown context not to consume drain start")
	}

	liveCtx, cancelLive := context.WithCancel(t.Context())
	defer cancelLive()

	if err := app.Shutdown(liveCtx); err != nil {
		t.Fatalf("second Shutdown returned error: %v", err)
	}
	if !app.connTracker.drainStarted.Load() {
		t.Fatal("expected live shutdown context to start drain logging")
	}
	if app.connTracker.startDrain(liveCtx) {
		t.Fatal("expected live shutdown drain to keep once-only latch")
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

func TestConcurrentPrepareReturnsSameServer(t *testing.T) {
	app := newTestApp()
	mustRegisterRoute(t, app.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	const workers = 20
	start := make(chan struct{})
	errs := make(chan error, workers)
	servers := make(chan *http.Server, workers)

	for i := 0; i < workers; i++ {
		go func() {
			<-start
			if err := app.Prepare(); err != nil {
				errs <- err
				return
			}
			server, err := app.Server()
			if err != nil {
				errs <- err
				return
			}
			servers <- server
			errs <- nil
		}()
	}

	close(start)

	var first *http.Server
	for i := 0; i < workers; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("Prepare worker returned error: %v", err)
		}
		server := <-servers
		if first == nil {
			first = server
			continue
		}
		if server != first {
			t.Fatalf("expected same server pointer from concurrent Prepare, got %p and %p", first, server)
		}
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
	waitForHTTPStatusWithClient(t, client, "https://"+ln.Addr().String()+"/test", http.StatusOK)

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
		if ct.open.Load() != 1 {
			t.Errorf("expected 1 open connection, got %d", ct.open.Load())
		}

		ct.track(nil, http.StateNew)
		if ct.open.Load() != 2 {
			t.Errorf("expected 2 open connections, got %d", ct.open.Load())
		}

		ct.track(nil, http.StateClosed)
		if ct.open.Load() != 1 {
			t.Errorf("expected 1 open connection after close, got %d", ct.open.Load())
		}

		ct.track(nil, http.StateHijacked)
		if ct.open.Load() != 0 {
			t.Errorf("expected 0 open connections after hijack, got %d", ct.open.Load())
		}
	})

	t.Run("terminal states do not decrement below zero", func(t *testing.T) {
		ct := newConnectionTracker(nil, 100*time.Millisecond)

		ct.track(nil, http.StateClosed)
		ct.track(nil, http.StateHijacked)
		if ct.open.Load() != 0 {
			t.Fatalf("expected open connection count to stay at 0, got %d", ct.open.Load())
		}

		ct.track(nil, http.StateNew)
		ct.track(nil, http.StateClosed)
		ct.track(nil, http.StateClosed)
		if ct.open.Load() != 0 {
			t.Fatalf("expected duplicate terminal state to stay at 0, got %d", ct.open.Load())
		}
	})

	t.Run("drain with no open connections", func(t *testing.T) {
		ct := newConnectionTracker(nil, 100*time.Millisecond)
		ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
		defer cancel()

		ct.drain(ctx)
		// Should return immediately
	})

	t.Run("drain with open connections", func(t *testing.T) {
		ct := newConnectionTracker(nil, 50*time.Millisecond)
		ct.open.Store(1)

		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		defer cancel()

		// This should drain and timeout
		ct.drain(ctx)

		// Connection should still be active (timeout)
		if ct.open.Load() != 1 {
			t.Errorf("expected connection to still be active")
		}
	})

	t.Run("drain with context cancellation", func(t *testing.T) {
		ct := newConnectionTracker(nil, 50*time.Millisecond)
		ct.open.Store(1)

		ctx, cancel := context.WithCancel(t.Context())

		// Cancel immediately
		cancel()

		ct.drain(ctx)
		// Should return immediately
	})

	t.Run("drain cancellation with open connections allows retry", func(t *testing.T) {
		ct := newConnectionTracker(nil, time.Hour)
		ct.open.Store(1)

		canceledCtx, cancelCanceled := context.WithCancel(t.Context())
		if !ct.startDrain(canceledCtx) {
			t.Fatal("expected first drain attempt to start")
		}
		cancelCanceled()

		deadline := time.After(500 * time.Millisecond)
		for ct.drainStarted.Load() {
			select {
			case <-deadline:
				t.Fatal("expected canceled drain to release start latch")
			case <-time.After(10 * time.Millisecond):
			}
		}

		liveCtx, cancelLive := context.WithCancel(t.Context())
		defer cancelLive()
		if !ct.startDrain(liveCtx) {
			t.Fatal("expected live context to retry drain")
		}
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
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	logger := &testLifecycleLogger{}
	cfg := DefaultConfig()
	cfg.Addr = ln.Addr().String()
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
		done <- srv.Serve(ln)
	}()

	waitForHTTPStatus(t, "http://"+ln.Addr().String()+"/test", http.StatusOK)

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
