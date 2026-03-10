# Tenant Rate Limit

> **Packages**: `github.com/spcent/plumego/tenant`, `github.com/spcent/plumego/middleware/tenant`  
> **Stability**: Experimental (v1)

Tenant rate limiting uses a per-tenant token bucket.

---

## Rate Limit Model

```go
type RateLimitConfig struct {
    RequestsPerSecond int64
    Burst             int64
}
```

`Burst` defaults to `RequestsPerSecond` when unset.

---

## Limiter

```go
limiter := tenant.NewTokenBucketRateLimiter(
    &tenant.RateLimitConfigProviderFromConfig{Manager: configMgr},
)
```

This reads per-tenant `RateLimitConfig` from tenant config manager.

---

## Middleware Integration

```go
api := app.Router().Group("/api")

api.Use(tenanthttp.TenantRateLimit(tenanthttp.TenantRateLimitOptions{
    Limiter: limiter,
    Hooks: tenant.Hooks{
        OnRateLimit: func(ctx context.Context, d tenant.RateLimitDecision) {
            if !d.Allowed {
                log.Printf("ratelimit denied tenant=%s retry_after=%s", d.TenantID, d.RetryAfter)
            }
        },
    },
}))
```

When denied, response is `429` and includes retry/limit headers.

---

## Example Tenant Config

```go
configMgr.SetTenantConfig(tenant.Config{
    TenantID: "acme",
    RateLimit: tenant.RateLimitConfig{
        RequestsPerSecond: 20,
        Burst:             40,
    },
})
```

---

## Testing Focus

- steady-state rate enforcement
- burst allowance
- retry-after behavior
- per-tenant isolation
- concurrent request behavior (`go test -race`)
