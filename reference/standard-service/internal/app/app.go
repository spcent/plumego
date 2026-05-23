// Package app wires together the application dependencies and manages the
// server lifecycle.
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
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/timeout"
	"standard-service/internal/config"
)

// App holds application-wide dependencies.
type App struct {
	Core *core.App
	Cfg  config.Config
}

// New constructs the App with explicit stable-root wiring only.
func New(cfg config.Config) (*App, error) {
	app := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: app.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	accesslogMw, err := accesslog.Middleware(accesslog.Config{Logger: app.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure access log middleware: %w", err)
	}
	timeoutMw := timeout.Middleware(timeout.Config{Timeout: 30 * time.Second})
	// Middleware order — outermost to innermost (first registered runs first on inbound requests):
	//   requestid  → stamps correlation ID before any logging or error handling
	//   recovery   → converts panics to 500 responses; runs inside requestid so ID is in the response
	//   accesslog  → logs every request/response at transport level; runs after requestid and recovery
	//   bodylimit  → rejects oversized bodies with 413; placed after accesslog so the 413 is logged
	//   timeout    → enforces per-request wall-clock limit; innermost so only handler time is counted
	if err := app.Use(
		requestid.Middleware(),
		recoveryMw,
		accesslogMw,
		bodylimit.Middleware(bodylimit.Config{
			MaxBytes: cfg.App.MaxBodyBytes,
			Logger:   app.Logger(),
		}),
		timeoutMw,
	); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	return &App{
		Core: app,
		Cfg:  cfg,
	}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
// When ctx is canceled, it triggers a graceful shutdown.
func (a *App) Start(ctx context.Context) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()
		// Allow up to 15 s for in-flight requests to complete before forcing close.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		shutdownErr <- a.Core.Shutdown(shutdownCtx)
	}()

	var serveErr error
	if a.Cfg.Core.TLS.Enabled {
		// Empty cert/key paths rely on core.Server() having already loaded
		// cfg.Core.TLS.CertFile and cfg.Core.TLS.KeyFile into srv.TLSConfig.
		// Set those fields (and optionally cfg.Core.TLS.ClientAuth) in config.go
		// before enabling TLS. In most deployments TLS is terminated by the
		// proxy and this branch is not reached.
		serveErr = srv.ListenAndServeTLS("", "")
	} else {
		serveErr = srv.ListenAndServe()
	}
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return fmt.Errorf("server stopped: %w", serveErr)
	}
	// Always drain the shutdown channel so the goroutine is not leaked and
	// shutdown errors are not silently discarded.
	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
