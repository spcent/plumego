package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/timeout"
)

type RouteDependencies struct {
	Service    *Service
	Ready      func(context.Context) error
	Metrics    http.HandlerFunc
	WorkerAuth WorkerIngressAuthConfig
	AdminAuth  AdminAuthConfig
}

type RouteRegistrar func(app *core.App, deps RouteDependencies) error

type App struct {
	core    *core.App
	config  Config
	runtime *Runtime
	server  ServerConfig
}

func New(ctx context.Context, cfg Config, server ServerConfig, registerRoutes RouteRegistrar) (*App, error) {
	if registerRoutes == nil {
		return nil, fmt.Errorf("workerfleet route registrar is required")
	}
	runtime, err := Bootstrap(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap workerfleet: %w", err)
	}

	coreApp, err := newCoreApp(server)
	if err != nil {
		_ = runtime.Close(context.Background())
		return nil, err
	}
	routeDeps := RouteDependencies{
		Service:    runtime.Service,
		Ready:      runtime.Ready,
		Metrics:    runtime.Metrics.Handler(),
		WorkerAuth: cfg.WorkerAuth,
		AdminAuth:  cfg.AdminAuth,
	}
	if err := registerRoutes(coreApp, routeDeps); err != nil {
		_ = runtime.Close(context.Background())
		return nil, fmt.Errorf("register routes: %w", err)
	}

	return &App{
		core:    coreApp,
		config:  cfg,
		runtime: runtime,
		server:  server,
	}, nil
}

func newCoreApp(server ServerConfig) (*core.App, error) {
	logger := plumelog.NewLogger()
	coreApp := core.New(server.Core, core.AppDependencies{Logger: logger})
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: coreApp.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	accesslogMw, err := accesslog.Middleware(accesslog.Config{Logger: coreApp.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure access log middleware: %w", err)
	}
	if err := coreApp.Use(
		requestid.Middleware(),
		recoveryMw,
		accesslogMw,
		bodylimit.Middleware(bodylimit.Config{
			MaxBytes: server.MaxBodyBytes,
			Logger:   coreApp.Logger(),
		}),
		timeout.Middleware(timeout.Config{Timeout: server.HandlerTimeout}),
	); err != nil {
		return nil, fmt.Errorf("wire middleware: %w", err)
	}
	return coreApp, nil
}

func (a *App) Run(ctx context.Context) error {
	if a == nil {
		return fmt.Errorf("workerfleet app is required")
	}

	stopLoops, err := a.runtime.StartLoops(ctx, a.config)
	if err != nil {
		return fmt.Errorf("start runtime loops: %w", err)
	}
	stopAlertLoop, err := a.runtime.StartAlertLoop(ctx, a.config)
	if err != nil {
		stopLoops()
		return fmt.Errorf("start alert loop: %w", err)
	}
	runtimeClosed := false
	defer func() {
		stopAlertLoop()
		stopLoops()
		if !runtimeClosed {
			_ = a.runtime.Close(context.Background())
		}
	}()

	if err := a.core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	server, err := a.core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	a.core.Logger().Info("starting server", plumelog.Fields{
		"addr": a.server.Core.Addr,
	})
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	stopCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve workerfleet: %w", err)
		}
	case <-stopCtx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.server.ShutdownTimeout)
	defer cancel()
	stopAlertLoop()
	stopLoops()
	if err := a.core.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	if err := a.runtime.Close(shutdownCtx); err != nil {
		return fmt.Errorf("close runtime: %w", err)
	}
	runtimeClosed = true

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve workerfleet: %w", err)
		}
	default:
	}
	return nil
}
