# Tenant Module

> **Package**: `github.com/spcent/plumego/tenant`  
> **Stability**: Experimental (v1)

`tenant/*` provides multi-tenant primitives for tenant identity, quota, policy, and tenant-scoped rate limiting.

## Experimental Status

In v1, tenant APIs are available but not covered by GA compatibility guarantees. Use in production only with your own integration and isolation tests.

---

## Canonical Wiring Pattern

Use explicit router-group middleware composition.

```go
import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/spcent/plumego/core"
    tenantresolve "github.com/spcent/plumego/x/tenant/resolve"
    "github.com/spcent/plumego/tenant"
    tenantpolicy "github.com/spcent/plumego/x/tenant/policy"
    tenantquota "github.com/spcent/plumego/x/tenant/quota"
    tenantratelimit "github.com/spcent/plumego/x/tenant/ratelimit"
)

func setup() *core.App {
    tenantMgr := tenant.NewInMemoryConfigManager()
    tenantMgr.SetTenantConfig(tenant.Config{
        TenantID: "acme",
        Quota: tenant.QuotaConfig{
            Limits: []tenant.QuotaLimit{{Window: tenant.QuotaWindowMinute, Requests: 1200, Tokens: 200000}},
        },
        Policy: tenant.PolicyConfig{
            AllowedModels:  []string{"gpt-4o-mini"},
            AllowedTools:   []string{"search"},
            AllowedMethods: []string{"GET", "POST"},
            AllowedPaths:   []string{"/api/*"},
        },
        RateLimit: tenant.RateLimitConfig{RequestsPerSecond: 30, Burst: 60},
    })

    quotaMgr := tenant.NewWindowQuotaManager(tenantMgr, tenant.NewInMemoryQuotaStore())
    policyEval := tenant.NewConfigPolicyEvaluator(tenantMgr)
    rateLimiter := tenant.NewTokenBucketRateLimiter(
        &tenant.RateLimitConfigProviderFromConfig{Manager: tenantMgr},
    )

    app := core.New(core.WithAddr(":8080"))

    api := app.Router().Group("/api")
    api.Use(tenantresolve.Middleware(tenantresolve.Options{
        HeaderName:   "X-Tenant-ID",
        AllowMissing: false,
    }))
    api.Use(tenantratelimit.Middleware(tenantratelimit.Options{
        Limiter: rateLimiter,
    }))
    api.Use(tenantquota.Middleware(tenantquota.Options{
        Manager: quotaMgr,
        Hooks: tenant.Hooks{
            OnQuota: func(ctx context.Context, d tenant.QuotaDecision) {
                if !d.Allowed {
                    log.Printf("quota denied tenant=%s", d.TenantID)
                }
            },
        },
    }))
    api.Use(tenantpolicy.Middleware(tenantpolicy.Options{
        Evaluator: policyEval,
    }))

    api.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := tenant.TenantIDFromContext(r.Context())
        _, _ = w.Write([]byte("tenant=" + tenantID))
    }))

    return app
}
```

---

## Core Building Blocks

- config manager
  - `tenant.NewInMemoryConfigManager()`
  - `x/tenant/config.NewDBTenantConfigManager(...)`
- resolver middleware
  - `x/tenant/resolve.Middleware(...)`
- rate limiter middleware
  - `x/tenant/ratelimit.Middleware(...)`
- quota middleware
  - `x/tenant/quota.Middleware(...)`
- policy middleware
  - `x/tenant/policy.Middleware(...)`
- tenant-aware SQL helper
  - `x/tenant/store/db.NewTenantDB(...)`

---

## Middleware Order

Recommended order for tenant-aware APIs:

1. tenant resolve
2. tenant rate limit
3. tenant quota
4. tenant policy
5. business handler

---

## Tenant-Aware Database Access

```go
tenantDB := tenantdb.NewTenantDB(sqlDB)

api.Get("/orders", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    rows, err := tenantDB.QueryFromContext(r.Context(),
        "SELECT id, status FROM orders WHERE active = ?",
        true,
    )
    if err != nil {
        http.Error(w, "query failed", http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    w.WriteHeader(http.StatusOK)
}))
```

---

## Compatibility Notes

Use explicit middleware composition on router groups. Do not rely on removed core option-style tenant wiring.
