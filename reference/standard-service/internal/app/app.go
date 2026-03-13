// Package app wires together the application dependencies and manages the
// server lifecycle.
package app

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/health"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
	"github.com/spcent/plumego/reference/standard-service/internal/config"
	"github.com/spcent/plumego/x/devtools/pubsubdebug"
	"github.com/spcent/plumego/x/pubsub"
	"github.com/spcent/plumego/x/webhook"
	xwebsocket "github.com/spcent/plumego/x/websocket"
)

// App holds application-wide dependencies.
type App struct {
	Core       *core.App
	Cfg        config.Config
	StaticFS   fs.FS
	Bus        *pubsub.InProcBroker
	WebhookSvc *webhook.Service
	Prom       *metrics.PrometheusCollector
	Tracer     *metrics.OpenTelemetryTracer
	Health     health.HealthManager
}

// New constructs the App and initialises all infrastructure dependencies.
// staticFS must expose a "templates/" subtree for HTML template parsing.
func New(cfg config.Config, staticFS fs.FS) (*App, error) {
	bus := pubsub.New()

	healthManager, err := health.NewHealthManager(health.HealthCheckConfig{})
	if err != nil {
		return nil, fmt.Errorf("create health manager: %w", err)
	}

	webhookStore := webhook.NewMemStore()
	webhookCfg := webhook.ConfigFromEnv()
	webhookCfg.Enabled = cfg.EnableWebhooks
	webhookSvc := webhook.NewService(webhookStore, webhookCfg)

	var prom *metrics.PrometheusCollector
	var tracer *metrics.OpenTelemetryTracer
	if cfg.EnableMetrics {
		prom = metrics.NewPrometheusCollector("plumego_reference")
		tracer = metrics.NewOpenTelemetryTracer("plumego-reference")
	}

	opts := []core.Option{
		core.WithAddr(cfg.Core.Addr),
		core.WithEnvPath(cfg.Core.EnvFile),
		core.WithHealthManager(healthManager),
		core.WithLogger(plumelog.NewGLogger()),
	}
	if cfg.Core.Debug {
		opts = append(opts, core.WithDebug())
	}
	if prom != nil {
		opts = append(opts, core.WithPrometheusCollector(prom))
	}
	if tracer != nil {
		opts = append(opts, core.WithTracer(tracer))
	}

	// Mount extension components via WithComponent
	if cfg.EnableWebhooks {
		opts = append(opts,
			core.WithComponent(pubsubdebug.NewPubSubDebugComponent(pubsubdebug.PubSubConfig{
				Enabled: true,
				Path:    "/debug/pubsub",
			}, bus)),
			core.WithComponent(webhook.NewWebhookInComponent(webhook.WebhookInConfig{
				Enabled:           true,
				Pub:               bus,
				GitHubSecret:      cfg.GitHubSecret,
				StripeSecret:      cfg.StripeSecret,
				MaxBodyBytes:      1 << 20,
				StripeTolerance:   5 * time.Minute,
				TopicPrefixGitHub: "in.github.",
				TopicPrefixStripe: "in.stripe.",
			}, bus, nil)),
			core.WithComponent(webhook.NewWebhookOutComponent(webhook.WebhookOutConfig{
				Enabled:          true,
				Service:          webhookSvc,
				TriggerToken:     cfg.WebhookToken,
				BasePath:         "/webhooks",
				IncludeStats:     true,
				DefaultPageLimit: 50,
			})),
		)
	}

	app := core.New(opts...)
	app.Use(requestid.Middleware())
	app.Use(mwtracing.Middleware(tracer))
	app.Use(httpmetrics.Middleware(app.HTTPMetrics()))
	app.Use(accesslog.Middleware(app.Logger()))
	app.Use(recovery.Recovery(app.Logger()))
	app.Use(cors.CORS)

	wsCfg := xwebsocket.DefaultWebSocketConfig()
	wsCfg.Secret = []byte(cfg.WebSocketSecret)
	if comp, err := xwebsocket.NewComponent(wsCfg, cfg.Core.Debug, app.Logger()); err == nil {
		if err := app.MountComponent(comp); err != nil {
			return nil, fmt.Errorf("mount websocket component: %w", err)
		}
	} else {
		app.Logger().Warn("WebSocket disabled", plumelog.Fields{"error": err})
	}

	return &App{
		Core:       app,
		Cfg:        cfg,
		StaticFS:   staticFS,
		Bus:        bus,
		WebhookSvc: webhookSvc,
		Prom:       prom,
		Tracer:     tracer,
		Health:     healthManager,
	}, nil
}

// Start launches the webhook service and blocks while the HTTP server runs.
func (a *App) Start() error {
	ctx := context.Background()

	if a.WebhookSvc != nil {
		a.WebhookSvc.Start(ctx)
		defer a.WebhookSvc.Stop()
	}
	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	if err := a.Core.Start(ctx); err != nil {
		return fmt.Errorf("start runtime: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}
	defer a.Core.Shutdown(ctx)

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("server stopped: %w", err)
	}
	return nil
}
