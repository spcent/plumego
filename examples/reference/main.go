package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spcent/plumego/core"
	"github.com/spcent/plumego/frontend"
	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/metrics"
	webhookout "github.com/spcent/plumego/net/webhookout"
	"github.com/spcent/plumego/pubsub"
)

//go:embed ui/*
var staticFS embed.FS

func main() {
	// Optional in-process pub/sub powers inbound webhook fan-out and WebSocket broadcasts.
	bus := pubsub.New()

	// Outbound webhook management runs alongside the HTTP server.
	webhookStore := webhookout.NewMemStore()
	webhookCfg := webhookout.ConfigFromEnv()
	webhookCfg.Enabled = true
	webhookSvc := webhookout.NewService(webhookStore, webhookCfg)

	// Prometheus + OpenTelemetry hooks plugged into logging middleware.
	prom := metrics.NewPrometheusCollector("plumego_example")
	tracer := metrics.NewOpenTelemetryTracer("plumego-example")

	app := core.New(
		core.WithAddr(":8080"),
		core.WithDebug(),
		core.WithPubSub(bus),
		core.WithMetricsCollector(prom),
		core.WithTracer(tracer),
		core.WithWebhookIn(core.WebhookInConfig{
			Enabled:           true,
			Pub:               bus,
			GitHubSecret:      envOr("GITHUB_WEBHOOK_SECRET", "dev-github-secret"),
			StripeSecret:      envOr("STRIPE_WEBHOOK_SECRET", "whsec_dev"),
			MaxBodyBytes:      1 << 20,
			StripeTolerance:   5 * time.Minute,
			TopicPrefixGitHub: "in.github.",
			TopicPrefixStripe: "in.stripe.",
		}),
		core.WithWebhookOut(core.WebhookOutConfig{
			Enabled:          true,
			Service:          webhookSvc,
			TriggerToken:     envOr("WEBHOOK_TRIGGER_TOKEN", "dev-trigger"),
			BasePath:         "/webhooks",
			IncludeStats:     true,
			DefaultPageLimit: 50,
		}),
	)

	app.EnableRecovery()
	app.EnableLogging()
	app.EnableCORS()

	// Static frontend served from the embedded UI folder.
	_ = frontend.RegisterFS(app.Router(), http.FS(staticFS), frontend.WithPrefix("/"))

	// Minimal health endpoints for orchestration hooks.
	app.Get("/health/ready", health.ReadinessHandler().ServeHTTP)
	app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)

	// WebSocket hub with broadcast endpoint and simple echoing demo.
	wsCfg := core.DefaultWebSocketConfig()
	wsCfg.Secret = []byte(envOr("WS_SECRET", "dev-secret"))
	_, err := app.ConfigureWebSocketWithOptions(wsCfg)
	if err != nil {
		log.Fatalf("configure websocket: %v", err)
	}
	webhookSvc.Start(context.Background())
	defer webhookSvc.Stop()

	// Example API route demonstrating middleware and tracing hooks.
	app.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Served-At", time.Now().Format(time.RFC3339))
		w.Write([]byte("hello from plumego"))
	})

	// Expose metrics for scraping.
	app.Get("/metrics", prom.Handler().ServeHTTP)

	if err := app.Boot(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
