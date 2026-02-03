package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spcent/plumego"
	"github.com/spcent/plumego/store/db"
)

func main() {
	// Initialize database
	database, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create tenant config manager with caching
	tenantMgr := db.NewDBTenantConfigManager(
		database,
		db.WithTenantCache(1000, 5*time.Minute), // Cache up to 1000 tenants for 5 minutes
	)

	// Create quota and policy managers
	quotaMgr := plumego.NewInMemoryQuotaManager(tenantMgr)
	policyEval := plumego.NewConfigPolicyEvaluator(tenantMgr)

	// Create tenant-aware database wrapper
	tenantDB := db.NewTenantDB(database)

	// Create application with tenant support
	app := plumego.New(
		plumego.WithAddr(getEnv("APP_ADDR", ":8080")),
		plumego.WithDebug(),

		// Add tenant configuration manager
		plumego.WithTenantConfigManager(tenantMgr),

		// Add tenant middleware with quota and policy enforcement
		plumego.WithTenantMiddleware(plumego.TenantMiddlewareOptions{
			HeaderName:      "X-Tenant-ID",
			AllowMissing:    false, // Require tenant ID on all requests
			QuotaManager:    quotaMgr,
			PolicyEvaluator: policyEval,
			Hooks: plumego.TenantHooks{
				OnResolve: func(ctx context.Context, info plumego.TenantResolveInfo) {
					log.Printf("[TENANT] Resolved: %s from %s", info.TenantID, info.Source)
				},
				OnQuota: func(ctx context.Context, decision plumego.TenantQuotaDecision) {
					if !decision.Allowed {
						log.Printf("[QUOTA] Denied for %s (remaining: %d requests, %d tokens, retry after %v)",
							decision.TenantID, decision.RemainingRequests, decision.RemainingTokens, decision.RetryAfter)
					}
				},
				OnPolicy: func(ctx context.Context, decision plumego.TenantPolicyDecision) {
					if !decision.Allowed {
						log.Printf("[POLICY] Denied for %s: %s (model=%s, tool=%s)",
							decision.TenantID, decision.Reason, decision.Model, decision.Tool)
					}
				},
			},
		}),
	)

	// Register routes
	registerAdminRoutes(app, tenantMgr)
	registerAPIRoutes(app, tenantDB)
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
func registerAdminRoutes(app *plumego.App, mgr *db.DBTenantConfigManager) {
	admin := &AdminHandler{manager: mgr}

	// Admin routes (no tenant middleware - uses admin auth instead)
	app.PostCtx("/admin/tenants", admin.CreateTenant)
	app.GetCtx("/admin/tenants/:id", admin.GetTenant)
	app.PutCtx("/admin/tenants/:id", admin.UpdateTenant)
	app.DeleteCtx("/admin/tenants/:id", admin.DeleteTenant)
	app.GetCtx("/admin/tenants", admin.ListTenants)
}

// registerAPIRoutes registers tenant-scoped business API routes
func registerAPIRoutes(app *plumego.App, tenantDB *db.TenantDB) {
	api := &APIHandler{db: tenantDB}

	// Business API routes (protected by tenant middleware)
	app.GetCtx("/api/users", api.ListUsers)
	app.PostCtx("/api/users", api.CreateUser)
	app.GetCtx("/api/users/:id", api.GetUser)
	app.DeleteCtx("/api/users/:id", api.DeleteUser)

	app.GetCtx("/api/analytics/requests", api.GetRequestAnalytics)
}

// registerHealthRoutes registers health check endpoints
func registerHealthRoutes(app *plumego.App) {
	app.GetCtx("/health", func(ctx *plumego.Context) {
		ctx.JSON(200, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	app.GetCtx("/", func(ctx *plumego.Context) {
		ctx.JSON(200, map[string]string{
			"service": "Multi-Tenant SaaS Example",
			"version": "1.0.0",
			"docs":    "/api/docs",
		})
	})
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
