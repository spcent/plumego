// Package app wires the with-messaging demo dependencies.
// Non-canonical: this demo extends the standard-service layout with x/messaging.
package app

import (
	"context"
	"fmt"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/reference/with-messaging/internal/config"
	"github.com/spcent/plumego/reference/with-messaging/internal/handler"
	"github.com/spcent/plumego/x/messaging"
)

// App holds application-wide dependencies including the messaging broker.
type App struct {
	Core    *core.App
	Cfg     config.Config
	Handler handler.MessagingHandler
}

// New constructs the App with an in-process messaging broker.
func New(cfg config.Config) (*App, error) {
	a := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})
	a.Use(requestid.Middleware())
	a.Use(recovery.Recovery(a.Logger()))
	a.Use(accesslog.Middleware(a.Logger()))

	broker := messaging.NewInProcBroker()

	return &App{
		Core:    a,
		Cfg:     cfg,
		Handler: handler.MessagingHandler{Broker: broker},
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
