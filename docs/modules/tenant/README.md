# Tenant Module

> **Package Path**: `github.com/spcent/plumego/tenant` | **Stability**: High | **Priority**: P2

## Overview

The `tenant/` package provides multi-tenancy primitives for building SaaS applications. It enables a single application instance to serve multiple isolated tenants with their own configuration, quotas, policies, and data.

**Key Features**:
- **Tenant Identification**: Extract tenant ID from headers, subdomains, or JWT claims
- **Configuration Management**: Per-tenant settings with in-memory or database backends
- **Quota Enforcement**: Request and resource limits per tenant
- **Policy Evaluation**: Business rule enforcement per tenant
- **Rate Limiting**: Per-tenant rate limits
- **Database Isolation**: Automatic query scoping per tenant
- **Context Propagation**: Tenant ID flows through the request lifecycle

## Quick Start

### Basic Multi-Tenancy

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/tenant"
)

// Create in-memory config manager
configMgr := tenant.NewInMemoryConfigManager()

// Register tenants
configMgr.SetTenantConfig("tenant-a", tenant.Config{
    TenantID: "tenant-a",
    Quota: tenant.QuotaConfig{
        RequestsPerMinute: 1000,
        MaxUsers:          100,
        StorageLimitGB:    10,
    },
    Metadata: map[string]string{
        "plan": "pro",
        "name": "Acme Corp",
    },
})

// Create app with tenant middleware
app := core.New(
    core.WithAddr(":8080"),
    core.WithTenantConfigManager(configMgr),
    core.WithTenantMiddleware(core.TenantMiddlewareOptions{
        HeaderName:   "X-Tenant-ID",
        AllowMissing: false,
    }),
)

// All handlers now have tenant context
app.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())
    // Use tenantID to scope queries...
})

app.Boot()
```

### Production Setup with Database

```go
// Database-backed config manager with LRU cache
configMgr := db.NewDBTenantConfigManager(
    database,
    db.WithTenantCache(1000, 5*time.Minute),
)

// Quota manager
quotaMgr := tenant.NewInMemoryQuotaManager(configMgr)

// Policy evaluator
policyEval := tenant.NewConfigPolicyEvaluator(configMgr)

app := core.New(
    core.WithTenantConfigManager(configMgr),
    core.WithTenantMiddleware(core.TenantMiddlewareOptions{
        HeaderName:      "X-Tenant-ID",
        AllowMissing:    false,
        QuotaManager:    quotaMgr,
        PolicyEvaluator: policyEval,
        Hooks: tenant.Hooks{
            OnQuota: func(ctx context.Context, d tenant.QuotaDecision) {
                log.Warn("Quota exceeded", "tenant", d.TenantID, "resource", d.Resource)
            },
            OnPolicy: func(ctx context.Context, d tenant.PolicyDecision) {
                log.Warn("Policy denied", "tenant", d.TenantID, "policy", d.PolicyName)
            },
        },
    }),
)
```

## Core Types

### Config

Per-tenant configuration:

```go
type Config struct {
    TenantID  string
    Quota     QuotaConfig
    Policy    PolicyConfig
    Metadata  map[string]string
    UpdatedAt time.Time
}

type QuotaConfig struct {
    RequestsPerMinute int
    RequestsPerDay    int
    MaxUsers          int
    StorageLimitGB    float64
    MaxConnections    int
}

type PolicyConfig struct {
    AllowedFeatures  []string
    BlockedIPs       []string
    AllowedCountries []string
    CustomRules      map[string]interface{}
}
```

### Key Interfaces

```go
// ConfigManager retrieves tenant configuration
type ConfigManager interface {
    GetTenantConfig(ctx context.Context, tenantID string) (Config, error)
}

// QuotaManager enforces resource quotas
type QuotaManager interface {
    CheckQuota(ctx context.Context, tenantID, resource string, amount int) (QuotaDecision, error)
    RecordUsage(ctx context.Context, tenantID, resource string, amount int) error
}

// PolicyEvaluator evaluates tenant policies
type PolicyEvaluator interface {
    EvaluatePolicy(ctx context.Context, tenantID, policy string, input map[string]interface{}) (PolicyDecision, error)
}

// RateLimiter provides per-tenant rate limiting
type RateLimiter interface {
    Allow(ctx context.Context, tenantID string) bool
}
```

## Tenant Identification

### From HTTP Header

```go
core.WithTenantMiddleware(core.TenantMiddlewareOptions{
    HeaderName:   "X-Tenant-ID",
    AllowMissing: false,
})

