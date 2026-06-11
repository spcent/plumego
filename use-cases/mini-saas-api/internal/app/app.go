// Package app wires application dependencies and manages the server lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/abuseguard"
	"github.com/spcent/plumego/middleware/accesslog"
	"github.com/spcent/plumego/middleware/bodylimit"
	"github.com/spcent/plumego/middleware/cors"
	"github.com/spcent/plumego/middleware/httpmetrics"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/middleware/securityheaders"
	"github.com/spcent/plumego/middleware/timeout"
	"github.com/spcent/plumego/security/jwt"
	kvstore "github.com/spcent/plumego/store/kv"
	"mini-saas-api/internal/config"
	"mini-saas-api/internal/domain/audit"
	"mini-saas-api/internal/domain/project"
	"mini-saas-api/internal/domain/session"
	"mini-saas-api/internal/domain/tenantspace"
	"mini-saas-api/internal/domain/user"
	"mini-saas-api/internal/platform/idemstore"
)

// App holds application-wide dependencies, constructed once in New and wired
// into routes in RegisterRoutes. All construction is explicit — no globals.
type App struct {
	Core *core.App
	Cfg  config.Config

	Users    *user.Service
	Spaces   *tenantspace.Service
	Projects *project.Service
	Audit    *audit.Recorder
	Idem     *idemstore.Memory

	JWT    *jwt.JWTManager
	Issuer *session.Issuer

	kv        *kvstore.KVStore
	authGuard *abuseguard.AbuseGuardMiddleware
}

