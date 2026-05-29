// Package app wires guardus dependencies and manages the server lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/compression"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/timeout"

	"guardus/internal/config"
	"guardus/internal/metrics"
	"guardus/internal/security"
	"guardus/internal/storage"
	"guardus/internal/watchdog"
)

// App holds guardus-wide dependencies.
//
// Cfg is a pointer because handlers and the watchdog hold the same struct;
// future admin writes will mutate it in place. Store and Metrics are
// constructed once and reused for the process lifetime.
type App struct {
	Core     *core.App
	Cfg      *config.Config
	Store    storage.Store
	Metrics  *metrics.Endpoints
	Watchdog *watchdog.Watchdog
	Auth     *security.BasicAuthenticator
}

// New constructs the App with explicit middleware wiring and dependency
// construction. The store, watchdog, and Prometheus collector are owned by
// the App and released by Start when ctx is canceled.
//
// ctx is used for the storage operations performed during bootstrap and the
// initial endpoint load; it should be the same lifecycle context that Start
// will receive.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	c := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})

	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: c.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	accesslogMw, err := accesslog.Middleware(accesslog.Config{Logger: c.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure access log middleware: %w", err)
	}
	compressionMw := compression.Middleware(compression.Config{})
	timeoutMw := timeout.Middleware(timeout.Config{Timeout: 30 * time.Second})
	corsMw := cors.Middleware(cors.CORSOptions{})

	if err := c.Use(
		requestid.Middleware(),
		recoveryMw,
		corsMw,
		accesslogMw,
		compressionMw,
		timeoutMw,
	); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	store, err := storage.New(cfg.Storage, c.Logger())
	if err != nil {
		return nil, fmt.Errorf("construct store: %w", err)
	}

	if _, err := config.MaybeBootstrapEndpoints(ctx, store, config.BootstrapFilePath(), c.Logger()); err != nil {
		store.Close()
		return nil, fmt.Errorf("bootstrap endpoints: %w", err)
	}
	endpoints, externalEndpoints, err := store.ListEndpointConfigs(ctx)
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("list endpoint configs: %w", err)
	}
	cfg.Endpoints = endpoints
	cfg.ExternalEndpoints = externalEndpoints
	if err := config.FinalizeWithEndpoints(cfg, c.Logger()); err != nil {
		store.Close()
		return nil, fmt.Errorf("finalize endpoints: %w", err)
	}

	var m *metrics.Endpoints
	if cfg.Metrics {
		m = metrics.NewEndpoints(cfg.GetUniqueExtraMetricLabels())
	}

	wd := watchdog.New(cfg, store, m, c.Logger())

	var auth *security.BasicAuthenticator
	if cfg.Security != nil && cfg.Security.Basic != nil {
		auth = security.NewBasicAuthenticator(cfg.Security.Basic)
	}

	return &App{
		Core:     c,
		Cfg:      cfg,
		Store:    store,
		Metrics:  m,
		Watchdog: wd,
		Auth:     auth,
	}, nil
}

// Start prepares the runtime, launches the watchdog, then blocks while the
// HTTP server runs. When ctx is canceled, it triggers a graceful shutdown
// bounded at 15 s, plus an additional drain for watchdog probes and the
// storage layer.
func (a *App) Start(ctx context.Context) error {
	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	a.Core.Logger().Info("starting guardus", plumelog.Fields{
		"addr":    a.Cfg.Core.Addr,
		"version": a.Cfg.Version,
	})

	a.Watchdog.Monitor(ctx)

	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		shutdownErr <- a.Core.Shutdown(shutdownCtx)
	}()

	var serveErr error
	if a.Cfg.Core.TLS.Enabled {
		serveErr = srv.ListenAndServeTLS("", "")
	} else {
		serveErr = srv.ListenAndServe()
	}
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		a.releaseRuntime()
		return fmt.Errorf("server stopped: %w", serveErr)
	}
	if err := <-shutdownErr; err != nil {
		a.releaseRuntime()
		return fmt.Errorf("shutdown server: %w", err)
	}
	a.releaseRuntime()
	return nil
}

// releaseRuntime stops the watchdog and closes the store. Idempotent.
func (a *App) releaseRuntime() {
	if a.Watchdog != nil {
		a.Watchdog.Shutdown()
	}
	if a.Store != nil {
		if err := a.Store.Save(); err != nil {
			a.Core.Logger().Warn("shutdown: store save failed", plumelog.Fields{"error": err.Error()})
		}
		a.Store.Close()
	}
}
