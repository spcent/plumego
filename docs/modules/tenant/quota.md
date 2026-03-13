# Tenant Quota

> **Packages**: `github.com/spcent/plumego/x/tenant/core`, `github.com/spcent/plumego/x/tenant/quota`  
> **Stability**: Experimental (v1)

Quota management limits per-tenant request/token usage.

---

## Config Model

```go
type QuotaConfig struct {
    RequestsPerMinute int
    TokensPerMinute   int
    Limits            []QuotaLimit
}

type QuotaLimit struct {
    Window   QuotaWindow // minute/hour/day/month
    Requests int
    Tokens   int
}
```

When `Limits` is non-empty, it takes precedence.

---

## Managers

### Single-window in-memory

```go
quotaMgr := tenant.NewInMemoryQuotaManager(configMgr)
```

### Multi-window enforcement

```go
quotaMgr := tenant.NewWindowQuotaManager(configMgr, tenant.NewInMemoryQuotaStore())
```

Use `WindowQuotaManager` when you need simultaneous window checks (for example day + month).

---

## HTTP Middleware Integration

```go
api := app.Router().Group("/api")

api.Use(tenantquota.Middleware(tenantquota.Options{
    Manager: quotaMgr,
    Hooks: tenant.Hooks{
        OnQuota: func(ctx context.Context, d tenant.QuotaDecision) {
            if !d.Allowed {
                log.Printf("quota denied tenant=%s retry_after=%s", d.TenantID, d.RetryAfter)
            }
        },
    },
}))
```

Headers are added for remaining quota and retry timing.

---

## Example Tenant Config

```go
configMgr.SetTenantConfig(tenant.Config{
    TenantID: "acme",
    Quota: tenant.QuotaConfig{
        Limits: []tenant.QuotaLimit{
            {Window: tenant.QuotaWindowMinute, Requests: 1000, Tokens: 200000},
            {Window: tenant.QuotaWindowDay, Requests: 50000, Tokens: 5000000},
        },
    },
})
```

---

## Testing Focus

For production usage, test:

- allowed vs denied path
- retry-after correctness
- multi-window behavior
- concurrent usage updates
- hook callbacks on denial