// New constructs the App with explicit stable-root wiring only.
// Extension wiring (x/tenant, x/observability) is added per-route or per-card;
// all global middleware here is stdlib-only stable-root.
func New(cfg config.Config) (*App, error) {
	app := core.New(cfg.Core, core.AppDependencies{Logger: plumelog.NewLogger()})

	securityMw, err := securityheaders.Middleware(securityheaders.Config{})
	if err != nil {
		return nil, fmt.Errorf("configure security headers middleware: %w", err)
	}
	recoveryMw, err := recovery.Middleware(recovery.Config{Logger: app.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure recovery middleware: %w", err)
	}
	accesslogMw, err := accesslog.Middleware(accesslog.Config{Logger: app.Logger()})
	if err != nil {
		return nil, fmt.Errorf("configure access log middleware: %w", err)
	}
	timeoutMw := timeout.Middleware(timeout.Config{Timeout: 30 * time.Second})

	var corsOpts cors.CORSOptions
	if len(cfg.App.CORSAllowedOrigins) > 0 {
		strictOpts, err := cors.StrictDefaultOptions(cfg.App.CORSAllowedOrigins...)
		if err != nil {
			return nil, fmt.Errorf("configure CORS middleware: %w", err)
		}
		corsOpts = strictOpts
	}

	// Middleware order — outermost to innermost:
	//   requestid  → stamps correlation ID before any logging or error handling
	//   security   → security headers on all responses
	//   cors       → CORS preflight; set APP_CORS_ALLOWED_ORIGINS in production
	//   recovery   → converts panics to 500; inside cors/security so headers apply
	//   accesslog  → logs every request; after recovery so panics appear as 500
	//   bodylimit  → rejects oversized bodies; after accesslog so the 413 is logged
	//   httpmetrics→ measures handler latency and status; swap NewNoopCollector for
	//               observability.NewPrometheusCollector (card 1532, x/observability)
	//   timeout    → per-request wall-clock limit; innermost so only handler time counts
	if err := app.Use(
		requestid.Middleware(),
		securityMw,
		cors.Middleware(corsOpts),
		recoveryMw,
		accesslogMw,
		bodylimit.Middleware(bodylimit.Config{
			MaxBytes: cfg.App.MaxBodyBytes,
			Logger:   app.Logger(),
		}),
		httpmetrics.Middleware(metrics.NewNoopCollector()),
		timeoutMw,
	); err != nil {
		return nil, fmt.Errorf("register middleware: %w", err)
	}

	if cfg.App.JWTSecret == config.DevJWTSecret {
		app.Logger().Warn("using dev JWT secret — set APP_JWT_SECRET in production", plumelog.Fields{})
	}
	if cfg.App.MetricsToken == "" {
		app.Logger().Warn("metrics endpoint is unprotected — set APP_METRICS_TOKEN in production", plumelog.Fields{})
	}
	if len(cfg.App.CORSAllowedOrigins) == 0 {
		app.Logger().Warn("CORS allows all origins (*); set APP_CORS_ALLOWED_ORIGINS in production", plumelog.Fields{})
	}

	// Embedded KV store holds JWT signing keys and refresh-token records.
	kv, err := kvstore.NewKVStore(kvstore.DefaultConfig(cfg.App.DataDir))
	if err != nil {
		return nil, fmt.Errorf("open kv store: %w", err)
	}

	// JWT manager over the KV store with a deterministic HS256 key derived
	// from APP_JWT_SECRET. RotationInterval 0 keeps the seeded key active.
	if err := seedJWTKey(context.Background(), kv, cfg.App.JWTSecret); err != nil {
		_ = kv.Close()
		return nil, err
	}
	jwtCfg := jwt.DefaultJWTConfig()
	jwtCfg.Issuer = cfg.App.ServiceName
	jwtCfg.Audience = cfg.App.ServiceName
	jwtCfg.AccessExpiration = cfg.App.JWTAccessTTL
	jwtCfg.RefreshExpiration = cfg.App.JWTRefreshTTL
	jwtCfg.RotationInterval = 0
	manager, err := jwt.NewJWTManager(kv, jwtCfg)
	if err != nil {
		_ = kv.Close()
		return nil, fmt.Errorf("create jwt manager: %w", err)
	}

	// Domain services over in-memory repositories. Replace the memory stores
	// with database-backed repositories for durable deployments.
	users := user.NewService(user.NewMemoryStore())
	spaces := tenantspace.NewService(tenantspace.NewMemoryStore())
	planLookup := func(ctx context.Context, tenantID string) (string, error) {
		t, err := spaces.Get(ctx, tenantID)
		if err != nil {
			return "", err
		}
		return t.Plan, nil
	}
	projects := project.NewService(project.NewMemoryStore(), project.DefaultLimits(), planLookup)
	auditLog := audit.NewRecorder(1000)
	idem := idemstore.NewMemory()

	issuer := &session.Issuer{
		JWT:     manager,
		Refresh: session.NewStore(kvAdapter{kv: kv}, cfg.App.JWTRefreshTTL),
	}

	// Token-bucket guard against credential stuffing on /api/v1/auth/*.
	guardCfg := abuseguard.DefaultAbuseGuardConfig()
	guardCfg.Logger = app.Logger()
	authGuard := abuseguard.NewAbuseGuard(guardCfg)

	return &App{
		Core:      app,
		Cfg:       cfg,
		Users:     users,
		Spaces:    spaces,
		Projects:  projects,
		Audit:     auditLog,
		Idem:      idem,
		JWT:       manager,
		Issuer:    issuer,
		kv:        kv,
		authGuard: authGuard,
	}, nil
}

// Start prepares the runtime and blocks while the HTTP server runs.
// When ctx is canceled it triggers graceful shutdown.
func (a *App) Start(ctx context.Context) error {
	if err := a.Core.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := a.Core.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}

	a.Core.Logger().Info("starting server", plumelog.Fields{
		"addr": a.Cfg.Core.Addr,
		"tls":  a.Cfg.Core.TLS.Enabled,
	})

	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		shutdownErr <- a.Core.Shutdown(shutdownCtx)
	}()

	var serveErr error
	if a.Cfg.Core.TLS.Enabled {
		serveErr = srv.ListenAndServeTLS("", "")
	} else {
		serveErr = srv.ListenAndServe()
	}
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		a.closeResources()
		return fmt.Errorf("server stopped: %w", serveErr)
	}
	err = <-shutdownErr
	a.closeResources()
	if err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}

// closeResources releases app-owned resources after the HTTP server stops.
func (a *App) closeResources() {
	if a.authGuard != nil {
		a.authGuard.Stop()
	}
	if a.kv != nil {
		if err := a.kv.Close(); err != nil {
			a.Core.Logger().Warn("close kv store", plumelog.Fields{"error": err.Error()})
		}
	}
}
