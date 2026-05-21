// Package app wires the production reference service.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/ratelimit"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	securitymw "github.com/spcent/plumego/middleware/security"
	"github.com/spcent/plumego/middleware/timeout"
	"github.com/spcent/plumego/middleware/tracing"
	"production-service/internal/config"
)

// App holds application-wide dependencies.
type App struct {
	Core      *core.App
	Cfg       config.Config
	Metrics   *metrics.BaseMetricsCollector
	Profiles  *profileStore
	RateLimit *ratelimit.AbuseGuardMiddleware
}

// New constructs the production reference with explicit middleware wiring.
func New(cfg config.Config) (*App, error) {
	logger := plumelog.NewLogger()
	collector := metrics.NewBaseMetricsCollector()
	profiles, err := newProfileStore(cfg.App.ProfileStorePath)
	if err != nil {
		return nil, fmt.Errorf("load profile store: %w", err)
	}
	app := core.New(cfg.Core, core.AppDependencies{Logger: logger})
	rateLimitGuard := ratelimit.NewAbuseGuard(ratelimit.AbuseGuardConfig{
		Rate:     cfg.App.RateLimit,
		Capacity: cfg.App.RateBurst,
		Logger:   app.Logger(),
	})
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: app.Logger()})
	if err != nil {
		rateLimitGuard.Stop()
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	accesslogMw, err := accesslog.Middleware(accesslog.Config{Logger: app.Logger()})
	if err != nil {
		rateLimitGuard.Stop()
		return nil, fmt.Errorf("configure access log middleware: %w", err)
	}
	securityMw, err := securitymw.Middleware(securitymw.Config{})
	if err != nil {
		rateLimitGuard.Stop()
		return nil, fmt.Errorf("configure security middleware: %w", err)
	}

	if err := app.Use(
		requestid.Middleware(),
		recoveryMw,
		bodylimit.Middleware(bodylimit.Config{MaxBytes: cfg.App.BodyLimitBytes, Logger: app.Logger()}),
		timeout.Middleware(timeout.Config{Timeout: cfg.App.RequestTimeout}),
		securityMw,
		rateLimitGuard.Middleware(),
		tracing.Middleware(noopTracer{}),
		httpmetrics.Middleware(collector),
		accesslogMw,
	); err != nil {
		rateLimitGuard.Stop()
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	return &App{
		Core:      app,
		Cfg:       cfg,
		Metrics:   collector,
		Profiles:  profiles,
		RateLimit: rateLimitGuard,
	}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
// It listens for SIGTERM and SIGINT and triggers a graceful shutdown.
func (a *App) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if a.RateLimit != nil {
		defer a.RateLimit.Stop()
	}

	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}
	go func() {
		<-ctx.Done()
		_ = a.Core.Shutdown(context.Background())
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
	return nil
}

type noopTracer struct{}

func (noopTracer) Start(ctx context.Context, r *http.Request) (context.Context, tracing.TraceSpan) {
	return ctx, noopSpan{}
}

type noopSpan struct{}

func (noopSpan) End(status, bytes int, requestID string) {}
func (noopSpan) TraceID() string                         { return "" }
func (noopSpan) SpanID() string                          { return "" }

func utcNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}
