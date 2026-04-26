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
	"github.com/spcent/plumego/reference/production-service/internal/config"
)

// App holds application-wide dependencies.
type App struct {
	Core    *core.App
	Cfg     config.Config
	Metrics *metrics.BaseMetricsCollector
}

// New constructs the production reference with explicit middleware wiring.
func New(cfg config.Config) (*App, error) {
	logger := plumelog.NewLogger()
	collector := metrics.NewBaseMetricsCollector()
	app := core.New(cfg.Core, core.AppDependencies{Logger: logger})

	if err := app.Use(
		requestid.Middleware(),
		recovery.Recovery(app.Logger()),
		bodylimit.BodyLimit(cfg.App.BodyLimitBytes, app.Logger()),
		timeout.Timeout(timeout.TimeoutConfig{Timeout: cfg.App.RequestTimeout}),
		securitymw.SecurityHeaders(nil),
		ratelimit.AbuseGuard(ratelimit.AbuseGuardConfig{
			Rate:     cfg.App.RateLimit,
			Capacity: cfg.App.RateBurst,
			Logger:   app.Logger(),
		}),
		tracing.Middleware(noopTracer{}),
		httpmetrics.Middleware(collector),
		accesslog.Middleware(app.Logger(), nil, nil),
	); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	return &App{
		Core:    app,
		Cfg:     cfg,
		Metrics: collector,
	}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
func (a *App) Start() error {
	ctx := context.Background()

	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}
	defer a.Core.Shutdown(ctx)

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
