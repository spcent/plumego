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
	"reflect"
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
	// Get all components (built-in + custom)
	comps := append([]Component{}, a.builtInComponents()...)
	comps = append(comps, a.components...)

	if a.middlewareReg == nil {
		a.middlewareReg = middleware.NewRegistry()
	}

	// 1. Register all components in DI container
	for _, c := range comps {
		if c == nil {
			continue
		}
		a.diContainer.Register(c)
	}

	// 2. Inject dependencies into all components
	for _, c := range comps {
		if c == nil {
			continue
		}
		if err := a.diContainer.Inject(c); err != nil {
			a.logger.Error("Failed to inject dependencies", log.Fields{"component": reflect.TypeOf(c).String(), "error": err})
			continue
		}
	}

	// 3. Sort components based on dependencies
	sortedComps, err := a.sortComponentsByDependencies(comps)
	if err != nil {
		a.logger.Error("Failed to sort components", log.Fields{"error": err})
		// Fallback to original order if sorting fails
		sortedComps = comps
	}

	// 4. Register middleware and routes for sorted components
	for _, c := range sortedComps {
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

	return sortedComps
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

// sortComponentsByDependencies sorts components topologically based on their dependencies.
// It ensures that dependent components are started before their dependencies and stopped after.
func (a *App) sortComponentsByDependencies(components []Component) ([]Component, error) {
	if len(components) == 0 {
		return components, nil
	}

	// Create a map of component types to components
	typeToComponent := make(map[reflect.Type]Component)
	for _, comp := range components {
		if comp != nil {
			typeToComponent[reflect.TypeOf(comp)] = comp
		}
	}

	// Helper function to get all dependencies including transitive dependencies
	getAllDependencies := func(comp Component) ([]Component, error) {
		if comp == nil {
			return nil, nil
		}
		visited := make(map[reflect.Type]bool)
		var result []Component

		var dfs func(depType reflect.Type) error
		dfs = func(depType reflect.Type) error {
			if visited[depType] {
				return errors.New("circular dependency detected")
			}
			visited[depType] = true

			if depComp, exists := typeToComponent[depType]; exists && depComp != nil {
				// Add this dependency
				result = append(result, depComp)

				// Recursively get its dependencies
				for _, subDepType := range depComp.Dependencies() {
					if err := dfs(subDepType); err != nil {
						return err
					}
				}
			}

			visited[depType] = false
			return nil
		}

		// Start with the component's direct dependencies
		for _, depType := range comp.Dependencies() {
			if err := dfs(depType); err != nil {
				return nil, err
			}
		}

		return result, nil
	}

	// Collect all components with their dependencies in the correct order
	visited := make(map[Component]bool)
	var sortedComponents []Component

	// Process each component
	for _, comp := range components {
		if comp != nil && !visited[comp] {
			// Get all dependencies for this component
			dependencies, err := getAllDependencies(comp)
			if err != nil {
				return nil, err
			}

			// Add dependencies to sorted list (if not already present)
			for _, dep := range dependencies {
				if dep != nil && !visited[dep] {
					sortedComponents = append(sortedComponents, dep)
					visited[dep] = true
				}
			}

			// Add the component itself
			sortedComponents = append(sortedComponents, comp)
			visited[comp] = true
		}
	}

	return sortedComponents, nil
}
