package core

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spcent/plumego/log"
)

// Boot is the legacy convenience entrypoint.
// Prefer Prepare + Start + Server/Shutdown for explicit lifecycle control.
func (a *App) Boot() error {
	return a.Run(context.Background())
}

// Run prepares the app, starts runtime hooks, then serves until shutdown.
func (a *App) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := a.Prepare(); err != nil {
		return err
	}
	if err := a.Start(ctx); err != nil {
		return err
	}

	stopSignalWatcher := make(chan struct{})
	go a.handleShutdownSignal(stopSignalWatcher)

	err := a.startServer()
	close(stopSignalWatcher)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		_ = a.Shutdown(context.Background())
		return err
	}

	return nil
}

// Prepare freezes routing/middleware state and constructs the HTTP server.
func (a *App) Prepare() error {
	if err := a.setupServer(); err != nil {
		return err
	}

	return nil
}

// Server returns the prepared HTTP server instance.
func (a *App) Server() (*http.Server, error) {
	a.mu.RLock()
	server := a.httpServer
	a.mu.RUnlock()

	if server == nil {
		return nil, fmt.Errorf("server not prepared")
	}
	return server, nil
}

// Start starts logger lifecycle hooks and marks the app ready.
func (a *App) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.mu.RLock()
	if a.started {
		a.mu.RUnlock()
		return fmt.Errorf("app already started")
	}
	prepared := a.httpServer != nil
	logger := a.logger
	a.mu.RUnlock()

	if !prepared {
		return fmt.Errorf("app not prepared")
	}

	if lifecycle, ok := logger.(log.Lifecycle); ok {
		if err := lifecycle.Start(ctx); err != nil {
			return err
		}
		a.mu.Lock()
		a.loggerStarted = true
		a.mu.Unlock()
	}

	a.mu.Lock()
	a.started = true
	a.mu.Unlock()

	if a.healthManager != nil {
		a.healthManager.MarkReady()
	}

	return nil
}

// Shutdown gracefully stops the HTTP server and all runtime hooks.
func (a *App) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := a.serverStartConfig()
	if a.healthManager != nil {
		a.healthManager.MarkNotReady("shutting down")
	}

	if cfg.connTracker != nil {
		go cfg.connTracker.drain(ctx)
	}

	var shutdownErr error
	if cfg.httpServer != nil {
		if err := cfg.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("Server shutdown error", log.Fields{"error": err})
			shutdownErr = err
		}
	}

	a.stopRuntime(ctx)
	a.stopLoggerLifecycle(ctx)

	a.mu.Lock()
	a.started = false
	a.mu.Unlock()

	return shutdownErr
}

func (a *App) setupServer() error {
	a.mu.RLock()
	if a.httpServer != nil {
		a.mu.RUnlock()
		return nil
	}
	a.mu.RUnlock()

	a.ensureHandler()

	a.mu.RLock()
	handler := a.handler
	cfg := a.serverRuntimeConfigLocked()
	a.mu.RUnlock()

	if handler == nil {
		return fmt.Errorf("handler not configured")
	}

	a.mu.Lock()
	a.httpServer = &http.Server{
		Addr:              cfg.addr,
		Handler:           handler,
		ReadTimeout:       cfg.readTimeout,
		ReadHeaderTimeout: cfg.readHeaderTimeout,
		WriteTimeout:      cfg.writeTimeout,
		IdleTimeout:       cfg.idleTimeout,
		MaxHeaderBytes:    cfg.maxHeaderBytes,
	}

	a.connTracker = newConnectionTracker(a.logger, cfg.drainInterval)
	a.httpServer.ConnState = a.connTracker.track

	if !cfg.enableHTTP2 {
		a.httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}
	a.mu.Unlock()

	return nil
}

