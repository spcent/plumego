package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	workerapp "workerfleet/internal/app"
	"workerfleet/internal/handler"
)

const defaultShutdownTimeout = 10 * time.Second

type serverConfig struct {
	Core            core.AppConfig
	ShutdownTimeout time.Duration
}

func main() {
	if err := run(context.Background(), os.LookupEnv); err != nil {
		log.Fatalf("workerfleet stopped: %v", err)
	}
}

func run(ctx context.Context, lookup func(string) (string, bool)) error {
	appCfg, err := workerapp.LoadConfig(lookup)
	if err != nil {
		return fmt.Errorf("load app config: %w", err)
	}
	serverCfg, err := loadServerConfig(lookup)
	if err != nil {
		return fmt.Errorf("load server config: %w", err)
	}

	runtime, err := workerapp.Bootstrap(ctx, appCfg)
	if err != nil {
		return fmt.Errorf("bootstrap workerfleet: %w", err)
	}
	stopLoops, err := runtime.StartLoops(ctx, appCfg)
	if err != nil {
		return fmt.Errorf("start runtime loops: %w", err)
	}
	runtimeClosed := false
	defer func() {
		stopLoops()
		if !runtimeClosed {
			_ = runtime.Close(context.Background())
		}
	}()

	logger := plumelog.NewLogger()
	coreApp := core.New(serverCfg.Core, core.AppDependencies{Logger: logger})
	if err := coreApp.Use(
		requestid.Middleware(),
		recovery.Recovery(coreApp.Logger()),
		accesslog.Middleware(coreApp.Logger(), nil, nil),
	); err != nil {
		return fmt.Errorf("wire middleware: %w", err)
	}
	if err := workerapp.RegisterRoutes(coreApp, runtime.Handler, handler.NewHealthHandler(runtime.Ready), runtime.Metrics.Handler()); err != nil {
		return fmt.Errorf("register routes: %w", err)
	}
	if err := coreApp.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	server, err := coreApp.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("starting workerfleet on %s", serverCfg.Core.Addr)
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverCfg.ShutdownTimeout)
	defer cancel()
	stopLoops()
	if err := coreApp.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	if err := runtime.Close(shutdownCtx); err != nil {
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

func loadServerConfig(lookup func(string) (string, bool)) (serverConfig, error) {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	coreCfg := core.DefaultConfig()
	cfg := serverConfig{
		Core:            coreCfg,
		ShutdownTimeout: defaultShutdownTimeout,
	}
	if value, ok := lookup("WORKERFLEET_HTTP_ADDR"); ok {
		addr := strings.TrimSpace(value)
		if addr == "" {
			return serverConfig{}, fmt.Errorf("WORKERFLEET_HTTP_ADDR must not be empty")
		}
		cfg.Core.Addr = addr
	}
	if value, ok := lookup("WORKERFLEET_SHUTDOWN_TIMEOUT"); ok {
		timeout, err := time.ParseDuration(strings.TrimSpace(value))
		if err != nil {
			return serverConfig{}, fmt.Errorf("parse WORKERFLEET_SHUTDOWN_TIMEOUT: %w", err)
		}
		if timeout <= 0 {
			return serverConfig{}, fmt.Errorf("WORKERFLEET_SHUTDOWN_TIMEOUT must be greater than zero")
		}
		cfg.ShutdownTimeout = timeout
	}
	return cfg, nil
}
