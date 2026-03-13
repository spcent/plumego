# Tenant Quick Start

> Location note (March 10, 2026): this file is kept under `docs/legacy/topics/tenant/` for legacy discoverability.
> Canonical v1 tenant docs live in `docs/modules/tenant/*` and `README.md` / `README_CN.md`.

Status: Experimental in v1

This guide shows the fastest path to wire tenant resolution, rate limiting, quota, and policy with current APIs.

---

## 1) Build Tenant Config Manager

```go
tenantMgr := tenant.NewInMemoryConfigManager()

tenantMgr.SetTenantConfig(tenant.Config{
    TenantID: "acme",
    Quota: tenant.QuotaConfig{
        Limits: []tenant.QuotaLimit{{Window: tenant.QuotaWindowMinute, Requests: 600, Tokens: 100000}},
    },
    Policy: tenant.PolicyConfig{
        AllowedModels:  []string{"gpt-4o-mini"},
        AllowedTools:   []string{"search"},
        AllowedMethods: []string{"GET", "POST"},
        AllowedPaths:   []string{"/api/*"},
    },
    RateLimit: tenant.RateLimitConfig{
        RequestsPerSecond: 15,
        Burst:             30,
    },
})
```

---

## 2) Create Managers

```go
quotaMgr := tenant.NewWindowQuotaManager(tenantMgr, tenant.NewInMemoryQuotaStore())
policyEval := tenant.NewConfigPolicyEvaluator(tenantMgr)
rateLimiter := tenant.NewTokenBucketRateLimiter(
    &tenant.RateLimitConfigProviderFromConfig{Manager: tenantMgr},
)
```

---

## 3) Wire Middleware Chain on API Group

```go
app := core.New(core.WithAddr(":8080"))
api := app.Router().Group("/api")

api.Use(tenanthttp.TenantResolver(tenanthttp.TenantResolverOptions{
    HeaderName:   "X-Tenant-ID",
    AllowMissing: false,
}))
api.Use(tenanthttp.TenantRateLimit(tenanthttp.TenantRateLimitOptions{
    Limiter: rateLimiter,
}))
api.Use(tenantmw.TenantQuota(tenantmw.TenantQuotaOptions{
    Manager: quotaMgr,
}))
api.Use(tenantmw.TenantPolicy(tenantmw.TenantPolicyOptions{
    Evaluator: policyEval,
}))

api.Get("/profile", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    tenantID := tenant.TenantIDFromContext(r.Context())
    _, _ = w.Write([]byte("tenant=" + tenantID))
}))
```

---

## 4) Tenant-aware Database Access (Optional)

```go
tenantDB := db.NewTenantDB(sqlDB)

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

## 5) Run and Verify

```bash
curl -H "X-Tenant-ID: acme" http://localhost:8080/api/profile
```

Expected: success response with tenant id.

Missing tenant header should be rejected.

---

## Production Notes

- Treat tenant APIs as experimental in v1.
- Add isolation and negative-path tests for your deployment.
- Keep middleware order explicit and stable.
- Avoid storing tenant identifiers in untrusted mutable fields without validation.