func (a *App) startServer() error {
	cfg := a.serverStartConfig()
	if cfg.httpServer == nil {
		return fmt.Errorf("server not configured")
	}

	if cfg.tlsEnabled && (cfg.tlsCertFile == "" || cfg.tlsKeyFile == "") {
		a.logger.Error("TLS enabled but certificate or key file not provided")
		return fmt.Errorf("TLS enabled but certificate or key file not provided")
	}

	a.logger.Info("Server running", log.Fields{"addr": cfg.addr})
	if cfg.debug {
		a.logger.Info("Debug mode enabled")
	}

	var err error
	if cfg.tlsEnabled {
		a.logger.Info("HTTPS enabled", log.Fields{"cert": cfg.tlsCertFile})
		err = cfg.httpServer.ListenAndServeTLS(cfg.tlsCertFile, cfg.tlsKeyFile)
	} else {
		err = cfg.httpServer.ListenAndServe()
	}
	return err
}

type serverRuntimeConfig struct {
	addr              string
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	maxHeaderBytes    int
	drainInterval     time.Duration
	enableHTTP2       bool
}

type serverStartConfig struct {
	addr            string
	debug           bool
	shutdownTimeout time.Duration
	tlsEnabled      bool
	tlsCertFile     string
	tlsKeyFile      string
	httpServer      *http.Server
	connTracker     *connectionTracker
}

func (a *App) serverRuntimeConfigLocked() serverRuntimeConfig {
	return serverRuntimeConfig{
		addr:              a.config.Addr,
		readTimeout:       a.config.ReadTimeout,
		readHeaderTimeout: a.config.ReadHeaderTimeout,
		writeTimeout:      a.config.WriteTimeout,
		idleTimeout:       a.config.IdleTimeout,
		maxHeaderBytes:    a.config.MaxHeaderBytes,
		drainInterval:     a.config.DrainInterval,
		enableHTTP2:       a.config.EnableHTTP2,
	}
}

func (a *App) serverStartConfig() serverStartConfig {
	a.mu.RLock()
	cfg := serverStartConfig{
		addr:            a.config.Addr,
		debug:           a.config.Debug,
		shutdownTimeout: a.config.ShutdownTimeout,
		tlsEnabled:      a.config.TLS.Enabled,
		tlsCertFile:     a.config.TLS.CertFile,
		tlsKeyFile:      a.config.TLS.KeyFile,
		httpServer:      a.httpServer,
		connTracker:     a.connTracker,
	}
	a.mu.RUnlock()
	return cfg
}

func (a *App) handleShutdownSignal(stop <-chan struct{}) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)

	select {
	case <-stop:
		return
	case <-sig:
	}

	a.logger.Info("SIGTERM received, shutting down")

	cfg := a.serverStartConfig()
	shutdownTimeout := cfg.shutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	_ = a.Shutdown(ctx)
}

func (a *App) stopRuntime(ctx context.Context) {
	_ = ctx
}

func (a *App) stopLoggerLifecycle(ctx context.Context) {
	a.mu.RLock()
	started := a.loggerStarted
	logger := a.logger
	a.mu.RUnlock()
	if !started {
		return
	}
	if lifecycle, ok := logger.(log.Lifecycle); ok {
		_ = lifecycle.Stop(ctx)
	}
	a.mu.Lock()
	a.loggerStarted = false
	a.mu.Unlock()
}

type connectionTracker struct {
	active   atomic.Int64         // Number of active connections
	logger   log.StructuredLogger // Logger for connection tracking
	interval time.Duration        // Interval at which to log active connections
}

func newConnectionTracker(logger log.StructuredLogger, interval time.Duration) *connectionTracker {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &connectionTracker{logger: logger, interval: interval}
}

func (t *connectionTracker) track(_ net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		t.active.Add(1)
	case http.StateHijacked, http.StateClosed:
		t.active.Add(-1)
	}
}

func (t *connectionTracker) drain(ctx context.Context) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		if t.active.Load() <= 0 {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if t.logger != nil {
				t.logger.Info("draining active connections", log.Fields{"active_connections": t.active.Load()})
			}
		}
	}
}
