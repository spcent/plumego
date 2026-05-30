package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"cloud-vault/internal/app"
	"cloud-vault/internal/config"
	"cloud-vault/internal/desktop"
	plumelog "github.com/spcent/plumego/log"
)

//go:embed all:frontend/dist
var webAssets embed.FS

var version = "0.9.0"

func main() {
	if err := run(); err != nil {
		log.Printf("desktop error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg.App.Version = version

	// Initialize desktop config
	desktopCfg := desktop.DefaultConfig()
	if cfg.Desktop.AppName != "" {
		desktopCfg.AppName = cfg.Desktop.AppName
	}
	if cfg.Desktop.DataDir != "" {
		desktopCfg.DataDir = cfg.Desktop.DataDir
	}
	if cfg.Desktop.CloseToTraySet {
		desktopCfg.CloseToTray = cfg.Desktop.CloseToTray
	}
	if cfg.Desktop.NativeNotificationsSet {
		desktopCfg.NativeNotifications = cfg.Desktop.NativeNotifications
	}

	// Ensure data directories exist
	if err := desktopCfg.EnsureDataDir(); err != nil {
		return fmt.Errorf("ensure data dir: %w", err)
	}

	// Override server config for desktop mode:
	// - Always bind to 127.0.0.1:0 (never 0.0.0.0)
	// - Use desktop data dir paths
	cfg.Core.Addr = "127.0.0.1:0"
	cfg.Core.TLS.Enabled = false // Desktop always plain HTTP
	cfg.DB.Path = filepath.Join(desktopCfg.DataDir, "app.db")
	cfg.Local.Root = filepath.Join(desktopCfg.DataDir, "objects")

	// Setup logging
	logger := plumelog.NewLogger()
	logger.Info("starting Cloud Vault Desktop", plumelog.Fields{
		"version": version,
		"dataDir": desktopCfg.DataDir,
	})

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("shutdown signal received", plumelog.Fields{})
		cancel()
	}()

	// Create application (wires handlers, middleware, routes)
	application, err := app.New(cfg)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	if err := application.RegisterRoutes(); err != nil {
		return fmt.Errorf("register routes: %w", err)
	}

	// Start local HTTP server on 127.0.0.1:0
	handler, err := application.HTTPHandler()
	if err != nil {
		return fmt.Errorf("build http handler: %w", err)
	}

	localServer, err := desktop.NewLocalServer(handler, desktopCfg.DataDir, logger)
	if err != nil {
		return fmt.Errorf("create local server: %w", err)
	}

	if err := localServer.Start(ctx); err != nil {
		return fmt.Errorf("start local server: %w", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		localServer.Stop(shutdownCtx)
	}()

	logger.Info("local server started", plumelog.Fields{
		"port":    fmt.Sprintf("%d", localServer.Port()),
		"baseURL": localServer.BaseURL(),
	})

	// Start background indexer
	application.StartBackgroundTasks(ctx)

	// Create desktop service
	desktopService := desktop.NewService(&desktopCfg, localServer, logger)

	// Create tray manager
	trayManager := desktop.NewTrayManager(&desktopCfg, localServer, desktopService, logger)
	trayManager.SetCancelFunc(cancel)

	// Create Wails bindings
	bindings := desktop.NewWailsBindings(&desktopCfg, localServer, desktopService, trayManager, logger)

	// Create Wails application options
	//
	// Note: V0.9 uses Wails AssetServer pointed at the web/dist build output.
	// The React frontend detects desktop mode via /api/v1/system/runtime and
	// uses the dynamic baseURL for API calls.
	appOptions := &options.App{
		Title:     desktopCfg.AppName,
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: webAssets,
		},
		OnStartup: func(wailsCtx context.Context) {
			bindings.SetContext(wailsCtx)
			logger.Info("wails startup complete", plumelog.Fields{})

			// Start tray in background goroutine
			go trayManager.Start()
		},
		OnShutdown: func(wailsCtx context.Context) {
			logger.Info("wails shutdown", plumelog.Fields{})
			cancel()
		},
		OnBeforeClose: func(wailsCtx context.Context) (preventClose bool) {
			if trayManager.ShouldCloseToTray() {
				logger.Info("close intercepted, hiding window", plumelog.Fields{})
				trayManager.HideWindow()
				return true
			}
			return false
		},
		Bind: []interface{}{
			bindings,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 true,
				HideToolbarSeparator:       true,
			},
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	}

	// Run Wails application (blocks until quit)
	if err := wails.Run(appOptions); err != nil {
		return fmt.Errorf("run wails: %w", err)
	}

	return nil
}