// Client sends: X-Tenant-ID: tenant-123
```

### From Subdomain

```go
core.WithTenantMiddleware(core.TenantMiddlewareOptions{
    ExtractFrom: tenant.ExtractFromSubdomain,
    // tenant-123.example.com → tenant ID: "tenant-123"
})
```

### From JWT Claims

```go
core.WithTenantMiddleware(core.TenantMiddlewareOptions{
    ExtractFrom: tenant.ExtractFromJWT("tenant_id"),
    // JWT payload: {"tenant_id": "tenant-123", ...}
})
```

### Custom Extraction

```go
core.WithTenantMiddleware(core.TenantMiddlewareOptions{
    ExtractFn: func(r *http.Request) (string, error) {
        // Custom logic: extract from path, query, or session
        return r.URL.Query().Get("tenant"), nil
    },
})
```

## Context Helpers

```go
// Extract tenant ID from context
tenantID := tenant.TenantIDFromContext(ctx)

// Add tenant ID to context
ctx = tenant.ContextWithTenantID(ctx, "tenant-123")

// Extract full tenant config from context
config, err := tenant.ConfigFromContext(ctx)
```

## Architecture

```
Request
  │
  ▼
Tenant Middleware
  │
  ├─ Extract Tenant ID (header/subdomain/JWT)
  ├─ Load Tenant Config (memory/database)
  ├─ Check Quota (optional)
  ├─ Evaluate Policy (optional)
  └─ Add to Context
        │
        ▼
Handler
  │
  ├─ tenant.TenantIDFromContext(ctx)
  ├─ TenantDB.QueryFromContext(ctx, ...)
  └─ Business logic with tenant isolation
```

## Plan Tiers

### Implementing Multiple Plans

```go
const (
    PlanFree       = "free"
    PlanStarter    = "starter"
    PlanPro        = "pro"
    PlanEnterprise = "enterprise"
)

var planQuotas = map[string]tenant.QuotaConfig{
    PlanFree:       {RequestsPerMinute: 60, MaxUsers: 3, StorageLimitGB: 1},
    PlanStarter:    {RequestsPerMinute: 300, MaxUsers: 10, StorageLimitGB: 5},
    PlanPro:        {RequestsPerMinute: 1000, MaxUsers: 100, StorageLimitGB: 50},
    PlanEnterprise: {RequestsPerMinute: 10000, MaxUsers: 10000, StorageLimitGB: 1000},
}

func getTenantConfig(tenantID string) (tenant.Config, error) {
    t, err := db.GetTenant(tenantID)
    if err != nil {
        return tenant.Config{}, err
    }

    quota := planQuotas[t.Plan]
    return tenant.Config{
        TenantID: tenantID,
        Quota:    quota,
        Metadata: map[string]string{
            "plan": t.Plan,
            "name": t.Name,
        },
    }, nil
}
```

## Admin API

### Tenant CRUD Operations

```go
// Create tenant
func handleCreateTenant(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
        Plan  string `json:"plan" validate:"required,oneof=free starter pro enterprise"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    tenantID := uuid.New().String()
    config := tenant.Config{
        TenantID: tenantID,
        Quota:    planQuotas[req.Plan],
        Metadata: map[string]string{
            "name":  req.Name,
            "email": req.Email,
            "plan":  req.Plan,
        },
    }

    if err := configMgr.SetTenantConfig(tenantID, config); err != nil {
        http.Error(w, "Failed to create tenant", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{
        "tenant_id": tenantID,
        "api_key":   generateAPIKey(tenantID),
    })
}

// Update tenant plan
func handleUpdatePlan(w http.ResponseWriter, r *http.Request) {
    tenantID := plumego.Param(r, "tenant_id")

    var req struct {
        Plan string `json:"plan" validate:"required,oneof=free starter pro enterprise"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    config, _ := configMgr.GetTenantConfig(r.Context(), tenantID)
    config.Quota = planQuotas[req.Plan]
    config.Metadata["plan"] = req.Plan

    configMgr.SetTenantConfig(tenantID, config)

    json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}
```

## Module Documentation

- **[Configuration](configuration.md)** — Tenant config management (in-memory, database)
- **[Quota](quota.md)** — Quota enforcement and usage tracking
- **[Policy](policy.md)** — Policy evaluation and feature flags
- **[Rate Limiting](ratelimit.md)** — Per-tenant rate limiting
- **[Database Isolation](database-isolation.md)** — Automatic tenant query scoping
- **[Examples](examples.md)** — Complete multi-tenant SaaS example

## Related Documentation

- [Core: Configuration Options](../core/configuration-options.md) — Tenant middleware options
- [Middleware: Tenant](../middleware/tenant.md) — Tenant routing middleware
- [Store: DB](../store/db/) — Tenant-isolated database

## Reference Implementation

- `examples/multi-tenant-saas/` — Complete working multi-tenant application
