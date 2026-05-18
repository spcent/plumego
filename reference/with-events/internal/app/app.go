// Package app wires the with-events scenario dependencies.
package app

import (
	"context"
	"fmt"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/reference/with-events/internal/config"
	"github.com/spcent/plumego/reference/with-events/internal/order"
	"github.com/spcent/plumego/x/messaging"
)

// Deps contains optional dependencies supplied by tests or embedders.
type Deps struct {
	Logger plumelog.StructuredLogger
	Bus    *messaging.Broker
}

// App holds application-wide dependencies for the event-driven scenario.
type App struct {
	Core      *core.App
	Cfg       config.Config
	Logger    plumelog.StructuredLogger
	Bus       *messaging.Broker
	Messaging *messaging.Service
	Orders    *order.Handler
	Consumer  *order.OrderConsumer
}

// New constructs the App with an in-process messaging service.
func New(cfg config.Config, deps Deps) (*App, error) {
	logger := deps.Logger
	if logger == nil {
		logger = plumelog.NewLogger()
	}

	coreCfg := core.DefaultConfig()
	coreCfg.Addr = cfg.Addr
	a := core.New(coreCfg, core.AppDependencies{Logger: logger})
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: logger})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	if err := a.Use(requestid.Middleware(), recoveryMw); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	bus := deps.Bus
	if bus == nil {
		bus = messaging.NewInProcBroker()
	}
	svc := messaging.New(messaging.Config{
		Bus:        bus,
		Logger:     logger,
		ConsumerID: "with-events",
	})
	publisher := order.NewPublisher(bus, order.NewMemoryIdempotencyStore())
	consumer := order.NewConsumer(bus, order.NewMemoryIdempotencyStore())

	return &App{
		Core:      a,
		Cfg:       cfg,
		Logger:    logger,
		Bus:       bus,
		Messaging: svc,
		Orders:    order.NewHandler(publisher),
		Consumer:  consumer,
	}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
func (a *App) Start(ctx context.Context) error {
	if a.Messaging != nil {
		if err := a.Messaging.Start(ctx); err != nil {
			return fmt.Errorf("start messaging service: %w", err)
		}
		defer func() {
			_ = a.Messaging.Stop(ctx)
		}()
	}
	if a.Consumer != nil {
		if err := a.Consumer.Start(ctx); err != nil {
			return fmt.Errorf("start order consumer: %w", err)
		}
	}
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
