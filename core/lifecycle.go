package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spcent/plumego/config"
	"github.com/spcent/plumego/health"
	log "github.com/spcent/plumego/log"
)

// Boot initializes and starts the server.
func (a *App) Boot() error {
	if lifecycle, ok := a.logger.(log.Lifecycle); ok {
		if err := lifecycle.Start(context.Background()); err != nil {
			return err
		}
		defer lifecycle.Stop(context.Background())
	}

	if err := a.loadEnv(); err != nil {
		return err
	}

	a.configureBuiltIns()

	if a.config.Debug {
		os.Setenv("APP_DEBUG", "true")
	}

	if err := a.setupServer(); err != nil {
		return err
	}

	if err := a.startServer(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (a *App) configureBuiltIns() {
	a.ConfigurePubSub()
	a.ConfigureWebhookOut()
	a.ConfigureWebhookIn()
}

func (a *App) loadEnv() error {
	if a.envLoaded || a.config.EnvFile == "" {
		return nil
	}

	if _, err := os.Stat(a.config.EnvFile); err == nil {
		a.logger.Info("Load .env file", log.Fields{"path": a.config.EnvFile})
		err := config.LoadEnv(a.config.EnvFile, true)
		if err != nil {
			a.logger.Error("Load .env failed", log.Fields{"error": err})
			return err
		}
	}

	a.envLoaded = true
	return nil
}

func (a *App) setupServer() error {
	if os.Getenv("APP_DEBUG") == "true" {
		a.router.Print(os.Stdout)
	}

	a.applyGuardrails()
	a.buildHandler()

	a.httpServer = &http.Server{
		Addr:              a.config.Addr,
		Handler:           a.handler,
		ReadTimeout:       a.config.ReadTimeout,
		ReadHeaderTimeout: a.config.ReadHeaderTimeout,
		WriteTimeout:      a.config.WriteTimeout,
		IdleTimeout:       a.config.IdleTimeout,
		MaxHeaderBytes:    a.config.MaxHeaderBytes,
	}

	a.connTracker = newConnectionTracker(a.logger, a.config.DrainInterval)
	a.httpServer.ConnState = a.connTracker.track

	if !a.config.EnableHTTP2 {
		a.httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}

	return nil
}

func (a *App) startServer() error {
	a.started = true

	health.SetReady()

	idleConnsClosed := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig

		health.SetNotReady("draining connections")
		a.logger.Info("SIGTERM received, shutting down", nil)

		timeout := a.config.ShutdownTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if a.connTracker != nil {
			go a.connTracker.drain(ctx)
		}

		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("Server shutdown error", log.Fields{"error": err})
		}
		close(idleConnsClosed)
	}()

	a.logger.Info("Server running", log.Fields{"addr": a.config.Addr})

	var err error
	if a.config.TLS.Enabled {
		if a.config.TLS.CertFile == "" || a.config.TLS.KeyFile == "" {
			a.logger.Error("TLS enabled but certificate or key file not provided", nil)
			return fmt.Errorf("TLS enabled but certificate or key file not provided")
		}
		a.logger.Info("HTTPS enabled", log.Fields{"cert": a.config.TLS.CertFile})
		err = a.httpServer.ListenAndServeTLS(a.config.TLS.CertFile, a.config.TLS.KeyFile)
	} else {
		err = a.httpServer.ListenAndServe()
	}

	health.SetNotReady("shutting down")

	<-idleConnsClosed

	if a.wsHub != nil {
		a.logger.Info("Stopping WebSocket hub", nil)
		a.wsHub.Stop()
		a.wsHub = nil
	}

	a.logger.Info("Server stopped gracefully", nil)
	return err
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
