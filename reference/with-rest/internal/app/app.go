// Package app wires together the application dependencies and manages the
// server lifecycle.
package app

import (
	"context"
	"fmt"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/reference/with-rest/internal/config"
	"github.com/spcent/plumego/reference/with-rest/internal/domain/user"
	"github.com/spcent/plumego/reference/with-rest/internal/handler"
	"github.com/spcent/plumego/reference/with-rest/internal/validation/playground"
	"github.com/spcent/plumego/x/rest"
)

// App holds application-wide dependencies.
type App struct {
	Core  *core.App
	Cfg   config.Config
	Users *rest.DBResourceController[user.User]
	Items *handler.CreateItem
}

// New constructs the App with all dependencies wired explicitly.
func New(cfg config.Config) (*App, error) {
	app := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: app.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	if err := app.Use(requestid.Middleware(), recoveryMw); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	spec := rest.DefaultResourceSpec("users").WithPrefix("/api/users")
	users := rest.NewDBResource[user.User](spec, user.NewRepository())

	return &App{
		Core:  app,
		Cfg:   cfg,
		Users: users,
		Items: handler.NewCreateItem(playground.NewValidator()),
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

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("server stopped: %w", err)
	}
	return nil
}
