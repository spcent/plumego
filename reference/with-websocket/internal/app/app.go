// Package app wires the with-websocket demo dependencies.
// Non-canonical: this demo extends the standard-service layout with x/websocket.
package app

import (
	"context"
	"fmt"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/reference/with-websocket/internal/config"
	"github.com/spcent/plumego/x/websocket"
)

// App holds application-wide dependencies including the WebSocket server.
type App struct {
	Core *core.App
	Cfg  config.Config
	WS   *websocket.Server
}

// New constructs the App with a WebSocket server.
func New(cfg config.Config) (*App, error) {
	a := core.New(cfg.Core, core.WithLogger(plumelog.NewGLogger()))
	a.Use(requestid.Middleware())
	a.Use(recovery.Recovery(a.Logger()))
	a.Use(accesslog.Middleware(a.Logger()))

	wsCfg := websocket.DefaultWebSocketConfig()
	wsCfg.Secret = []byte(cfg.WSSecret)

	ws, err := websocket.New(wsCfg, cfg.Core.Debug, a.Logger())
	if err != nil {
		return nil, fmt.Errorf("create websocket server: %w", err)
	}

	return &App{Core: a, Cfg: cfg, WS: ws}, nil
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
	defer func() {
		_ = a.WS.Shutdown(ctx)
		a.Core.Shutdown(ctx)
	}()

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
