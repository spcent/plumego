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
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
	"with-tenant-admin/internal/config"
	quotaadmin "with-tenant-admin/internal/quota/admin"
	tenantadmin "with-tenant-admin/internal/tenant/admin"
	"with-tenant-admin/internal/usage"
)

type Deps struct {
	Logger       plumelog.StructuredLogger
	TenantConfig *tenantcore.InMemoryConfigManager
	QuotaManager tenantcore.QuotaManager
	QuotaStore   *tenantcore.InMemoryQuotaStore
	TenantStore  *tenantadmin.InMemoryStore
	UsageStore   *usage.InMemoryUsageStore
}

type App struct {
	Core         *core.App
	Cfg          config.Config
	Logger       plumelog.StructuredLogger
	TenantConfig *tenantcore.InMemoryConfigManager
	QuotaManager tenantcore.QuotaManager
	QuotaStore   *tenantcore.InMemoryQuotaStore
	Tenants      *tenantadmin.Handler
	Quotas       *quotaadmin.Handler
	Usage        *usage.Handler
}

func New(cfg config.Config, deps Deps) (*App, error) {
	logger := deps.Logger
	if logger == nil {
		logger = plumelog.NewLogger()
	}

	coreCfg := core.DefaultConfig()
	coreCfg.Addr = cfg.Addr
	app := core.New(coreCfg, core.AppDependencies{Logger: logger})
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: logger})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	if err := app.Use(requestid.Middleware(), recoveryMw); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	tenantConfig := deps.TenantConfig
	if tenantConfig == nil {
		tenantConfig = tenantcore.NewInMemoryConfigManager()
	}
	quotaStore := deps.QuotaStore
	if quotaStore == nil {
		quotaStore = tenantcore.NewInMemoryQuotaStore()
	}
	quotaManager := deps.QuotaManager
	if quotaManager == nil {
		quotaManager = tenantcore.NewWindowQuotaManager(tenantConfig, quotaStore)
	}
	tenantStore := deps.TenantStore
	if tenantStore == nil {
		tenantStore = tenantadmin.NewInMemoryStore()
	}
	usageStore := deps.UsageStore
	if usageStore == nil {
		usageStore = usage.NewInMemoryUsageStore()
	}

	return &App{
		Core:         app,
		Cfg:          cfg,
		Logger:       logger,
		TenantConfig: tenantConfig,
		QuotaManager: quotaManager,
		QuotaStore:   quotaStore,
		Tenants:      tenantadmin.NewHandler(tenantStore),
		Quotas:       quotaadmin.NewHandler(tenantStore, tenantConfig, quotaStore),
		Usage:        usage.NewHandler(tenantStore, usageStore),
	}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
// It listens for SIGTERM and SIGINT and triggers a graceful shutdown.
func (a *App) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}
	go func() {
		<-ctx.Done()
		_ = a.Core.Shutdown(context.Background())
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server stopped: %w", err)
	}
	return nil
}
