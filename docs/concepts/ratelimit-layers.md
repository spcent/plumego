# Rate Limiting Layers

Plumego has three rate limiting layers with different scopes. They are not
redundant: each operates at a different level of granularity and integrates
into a different part of the stack.

## Decision table

| Layer | Package | Scope | When to use |
|---|---|---|---|
| **HTTP abuse guard** | `middleware/abuseguard` (wraps `security/abuse`) | Per-IP or per-key HTTP middleware | Protecting public HTTP endpoints from client abuse |
| **Token bucket primitive** | `x/resilience/ratelimit` | Per-call token bucket for extension packages | Wrapping internal provider calls, outbound integrations, messaging providers |
| **Per-tenant limiter** | `x/tenant` (via `x/tenant/core`) | Per-tenant policy-driven token bucket | Enforcing tenant quotas in multi-tenant applications |

## Layer details

### `middleware/abuseguard` + `security/abuse` — HTTP middleware

```
security/abuse.Limiter
middleware/abuseguard.New(limiter)
```

- Operates at the HTTP handler boundary
- Keyed by remote IP, header, or a caller-supplied key function
- Returns `429 Too Many Requests` via `contract.WriteError`
- **Use this** to protect public-facing routes

### `x/resilience/ratelimit` — token bucket primitive

```
x/resilience/ratelimit.New(rate, burst)         → *TokenBucket
x/resilience/ratelimit.NewKeyed(rate, burst)    → *KeyedTokenBucket
```

- Generic token bucket; no HTTP concerns
- Used by `x/messaging` as a type alias (`RateLimiter = ratelimit.TokenBucket`)
- Used by `x/ai/resilience` to wrap provider call budgets
- **Use this** anywhere inside an extension package that needs to throttle
  outbound calls, messaging providers, or internal work queues

### `x/tenant/core` — per-tenant policy limiter

```
x/tenant/core.RateLimiter
x/tenant/core.TokenBucketRateLimiter
x/tenant/core.RateLimitConfigProvider
```

- Policy-driven: per-tenant `RequestsPerSecond` and `Burst` loaded from a
  `RateLimitConfigProvider`
- Supports caller-supplied `Now` for deterministic testing and in-place
  reconfiguration without losing accumulated tokens
- Intentionally does **not** delegate to `x/resilience/ratelimit.TokenBucket`
  because it requires per-tenant state management and configurable time injection
- **Use this** for multi-tenant SaaS quota enforcement

## What does NOT belong here

- **Circuit breaking**: use `x/resilience/circuitbreaker`
- **Quota accumulation over sliding windows**: use `x/tenant/core.QuotaManager`
- **Per-operation retry budgets for AI providers**: use `x/ai/resilience`

## Common mistakes

| Mistake | Correct approach |
|---|---|
| Implementing a new token bucket inline in an extension package | Use `x/resilience/ratelimit.New` |
| Adding HTTP rate limit logic inside a domain handler | Wire `middleware/abuseguard` at the router level |
| Sharing a single global `x/resilience/ratelimit` bucket across tenants | Use `x/tenant/core.TokenBucketRateLimiter` with per-tenant state |

## Import rules summary

```
security/abuse          ← stdlib only; no x/* imports
middleware/abuseguard   ← imports security/abuse and contract
x/resilience/ratelimit  ← imports contract and middleware only
x/tenant/core           ← may import x/resilience (listed in allowed_imports)
x/ai/resilience         ← imports x/resilience; forbidden from x/tenant/**
```

See `specs/dependency-rules.yaml` for the machine-checked version.
