// Package app wires application dependencies and manages the server lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/securityheaders"
	"github.com/spcent/plumego/middleware/timeout"
	"mini-saas-api/internal/config"
)

// App holds application-wide dependencies.
type App struct {
	Core *core.App
	Cfg  config.Config
}

// New constructs the App with explicit stable-root wiring only.
// Extension wiring (x/tenant, x/observability) is added per-route or per-card;
// all global middleware here is stdlib-only stable-root.
func New(cfg config.Config) (*App, error) {
	app := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})

	securityMw, err := securityheaders.Middleware(securityheaders.Config{})
	if err != nil {
		return nil, fmt.Errorf("configure security headers middleware: %w", err)
	}
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: app.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	accesslogMw, err := accesslog.Middleware(accesslog.Config{Logger: app.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure access log middleware: %w", err)
	}
	timeoutMw := timeout.Middleware(timeout.Config{Timeout: 30 * time.Second})

	var corsOpts cors.CORSOptions
	if len(cfg.App.CORSAllowedOrigins) > 0 {
		strictOpts, err := cors.StrictDefaultOptions(cfg.App.CORSAllowedOrigins...)
		if err != nil {
			return nil, fmt.Errorf("configure CORS middleware: %w", err)
		}
		corsOpts = strictOpts
	}

	// Middleware order — outermost to innermost:
	//   requestid  → stamps correlation ID before any logging or error handling
	//   security   → security headers on all responses
	//   cors       → CORS preflight; set APP_CORS_ALLOWED_ORIGINS in production
	//   recovery   → converts panics to 500; inside cors/security so headers apply
	//   accesslog  → logs every request; after recovery so panics appear as 500
	//   bodylimit  → rejects oversized bodies; after accesslog so the 413 is logged
	//   httpmetrics→ measures handler latency and status; swap NewNoopCollector for
	//               observability.NewPrometheusCollector (card 1532, x/observability)
	//   timeout    → per-request wall-clock limit; innermost so only handler time counts
	if err := app.Use(
		requestid.Middleware(),
		securityMw,
		cors.Middleware(corsOpts),
		recoveryMw,
		accesslogMw,
		bodylimit.Middleware(bodylimit.Config{
			MaxBytes: cfg.App.MaxBodyBytes,
			Logger:   app.Logger(),
		}),
		httpmetrics.Middleware(metrics.NewNoopCollector()),
		timeoutMw,
	); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	if cfg.App.JWTSecret == config.DevJWTSecret {
		app.Logger().Warn("using dev JWT secret — set APP_JWT_SECRET in production", plumelog.Fields{})
	}
	if cfg.App.MetricsToken == "" {
		app.Logger().Warn("metrics endpoint is unprotected — set APP_METRICS_TOKEN in production", plumelog.Fields{})
	}
	if len(cfg.App.CORSAllowedOrigins) == 0 {
		app.Logger().Warn("CORS allows all origins (*); set APP_CORS_ALLOWED_ORIGINS in production", plumelog.Fields{})
	}

	return &App{Core: app, Cfg: cfg}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
// When ctx is canceled it triggers graceful shutdown.
func (a *App) Start(ctx context.Context) error {
	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	a.Core.Logger().Info("starting server", plumelog.Fields{
		"addr": a.Cfg.Core.Addr,
		"tls":  a.Cfg.Core.TLS.Enabled,
	})

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
		return fmt.Errorf("server stopped: %w", serveErr)
	}
	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
