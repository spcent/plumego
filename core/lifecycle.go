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
	"github.com/spcent/plumego/middleware"
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

	components := a.mountComponents()

	a.mu.RLock()
	debug := a.config.Debug
	a.mu.RUnlock()

	if debug {
		os.Setenv("APP_DEBUG", "true")
	}

	if err := a.setupServer(); err != nil {
		return err
	}

	if err := a.startComponents(context.Background(), components); err != nil {
		return err
	}

	if err := a.startServer(); err != nil && err != http.ErrServerClosed {
		a.stopComponents(context.Background())
		return err
	}

	return nil
}

func (a *App) loadEnv() error {
	a.mu.RLock()
	envFile := a.config.EnvFile
	a.mu.RUnlock()

	if a.envLoaded || envFile == "" {
		return nil
	}

	if _, err := os.Stat(envFile); err == nil {
		a.logger.Info("Load .env file", log.Fields{"path": envFile})
		err := config.LoadEnv(envFile, true)
		if err != nil {
			a.logger.Error("Load .env failed", log.Fields{"error": err})
			return err
		}
	}

	a.mu.Lock()
	a.envLoaded = true
	a.mu.Unlock()
	return nil
}

func (a *App) setupServer() error {
	a.mu.RLock()
	debug := a.config.Debug
	addr := a.config.Addr
	readTimeout := a.config.ReadTimeout
	readHeaderTimeout := a.config.ReadHeaderTimeout
	writeTimeout := a.config.WriteTimeout
	idleTimeout := a.config.IdleTimeout
	maxHeaderBytes := a.config.MaxHeaderBytes
	drainInterval := a.config.DrainInterval
	enableHTTP2 := a.config.EnableHTTP2
	a.mu.RUnlock()

	if debug {
		os.Setenv("APP_DEBUG", "true")
	}

	a.ensureHandler()

	a.mu.RLock()
	handler := a.handler
	a.mu.RUnlock()

	if handler == nil {
		return fmt.Errorf("handler not configured")
	}

	a.mu.Lock()
	a.httpServer = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

	a.connTracker = newConnectionTracker(a.logger, drainInterval)
	a.httpServer.ConnState = a.connTracker.track

	if !enableHTTP2 {
		a.httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}
	a.mu.Unlock()

	return nil
}

func (a *App) startServer() error {
	a.mu.Lock()
	a.started = true
	a.mu.Unlock()

	health.SetReady()

	// Get configuration for server startup
	a.mu.RLock()
	tlsEnabled := a.config.TLS.Enabled
	tlsCertFile := a.config.TLS.CertFile
	tlsKeyFile := a.config.TLS.KeyFile
	addr := a.config.Addr
	httpServer := a.httpServer
	a.mu.RUnlock()

	idleConnsClosed := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig

		health.SetNotReady("draining connections")
		a.logger.Info("SIGTERM received, shutting down", nil)

		a.mu.RLock()
		shutdownTimeout := a.config.ShutdownTimeout
		httpServer := a.httpServer
		connTracker := a.connTracker
		a.mu.RUnlock()

		if shutdownTimeout <= 0 {
			shutdownTimeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if connTracker != nil {
			go connTracker.drain(ctx)
		}

		if httpServer != nil {
			if err := httpServer.Shutdown(ctx); err != nil {
				a.logger.Error("Server shutdown error", log.Fields{"error": err})
			}
		}

		a.stopComponents(ctx)
		close(idleConnsClosed)
	}()

	a.logger.Info("Server running", log.Fields{"addr": addr})
	if a.config.Debug {
		a.logger.Info("Debug mode enabled", log.Fields{
			"routes":     devToolsRoutesPath,
			"middleware": devToolsMiddlewarePath,
			"config":     devToolsConfigPath,
			"reload":     devToolsReloadPath,
		})
	}

	var err error
	if tlsEnabled {
		if tlsCertFile == "" || tlsKeyFile == "" {
			a.logger.Error("TLS enabled but certificate or key file not provided", nil)
			return fmt.Errorf("TLS enabled but certificate or key file not provided")
		}
		a.logger.Info("HTTPS enabled", log.Fields{"cert": tlsCertFile})
		err = httpServer.ListenAndServeTLS(tlsCertFile, tlsKeyFile)
	} else {
		err = httpServer.ListenAndServe()
	}

	health.SetNotReady("shutting down")
	a.stopComponents(context.Background())

	<-idleConnsClosed

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

func (a *App) mountComponents() []Component {
	comps := append([]Component{}, a.builtInComponents()...)
	comps = append(comps, a.components...)

	if a.middlewareReg == nil {
		a.middlewareReg = middleware.NewRegistry()
	}

	for _, c := range comps {
		if c == nil {
			continue
		}

		c.RegisterMiddleware(a.middlewareReg)
		c.RegisterRoutes(a.router)
	}

	a.router.Freeze()
	a.mu.Lock()
	a.componentsMounted = true
	a.mu.Unlock()

	return comps
}

func (a *App) startComponents(ctx context.Context, comps []Component) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(comps) == 0 {
		return nil
	}

	started := make([]Component, 0, len(comps))
	for _, c := range comps {
		if c == nil {
			continue
		}
		if err := c.Start(ctx); err != nil {
			for i := len(started) - 1; i >= 0; i-- {
				_ = started[i].Stop(ctx)
			}
			return err
		}
		started = append(started, c)
	}

	a.startedComponents = started
	return nil
}

func (a *App) stopComponents(ctx context.Context) {
	a.componentStopOnce.Do(func() {
		a.mu.Lock()
		comps := append([]Component{}, a.startedComponents...)
		a.mu.Unlock()

		for i := len(comps) - 1; i >= 0; i-- {
			if comps[i] == nil {
				continue
			}
			_ = comps[i].Stop(ctx)
		}
	})
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
