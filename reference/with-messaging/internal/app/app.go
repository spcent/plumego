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
	"github.com/spcent/plumego/x/messaging"
	"with-messaging/internal/config"
	"with-messaging/internal/handler"
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
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: a.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	accesslogMw, err := accesslog.Middleware(accesslog.Config{Logger: a.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure access log middleware: %w", err)
	}
	if err := a.Use(
		requestid.Middleware(),
		recoveryMw,
		accesslogMw,
	); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

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
	defer func() {
		_ = a.Core.Shutdown(ctx)
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
