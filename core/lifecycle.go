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
	"github.com/spcent/plumego/core/components/devtools"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
)

// Boot initializes and starts the server.
func (a *App) Boot() error {
	ctx := context.Background()

	if lifecycle, ok := a.logger.(log.Lifecycle); ok {
		if err := lifecycle.Start(ctx); err != nil {
			return err
		}
		defer lifecycle.Stop(ctx)
	}

	if err := a.runBootSequence(ctx); err != nil {
		return err
	}

	return nil
}

func (a *App) runBootSequence(ctx context.Context) error {
	if err := a.loadEnv(); err != nil {
		return err
	}

	if a.isDebugEnabled() {
		os.Setenv("APP_DEBUG", "true")
	}

	components := a.mountComponents()

	if err := a.setupServer(); err != nil {
		return err
	}

	if err := a.startComponents(ctx, components); err != nil {
		return err
	}

	if err := a.startRunners(ctx); err != nil {
		a.stopComponents(ctx)
		return err
	}

	if err := a.startServer(); err != nil && err != http.ErrServerClosed {
		a.stopRuntime(ctx)
		return err
	}

	return nil
}

func (a *App) isDebugEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config.Debug
}

func (a *App) loadEnv() error {
	a.mu.RLock()
	envFile := ""
	if a.config != nil {
		envFile = a.config.EnvFile
	}
	loaded := a.envLoaded
	a.mu.RUnlock()

	if loaded || envFile == "" {
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

	a.mu.Lock()
	a.started = true
	a.mu.Unlock()

	if a.healthManager != nil {
		a.healthManager.MarkReady()
	}

	stopSignalWatcher := make(chan struct{})
	go a.handleShutdownSignal(stopSignalWatcher)

	a.logger.Info("Server running", log.Fields{"addr": cfg.addr})
	if cfg.debug {
		a.logger.Info("Debug mode enabled", log.Fields{
			"routes":     devtools.DevToolsRoutesPath,
			"middleware": devtools.DevToolsMiddlewarePath,
			"config":     devtools.DevToolsConfigPath,
			"metrics":    devtools.DevToolsMetricsPath,
			"pprof":      devtools.DevToolsPprofBasePath,
			"reload":     devtools.DevToolsReloadPath,
		})
	}

	var err error
	if cfg.tlsEnabled {
		a.logger.Info("HTTPS enabled", log.Fields{"cert": cfg.tlsCertFile})
		err = cfg.httpServer.ListenAndServeTLS(cfg.tlsCertFile, cfg.tlsKeyFile)
	} else {
		err = cfg.httpServer.ListenAndServe()
	}

	close(stopSignalWatcher)

	if a.healthManager != nil {
		a.healthManager.MarkNotReady("shutting down")
	}
	a.stopRuntime(context.Background())
	a.logger.Info("Server stopped gracefully")
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

	if a.healthManager != nil {
		a.healthManager.MarkNotReady("draining connections")
	}
	a.logger.Info("SIGTERM received, shutting down")

	cfg := a.serverStartConfig()
	shutdownTimeout := cfg.shutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if cfg.connTracker != nil {
		go cfg.connTracker.drain(ctx)
	}

	if cfg.httpServer != nil {
		if err := cfg.httpServer.Shutdown(ctx); err != nil {
			a.logger.Error("Server shutdown error", log.Fields{"error": err})
		}
	}

	a.stopRuntime(ctx)
}

func (a *App) stopRuntime(ctx context.Context) {
	a.stopRunners(ctx)
	a.stopComponents(ctx)
	a.runShutdownHooks(ctx)
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
	// Get all components (built-in + custom) in declared order.
	declared := append([]Component{}, a.builtInComponents()...)
	declared = append(declared, a.components...)

	if a.middlewareReg == nil {
		a.middlewareReg = middleware.NewRegistry()
	}

	components := filterNilComponents(declared)

	for _, c := range components {
		c.RegisterMiddleware(a.middlewareReg)
		c.RegisterRoutes(a.router)
	}

	a.router.Freeze()
	a.mu.Lock()
	a.componentsMounted = true
	a.mu.Unlock()

	return components
}

func (a *App) startComponents(ctx context.Context, comps []Component) error {
	components := filterNilComponents(comps)
	if len(components) == 0 {
		return nil
	}

	started := make([]Component, 0, len(components))
	for _, c := range components {
		if err := c.Start(ctx); err != nil {
			for i := len(started) - 1; i >= 0; i-- {
				_ = started[i].Stop(ctx)
			}
			return err
		}
		started = append(started, c)
	}

	a.mu.Lock()
	a.startedComponents = started
	a.mu.Unlock()
	return nil
}

func (a *App) startRunners(ctx context.Context) error {
	a.mu.RLock()
	runners := append([]Runner{}, a.runners...)
	a.mu.RUnlock()

	runners = filterNilRunners(runners)
	if len(runners) == 0 {
		return nil
	}

	started := make([]Runner, 0, len(runners))
	for _, runner := range runners {
		if err := runner.Start(ctx); err != nil {
			for i := len(started) - 1; i >= 0; i-- {
				_ = started[i].Stop(ctx)
			}
			return err
		}
		started = append(started, runner)
	}

	a.mu.Lock()
	a.startedRunners = started
	a.mu.Unlock()
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
			if err := comps[i].Stop(ctx); err != nil {
				a.logger.WithFields(log.Fields{
					"component_index": i,
					"error":           err.Error(),
				}).Error("failed to stop component")
			}
		}
	})
}

func (a *App) stopRunners(ctx context.Context) {
	a.runnerStopOnce.Do(func() {
		a.mu.Lock()
		runners := append([]Runner{}, a.startedRunners...)
		a.mu.Unlock()

		for i := len(runners) - 1; i >= 0; i-- {
			if runners[i] == nil {
				continue
			}
			if err := runners[i].Stop(ctx); err != nil {
				a.logger.WithFields(log.Fields{
					"runner_index": i,
					"error":        err.Error(),
				}).Error("failed to stop runner")
			}
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

func filterNilComponents(components []Component) []Component {
	filtered := make([]Component, 0, len(components))
	for _, comp := range components {
		if comp != nil {
			filtered = append(filtered, comp)
		}
	}
	return filtered
}

func filterNilRunners(runners []Runner) []Runner {
	filtered := make([]Runner, 0, len(runners))
	for _, runner := range runners {
		if runner != nil {
			filtered = append(filtered, runner)
		}
	}
	return filtered
}
