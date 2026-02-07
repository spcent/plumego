# Tenant Package - Quick Start Guide

> **Note**: This is for v1.1+ when tenant package becomes stable. For v1.0 see experimental usage.

## 30-Second Overview

Plumego's tenant package enables multi-tenant SaaS applications with:
- **Tenant isolation**: Automatic request scoping by tenant ID
- **Quota enforcement**: Per-tenant rate limits (requests/min, tokens/min)
- **Policy control**: Feature flags and access control per tenant
- **Database isolation**: Helpers for tenant-scoped queries

---

## Installation

```bash
go get github.com/spcent/plumego@v1.1.0
```

---

## Quick Example

### Basic Setup (In-Memory)

```go
package main

import (
    "github.com/spcent/plumego"
    "log"
)

func main() {
    // 1. Create tenant config manager
    tenantMgr := plumego.NewInMemoryTenantConfigManager()

    // 2. Add a test tenant
    tenantMgr.SetTenantConfig(plumego.TenantConfig{
        TenantID: "acme-corp",
        Quota: plumego.TenantQuotaConfig{
            RequestsPerMinute: 100,
            TokensPerMinute:   10000,
        },
        Policy: plumego.TenantPolicyConfig{
            AllowedModels: []string{"gpt-4", "gpt-3.5-turbo"},
            AllowedTools:  []string{"search", "calculator"},
        },
    })

    // 3. Create quota and policy managers
    quotaMgr := plumego.NewInMemoryQuotaManager(tenantMgr)
    policyEval := plumego.NewConfigPolicyEvaluator(tenantMgr)

    // 4. Create app with tenant support
    app := plumego.New(
        plumego.WithAddr(":8080"),
        plumego.WithTenantConfigManager(tenantMgr),
        plumego.WithTenantMiddleware(plumego.TenantMiddlewareOptions{
            HeaderName:      "X-Tenant-ID",
            AllowMissing:    false,  // Require tenant on all requests
            QuotaManager:    quotaMgr,
            PolicyEvaluator: policyEval,
        }),
    )

    // 5. Add tenant-scoped routes
    app.Get("/api/data", func(ctx *plumego.Context) {
        tenantID := plumego.TenantIDFromContext(ctx.Request.Context())
        ctx.JSON(200, map[string]string{
            "tenant": tenantID,
            "message": "This endpoint is tenant-scoped",
        })
    })

    log.Fatal(app.Boot())
}
```

### Test It

```bash
# Successful request
curl -H "X-Tenant-ID: acme-corp" http://localhost:8080/api/data

# Quota exceeded (after 100 requests in a minute)
for i in {1..101}; do curl -H "X-Tenant-ID: acme-corp" http://localhost:8080/api/data; done

# Missing tenant ID
curl http://localhost:8080/api/data
# Returns: 401 Unauthorized

# Unknown tenant
curl -H "X-Tenant-ID: unknown" http://localhost:8080/api/data
# Returns: 404 Tenant Not Found
```

---

## Database-Backed Configuration

```go
import (
    "database/sql"
    _ "github.com/lib/pq"  // or your DB driver
)

func main() {
    // 1. Connect to database
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/plumego")
    if err != nil {
        log.Fatal(err)
    }

    // 2. Create DB-backed tenant manager
    tenantMgr := plumego.NewDBTenantConfigManager(db,
        plumego.WithTenantCache(true),  // Enable LRU caching
        plumego.WithCacheSize(1000),
        plumego.WithCacheTTL(5 * time.Minute),
    )

    // 3. Rest is the same as in-memory
    quotaMgr := plumego.NewInMemoryQuotaManager(tenantMgr)
    // ...
}
```

**Database Schema**:
```sql
CREATE TABLE tenants (
    id VARCHAR(255) PRIMARY KEY,
    quota_requests_per_minute INT DEFAULT 0,
    quota_tokens_per_minute INT DEFAULT 0,
    allowed_models TEXT,  -- JSON array
    allowed_tools TEXT,   -- JSON array
    metadata TEXT,        -- JSON object
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Tenant-Scoped Database Queries

```go
import "github.com/spcent/plumego/store/db"

func listResources(ctx *plumego.Context) {
    // Automatic tenant filtering
    rows, err := db.QueryFromContext(ctx.Request.Context(),
        "SELECT id, name FROM resources WHERE status = ?",
        "active",
    )
    // Automatically becomes:
    // SELECT id, name FROM resources WHERE tenant_id = 'acme-corp' AND status = 'active'

    // ... process rows
}
```

---

## Custom Tenant Resolution

```go
// Extract tenant from JWT claims instead of header
app := plumego.New(
    plumego.WithTenantMiddleware(plumego.TenantMiddlewareOptions{
        DisablePrincipal: false,  // Allow from Principal (JWT)
        AllowMissing:     false,
    }),
)

