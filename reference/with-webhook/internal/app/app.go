// Package app wires the with-webhook demo dependencies.
// Non-canonical: this demo extends the standard-service layout with x/webhook.
package app

import (
	"context"
	"fmt"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/reference/with-webhook/internal/config"
	"github.com/spcent/plumego/x/pubsub"
	"github.com/spcent/plumego/x/webhook"
)

// App holds application-wide dependencies including the inbound webhook receiver.
type App struct {
	Core    *core.App
	Cfg     config.Config
	Inbound *webhook.Inbound
}

// New constructs the App with an inbound webhook receiver backed by an in-process broker.
func New(cfg config.Config) (*App, error) {
	opts := []core.Option{
		core.WithAddr(cfg.Core.Addr),
		core.WithEnvPath(cfg.Core.EnvFile),
		core.WithLogger(plumelog.NewGLogger()),
	}
	if cfg.Core.Debug {
		opts = append(opts, core.WithDebug())
	}

	a := core.New(opts...)
	a.Use(requestid.Middleware())
	a.Use(recovery.Recovery(a.Logger()))
	a.Use(accesslog.Middleware(a.Logger()))

	broker := pubsub.New()
	inbound := webhook.NewInbound(
		webhook.WebhookInConfig{Enabled: true},
		broker,
		a.Logger(),
	)

	return &App{Core: a, Cfg: cfg, Inbound: inbound}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
func (a *App) Start() error {
	ctx := context.Background()

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
