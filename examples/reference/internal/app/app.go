// Package app wires together the application dependencies and manages the
// server lifecycle.
package app

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/examples/reference/internal/config"
	"github.com/spcent/plumego/metrics"
	webhookout "github.com/spcent/plumego/net/webhookout"
	"github.com/spcent/plumego/pubsub"
)

// App holds application-wide dependencies.
type App struct {
	Core       *core.App
	Cfg        config.Config
	StaticFS   fs.FS
	Bus        *pubsub.InProcPubSub
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
		core.WithRequestID(),
		core.WithRecovery(),
		core.WithLogging(),
		core.WithCORS(),
		core.WithPubSub(bus),
		core.WithWebhookIn(core.WebhookInConfig{
			Enabled:           cfg.EnableWebhooks,
			Pub:               bus,
			GitHubSecret:      cfg.GitHubSecret,
			StripeSecret:      cfg.StripeSecret,
			MaxBodyBytes:      1 << 20,
			StripeTolerance:   5 * time.Minute,
			TopicPrefixGitHub: "in.github.",
			TopicPrefixStripe: "in.stripe.",
		}),
		core.WithWebhookOut(core.WebhookOutConfig{
			Enabled:          cfg.EnableWebhooks,
			Service:          webhookSvc,
			TriggerToken:     cfg.WebhookToken,
			BasePath:         "/webhooks",
			IncludeStats:     true,
			DefaultPageLimit: 50,
		}),
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

	return &App{
		Core:       core.New(opts...),
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