// JWT middleware should set principal with tenant
app.Use(middleware.JWT(middleware.JWTOptions{
    Secret: "your-secret",
    OnVerified: func(w http.ResponseWriter, r *http.Request, claims jwt.Claims) {
        principal := &plumego.Principal{
            UserID:   claims.Subject,
            TenantID: claims.TenantID,  // Extract from JWT
        }
        r = plumego.RequestWithPrincipal(r, principal)
    },
}))
```

---

## Monitoring & Hooks

```go
app := plumego.New(
    plumego.WithTenantMiddleware(plumego.TenantMiddlewareOptions{
        Hooks: plumego.TenantHooks{
            // Log all tenant resolutions
            OnResolve: func(ctx context.Context, info plumego.TenantResolveInfo) {
                log.Printf("Tenant resolved: %s from %s", info.TenantID, info.Source)
            },

            // Alert on quota rejections
            OnQuota: func(ctx context.Context, info plumego.TenantQuotaDecision) {
                if !info.Allowed {
                    log.Printf("Quota exceeded for tenant %s: retry after %s",
                        info.TenantID, info.RetryAfter)
                    // Send to monitoring system
                }
            },

            // Audit policy denials
            OnPolicy: func(ctx context.Context, info plumego.TenantPolicyDecision) {
                if !info.Allowed {
                    log.Printf("Policy denied for tenant %s: %s",
                        info.TenantID, info.Reason)
                    // Log to audit trail
                }
            },
        },
    }),
)
```

---

## Admin API for Tenant Management

```go
// Admin routes (no tenant required)
admin := app.Group("/admin")
admin.Use(middleware.BasicAuth(...))  // Protect admin routes

// Create tenant
admin.POST("/tenants", func(ctx *plumego.Context) {
    var req struct {
        TenantID string `json:"tenant_id"`
        Quota    struct {
            RequestsPerMinute int `json:"requests_per_minute"`
            TokensPerMinute   int `json:"tokens_per_minute"`
        } `json:"quota"`
    }
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(400, map[string]string{"error": err.Error()})
        return
    }

    cfg := plumego.TenantConfig{
        TenantID: req.TenantID,
        Quota: plumego.TenantQuotaConfig{
            RequestsPerMinute: req.Quota.RequestsPerMinute,
            TokensPerMinute:   req.Quota.TokensPerMinute,
        },
    }

    if err := tenantMgr.SetTenantConfig(ctx.Request.Context(), cfg); err != nil {
        ctx.JSON(500, map[string]string{"error": err.Error()})
        return
    }

    ctx.JSON(201, cfg)
})

// Get tenant
admin.GET("/tenants/:id", func(ctx *plumego.Context) {
    tenantID := ctx.Param("id")
    cfg, err := tenantMgr.GetTenantConfig(ctx.Request.Context(), tenantID)
    if err != nil {
        ctx.JSON(404, map[string]string{"error": "tenant not found"})
        return
    }
    ctx.JSON(200, cfg)
})

// Update tenant quota
admin.PUT("/tenants/:id/quota", func(ctx *plumego.Context) {
    // ... implementation
})
```

---

## Advanced: Sliding Window Quota

For smoother rate limiting without burst at window boundaries:

```go
quotaMgr := plumego.NewSlidingWindowQuotaManager(tenantMgr,
    plumego.WithRedisBackend(redisClient),  // Distributed quota
    plumego.WithWindowSize(60 * time.Second),
    plumego.WithSubWindows(6),  // 10-second sub-windows
)
```

---

## Best Practices

### 1. Always Validate Tenant Ownership

```go
func getResource(ctx *plumego.Context) {
    tenantID := plumego.TenantIDFromContext(ctx.Request.Context())
    resourceID := ctx.Param("id")

    // Query with tenant filter
    var resource Resource
    err := db.QueryRowContext(ctx.Request.Context(),
        "SELECT * FROM resources WHERE id = ? AND tenant_id = ?",
        resourceID, tenantID,
    ).Scan(&resource.ID, &resource.Name)

    if err == sql.ErrNoRows {
        ctx.JSON(404, map[string]string{"error": "not found"})
        return
    }

    ctx.JSON(200, resource)
}
```

### 2. Use Transactions for Multi-Tenant Operations

```go
func createResourceWithRelations(ctx *plumego.Context) {
    tenantID := plumego.TenantIDFromContext(ctx.Request.Context())

    tx, err := db.BeginTx(ctx.Request.Context(), nil)
    if err != nil {
        ctx.JSON(500, map[string]string{"error": err.Error()})
        return
    }
    defer tx.Rollback()

    // All queries automatically scoped to tenant
    _, err = tx.ExecContext(ctx.Request.Context(),
        "INSERT INTO resources (tenant_id, name) VALUES (?, ?)",
        tenantID, "resource-1",
    )
    // ... more queries

    if err := tx.Commit(); err != nil {
        ctx.JSON(500, map[string]string{"error": err.Error()})
        return
    }

    ctx.JSON(201, map[string]string{"status": "created"})
}
```

### 3. Cache Tenant Configs

```go
// Enable caching to reduce DB load
tenantMgr := plumego.NewDBTenantConfigManager(db,
    plumego.WithTenantCache(true),
    plumego.WithCacheSize(1000),        // Max tenants in cache
    plumego.WithCacheTTL(5*time.Minute),  // Refresh every 5min
)
```

### 4. Monitor Per-Tenant Metrics

```go
import "github.com/spcent/plumego/metrics"

