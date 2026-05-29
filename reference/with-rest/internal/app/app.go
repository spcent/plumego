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
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/securityheaders"
	"github.com/spcent/plumego/middleware/timeout"
	"github.com/spcent/plumego/x/rest"
	"with-rest/internal/config"
	"with-rest/internal/domain/user"
	"with-rest/internal/handler"
	"with-rest/internal/validation/playground"
)

// App holds application-wide dependencies.
type App struct {
	Core  *core.App
	Cfg   config.Config
	Users *rest.DBResourceController[user.User]
	Items *handler.CreateItem
}

// New constructs the App with all dependencies wired explicitly.
// The middleware stack mirrors standard-service so x/rest integrations
// run with the full production-oriented middleware baseline.
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

	// Middleware order matches standard-service (outermost to innermost):
	//   requestid  → stamps correlation ID before any logging or error handling
	//   security   → security headers on all responses
	//   cors       → CORS preflight and headers; CORSOptions{} allows all origins (dev default)
	//   recovery   → converts panics to 500 responses
	//   accesslog  → logs all requests including 413; outer to bodylimit so rejections are logged
	//   bodylimit  → rejects oversized bodies with 413
	//   httpmetrics→ noop collector; swap for observability.NewPrometheusCollector in production
	//   timeout    → per-request wall-clock limit; innermost so only handler time is counted
	if err := app.Use(
		requestid.Middleware(),
		securityMw,
		cors.Middleware(cors.CORSOptions{}),
		recoveryMw,
		accesslogMw,
		bodylimit.Middleware(bodylimit.Config{
			MaxBytes: 1 << 20, // 1 MiB default; make configurable via AppConfig if needed
			Logger:   app.Logger(),
		}),
		httpmetrics.Middleware(metrics.NewNoopCollector()),
		timeout.Middleware(timeout.Config{Timeout: 30 * time.Second}),
	); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	spec := rest.DefaultResourceSpec("users").WithPrefix("/api/users")
	users := rest.NewDBResource[user.User](spec, user.NewRepository())

	return &App{
		Core:  app,
		Cfg:   cfg,
		Users: users,
		Items: handler.NewCreateItem(playground.NewValidator(), app.Logger()),
	}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
// Shutdown is driven by ctx so the application owner controls process signals.
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
	})
	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		shutdownErr <- a.Core.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server stopped: %w", err)
	}
	if err := <-shutdownErr; err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
