package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/tenant"
	tenantconfig "github.com/spcent/plumego/x/tenant/config"
	tenantpolicy "github.com/spcent/plumego/x/tenant/policy"
	tenantquota "github.com/spcent/plumego/x/tenant/quota"
	tenantresolve "github.com/spcent/plumego/x/tenant/resolve"
	tenantdb "github.com/spcent/plumego/x/tenant/store/db"
)

func main() {
	// Initialize database
	database, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create tenant config manager with caching
	tenantMgr := tenantconfig.NewDBTenantConfigManager(
		database,
		tenantconfig.WithTenantCache(1000, 5*time.Minute), // Cache up to 1000 tenants for 5 minutes
	)

	// Create quota and policy managers
	quotaMgr := tenant.NewInMemoryQuotaManager(tenantMgr)
	policyEval := tenant.NewConfigPolicyEvaluator(tenantMgr)

	// Create tenant-aware database wrapper
	tenantDB := tenantdb.NewTenantDB(database)

	// Create application with tenant support
	app := core.New(
		core.WithAddr(getEnv("APP_ADDR", ":8080")),
		core.WithDebug(),
	)

	api := app.Router().Group("/api")
	api.Use(tenantresolve.Middleware(tenantresolve.Options{
		HeaderName:   "X-Tenant-ID",
		AllowMissing: false,
		Hooks: tenant.Hooks{
			OnResolve: func(ctx context.Context, info tenant.ResolveInfo) {
				log.Printf("[TENANT] Resolved: %s from %s", info.TenantID, info.Source)
			},
		},
	}))
	api.Use(tenantquota.Middleware(tenantquota.Options{
		Manager: quotaMgr,
		Hooks: tenant.Hooks{
			OnQuota: func(ctx context.Context, decision tenant.QuotaDecision) {
				if !decision.Allowed {
					log.Printf("[QUOTA] Denied for %s (remaining: %d requests, %d tokens, retry after %v)",
						decision.TenantID, decision.RemainingRequests, decision.RemainingTokens, decision.RetryAfter)
				}
			},
		},
	}))
	api.Use(tenantpolicy.Middleware(tenantpolicy.Options{
		Evaluator: policyEval,
		Hooks: tenant.Hooks{
			OnPolicy: func(ctx context.Context, decision tenant.PolicyDecision) {
				if !decision.Allowed {
					log.Printf("[POLICY] Denied for %s: %s (model=%s, tool=%s)",
						decision.TenantID, decision.Reason, decision.Model, decision.Tool)
				}
			},
		},
	}))

	// Register routes
	registerAdminRoutes(app, tenantMgr)
	registerAPIRoutes(api, app.Logger(), tenantDB)
	registerHealthRoutes(app)

	// Start application
	log.Printf("Starting multi-tenant SaaS application on %s", getEnv("APP_ADDR", ":8080"))
	if err := app.Boot(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}

// initDB initializes the SQLite database and runs migrations
func initDB() (*sql.DB, error) {
	dbPath := getEnv("DB_PATH", "./tenants.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations creates the necessary database tables
func runMigrations(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS tenants (
			id VARCHAR(255) PRIMARY KEY,
			quota_requests_per_minute INT NOT NULL DEFAULT 0,
			quota_tokens_per_minute INT NOT NULL DEFAULT 0,
			allowed_models TEXT,
			allowed_tools TEXT,
			metadata TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_tenants_updated_at ON tenants(updated_at);

		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			name VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

		CREATE TABLE IF NOT EXISTS api_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id VARCHAR(255) NOT NULL,
			user_id INTEGER,
			method VARCHAR(10) NOT NULL,
			path VARCHAR(500) NOT NULL,
			status_code INT NOT NULL,
			duration_ms INT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		);

		CREATE INDEX IF NOT EXISTS idx_api_requests_tenant_id ON api_requests(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_api_requests_created_at ON api_requests(created_at);
	`

	_, err := db.Exec(schema)
	return err
}

// registerAdminRoutes registers admin endpoints for tenant management
func registerAdminRoutes(app *core.App, mgr *tenantconfig.DBTenantConfigManager) {
	admin := &AdminHandler{manager: mgr}
	adaptCtx := func(handler contract.CtxHandlerFunc) http.HandlerFunc {
		return contract.AdaptCtxHandler(handler, app.Logger()).ServeHTTP
	}

	// Admin routes (no tenant middleware - uses admin auth instead)
	app.Post("/admin/tenants", adaptCtx(admin.CreateTenant))
	app.Get("/admin/tenants/:id", adaptCtx(admin.GetTenant))
	app.Put("/admin/tenants/:id", adaptCtx(admin.UpdateTenant))
	app.Delete("/admin/tenants/:id", adaptCtx(admin.DeleteTenant))
	app.Get("/admin/tenants", adaptCtx(admin.ListTenants))
}

// registerAPIRoutes registers tenant-scoped business API routes
func registerAPIRoutes(apiRoutes *router.Router, logger plumelog.StructuredLogger, tenantDB *tenantdb.TenantDB) {
	handler := &APIHandler{db: tenantDB}
	adaptCtx := func(handler contract.CtxHandlerFunc) http.HandlerFunc {
		return contract.AdaptCtxHandler(handler, logger).ServeHTTP
	}

	// Business API routes (protected by tenant middleware)
	apiRoutes.Get("/users", adaptCtx(handler.ListUsers))
	apiRoutes.Post("/users", adaptCtx(handler.CreateUser))
	apiRoutes.Get("/users/:id", adaptCtx(handler.GetUser))
	apiRoutes.Delete("/users/:id", adaptCtx(handler.DeleteUser))

	apiRoutes.Get("/analytics/requests", adaptCtx(handler.GetRequestAnalytics))
}

// registerHealthRoutes registers health check endpoints
func registerHealthRoutes(app *core.App) {
	app.Get("/health", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
		ctx.JSON(200, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	}, app.Logger()).ServeHTTP)

	app.Get("/", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
		ctx.JSON(200, map[string]string{
			"service": "Multi-Tenant SaaS Example",
			"version": "1.0.0",
			"docs":    "/api/docs",
		})
	}, app.Logger()).ServeHTTP)
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