tenantMetrics := metrics.NewTenantMetrics()

app.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)

        tenantMetrics.RecordRequest(r.Context(), time.Since(start))
    })
})
```

---

## Common Patterns

### Pattern 1: Optional Tenant for Public + Private Routes

```go
// Public routes (no tenant)
app.GET("/public/status", func(ctx *plumego.Context) {
    ctx.JSON(200, map[string]string{"status": "ok"})
})

// Private routes (tenant required)
api := app.Group("/api")
api.Use(middleware.TenantResolver(middleware.TenantResolverOptions{
    AllowMissing: false,  // Require tenant
}))
api.GET("/data", getTenantData)
```

### Pattern 2: Admin Routes Bypass Tenant

```go
admin := app.Group("/admin")
admin.Use(middleware.BasicAuth(...))

// Admin can view all tenants' data
admin.GET("/all-resources", func(ctx *plumego.Context) {
    // Don't filter by tenant - use raw queries
    rows, err := db.Query("SELECT * FROM resources")
    // ...
})
```

### Pattern 3: Tenant-Aware Caching

```go
import "github.com/spcent/plumego/store/cache"

tenantCache := cache.NewTenantCache(baseCache)

func getCachedData(ctx *plumego.Context) {
    // Cache key automatically prefixed with tenant ID
    data, err := tenantCache.Get(ctx.Request.Context(), "user-profile")
    // Actual key: "tenant:acme-corp:user-profile"

    if err != nil {
        // Fetch from DB, cache it
        data = fetchFromDB(ctx)
        tenantCache.Set(ctx.Request.Context(), "user-profile", data, 5*time.Minute)
    }

    ctx.JSON(200, data)
}
```

---

## Troubleshooting

### Tenant ID Not Found in Context

**Problem**: `plumego.TenantIDFromContext()` returns empty string

**Solution**: Ensure `TenantResolver` middleware runs before your handler:
```go
app.Use(middleware.TenantResolver(...))  // Must be before routes
app.GET("/api/data", handler)
```

### Quota Always Allowed

**Problem**: Quota never enforces limits

**Solution**: Check quota config is non-zero:
```go
cfg := plumego.TenantConfig{
    Quota: plumego.TenantQuotaConfig{
        RequestsPerMinute: 100,  // Must be > 0
    },
}
```

### Cross-Tenant Data Leaks

**Problem**: User A from tenant X sees user B's data from tenant Y

**Solution**: Always filter queries by tenant ID:
```go
// ❌ BAD: No tenant filter
rows, _ := db.Query("SELECT * FROM users WHERE id = ?", userID)

// ✅ GOOD: Tenant-scoped
rows, _ := db.QueryFromContext(ctx, "SELECT * FROM users WHERE id = ?", userID)
```

---

## Next Steps

- Read **TENANT_PRODUCTION_PLAN.md** for architecture details
- See **examples/multi-tenant-saas/** for complete working example
- Review **MULTI_TENANT_ARCHITECTURE.md** for design patterns

---

## API Reference

### Core Types

```go
type TenantConfig struct {
    TenantID  string
    Quota     QuotaConfig
    Policy    PolicyConfig
    Metadata  map[string]string
    UpdatedAt time.Time
}

type QuotaConfig struct {
    RequestsPerMinute int  // 0 = unlimited
    TokensPerMinute   int  // 0 = unlimited
}

type PolicyConfig struct {
    AllowedModels []string  // Empty = allow all
    AllowedTools  []string  // Empty = allow all
}
```

### Context Functions

```go
// Extract tenant ID from context
func TenantIDFromContext(ctx context.Context) string

// Inject tenant ID into context
func ContextWithTenantID(ctx context.Context, tenantID string) context.Context

// Wrap HTTP request with tenant
func RequestWithTenantID(r *http.Request, tenantID string) *http.Request
```

### Config Manager Interface

```go
type ConfigManager interface {
    GetTenantConfig(ctx context.Context, tenantID string) (Config, error)
}

// Implementations:
func NewInMemoryConfigManager() *InMemoryConfigManager
func NewDBTenantConfigManager(db *sql.DB, options...) *DBTenantConfigManager
```

### Quota Manager Interface

```go
type QuotaManager interface {
    Allow(ctx context.Context, tenantID string, req QuotaRequest) (QuotaResult, error)
}

// Implementations:
func NewInMemoryQuotaManager(provider QuotaConfigProvider) *InMemoryQuotaManager
func NewSlidingWindowQuotaManager(provider, options...) *SlidingWindowQuotaManager
```

---

**Last Updated**: 2026-02-03
**Version**: v1.1 (unreleased)
**Status**: Planning / Draft
