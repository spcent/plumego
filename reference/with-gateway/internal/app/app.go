// Package app wires the with-gateway demo dependencies.
// Non-canonical: this demo extends the standard-service layout with x/gateway.
package app

import (
	"context"
	"fmt"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/reference/with-gateway/internal/config"
	"github.com/spcent/plumego/x/gateway"
)

// App holds application-wide dependencies including the reverse proxy.
type App struct {
	Core  *core.App
	Cfg   config.Config
	Proxy *gateway.GatewayProxy
}

// New constructs the App with a gateway reverse proxy.
func New(cfg config.Config) (*App, error) {
	a := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})
	a.Use(requestid.Middleware())
	a.Use(recovery.Recovery(a.Logger()))
	a.Use(accesslog.Logging(a.Logger(), nil, nil))

	proxy := gateway.NewGateway(gateway.GatewayConfig{
		Targets: []string{cfg.GatewayBackend},
	})

	return &App{Core: a, Cfg: cfg, Proxy: proxy}, nil
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

	if a.Cfg.Core.TLS.Enabled {
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			return fmt.Errorf("server stopped: %w", err)
		}
		return nil
	}
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("server stopped: %w", err)
	}
	return nil
}
