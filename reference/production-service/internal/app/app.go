// Package app wires the production reference service.
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
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/ratelimit"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	securitymw "github.com/spcent/plumego/middleware/security"
	"github.com/spcent/plumego/middleware/timeout"
	"github.com/spcent/plumego/middleware/tracing"
	"production-service/internal/config"
	"production-service/internal/domain/tenant"
)

// App holds application-wide dependencies.
type App struct {
	Core      *core.App
	Cfg       config.Config
	Metrics   *metrics.BaseMetricsCollector
	Profiles  *tenant.Store
	RateLimit *ratelimit.AbuseGuardMiddleware
}

// New constructs the production reference with explicit middleware wiring.
func New(cfg config.Config) (*App, error) {
	logger := plumelog.NewLogger()
	collector := metrics.NewBaseMetricsCollector()
	profiles, err := tenant.NewStore(cfg.App.ProfileStorePath)
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

	// Middleware order — outermost to innermost (first registered runs first on inbound):
	//   requestid    → stamps correlation ID before any logging or error handling
	//   recovery     → converts panics to 500; inside requestid so ID appears in the response
	//   bodylimit    → rejects oversized bodies before auth or rate-limit processing
	//   timeout      → enforces per-request wall-clock limit on everything below
	//   security     → adds security headers to all responses, including timed-out ones
	//   ratelimit    → rejects abusive clients; after security so headers are always set
	//   tracing      → hooks span lifecycle after rejection decisions are made
	//   httpmetrics  → records request stats including rejected and timed-out requests
	//   accesslog    → innermost: logs the final status with full middleware context
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
// When ctx is canceled, it triggers a graceful shutdown.
func (a *App) Start(ctx context.Context) error {
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
	a.Core.Logger().Info("starting server", plumelog.Fields{
		"addr": a.Cfg.Core.Addr,
		"tls":  a.Cfg.Core.TLS.Enabled,
	})
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
		// Empty cert/key paths rely on core.Server() having loaded
		// cfg.Core.TLS.CertFile and cfg.Core.TLS.KeyFile into srv.TLSConfig.
		// In most deployments TLS is terminated by the proxy; set those fields
		// in config.go before enabling self-terminating TLS.
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

type noopTracer struct{}

func (noopTracer) Start(ctx context.Context, r *http.Request) (context.Context, tracing.TraceSpan) {
	return ctx, noopSpan{}
}

type noopSpan struct{}

func (noopSpan) End(status, bytes int, requestID string) {}
func (noopSpan) TraceID() string                         { return "" }
func (noopSpan) SpanID() string                          { return "" }
