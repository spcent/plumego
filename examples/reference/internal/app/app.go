// Package app wires together the application dependencies and manages the
// server lifecycle.
package app

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/core/components/pubsubdebug"
	"github.com/spcent/plumego/core/components/webhook"
	"github.com/spcent/plumego/examples/reference/internal/config"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/observability"
	"github.com/spcent/plumego/middleware/recovery"
	webhookout "github.com/spcent/plumego/net/webhookout"
	"github.com/spcent/plumego/pubsub"
)

// App holds application-wide dependencies.
type App struct {
	Core       *core.App
	Cfg        config.Config
	StaticFS   fs.FS
	Bus        *pubsub.InProcBroker
	WebhookSvc *webhookout.Service
	Prom       *metrics.PrometheusCollector
	Tracer     *metrics.OpenTelemetryTracer
}

// New constructs the App and initialises all infrastructure dependencies.
// staticFS must expose a "templates/" subtree for HTML template parsing.
func New(cfg config.Config, staticFS fs.FS) (*App, error) {
	bus := pubsub.New()

	webhookStore := webhookout.NewMemStore()
	webhookCfg := webhookout.ConfigFromEnv()
	webhookCfg.Enabled = cfg.EnableWebhooks
	webhookSvc := webhookout.NewService(webhookStore, webhookCfg)

	var prom *metrics.PrometheusCollector
	var tracer *metrics.OpenTelemetryTracer
	if cfg.EnableMetrics {
		prom = metrics.NewPrometheusCollector("plumego_reference")
		tracer = metrics.NewOpenTelemetryTracer("plumego-reference")
	}

	opts := []core.Option{
		core.WithAddr(cfg.Core.Addr),
		core.WithEnvPath(cfg.Core.EnvFile),
	}
	if cfg.Core.Debug {
		opts = append(opts, core.WithDebug())
	}
	if prom != nil {
		opts = append(opts, core.WithMetricsCollector(prom))
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
	app.Use(recovery.RecoveryMiddleware)
	app.Use(observability.RequestID())
	app.Use(cors.CORS)
	app.Use(observability.Logging(nil, nil, nil))

	return &App{
		Core:       app,
		Cfg:        cfg,
		StaticFS:   staticFS,
		Bus:        bus,
		WebhookSvc: webhookSvc,
		Prom:       prom,
		Tracer:     tracer,
	}, nil
}

// Start launches the webhook service and blocks while the HTTP server runs.
func (a *App) Start() error {
	if a.WebhookSvc != nil {
		a.WebhookSvc.Start(context.Background())
		defer a.WebhookSvc.Stop()
	}
	if err := a.Core.Boot(); err != nil {
		return fmt.Errorf("server stopped: %w", err)
	}
	return nil
}
