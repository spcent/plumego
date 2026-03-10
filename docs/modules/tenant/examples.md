# Tenant Examples

> **Stability**: Experimental (v1)

This page provides compile-oriented example patterns for tenant wiring.

---

## 1) Minimal Header-Based Tenant API

```go
app := core.New(core.WithAddr(":8080"))

api := app.Router().Group("/api")
api.Use(tenanthttp.TenantResolver(tenanthttp.TenantResolverOptions{
    HeaderName:   "X-Tenant-ID",
    AllowMissing: false,
}))

api.Get("/me", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())
    _, _ = w.Write([]byte("tenant=" + tenantID))
}))
```

---

## 2) Full Chain (Resolve + Rate + Quota + Policy)

```go
tenantMgr := tenant.NewInMemoryConfigManager()
quotaMgr := tenant.NewWindowQuotaManager(tenantMgr, tenant.NewInMemoryQuotaStore())
policyEval := tenant.NewConfigPolicyEvaluator(tenantMgr)
rateLimiter := tenant.NewTokenBucketRateLimiter(&tenant.RateLimitConfigProviderFromConfig{Manager: tenantMgr})

api := app.Router().Group("/api")
api.Use(tenanthttp.TenantResolver(tenanthttp.TenantResolverOptions{HeaderName: "X-Tenant-ID"}))
api.Use(tenanthttp.TenantRateLimit(tenanthttp.TenantRateLimitOptions{Limiter: rateLimiter}))
api.Use(tenantmw.TenantQuota(tenantmw.TenantQuotaOptions{Manager: quotaMgr}))
api.Use(tenantmw.TenantPolicy(tenantmw.TenantPolicyOptions{Evaluator: policyEval}))
```

---

## 3) Tenant-Aware SQL Query

```go
tenantDB := db.NewTenantDB(sqlDB)

api.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    rows, err := tenantDB.QueryFromContext(r.Context(),
        "SELECT id, email FROM users WHERE active = ?",
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

## 4) DB-backed Tenant Config Manager

```go
tenantMgr := db.NewDBTenantConfigManager(
    sqlDB,
    db.WithTenantCache(1000, 5*time.Minute),
)

rateLimiter := tenant.NewTokenBucketRateLimiter(
    &tenant.RateLimitConfigProviderFromConfig{Manager: tenantMgr},
)
_ = rateLimiter
```

---

## 5) Hook-based Auditing

```go
hooks := tenant.Hooks{
    OnResolve: func(ctx context.Context, d tenant.ResolveInfo) {
        log.Printf("tenant resolved: %s", d.TenantID)
    },
    OnQuota: func(ctx context.Context, d tenant.QuotaDecision) {
        if !d.Allowed {
            log.Printf("quota denied: %s", d.TenantID)
        }
    },
    OnPolicy: func(ctx context.Context, d tenant.PolicyDecision) {
        if !d.Allowed {
            log.Printf("policy denied: %s", d.TenantID)
        }
    },
    OnRateLimit: func(ctx context.Context, d tenant.RateLimitDecision) {
        if !d.Allowed {
            log.Printf("ratelimit denied: %s", d.TenantID)
        }
    },
}
_ = hooks
```

---

For a larger runnable sample, see `examples/multi-tenant-saas/` and align APIs with current canonical middleware wiring when adapting.
