# Rate Limiting Layers

Plumego has four rate limiting layers with different scopes. They are not
redundant: each operates at a different level of granularity and integrates
into a different part of the stack. All four are token-bucket based; the first
three share the `x/resilience/ratelimit` primitive, and `x/tenant/core`
intentionally keeps its own implementation (see below).

## Decision table

| Layer | Package | Scope | When to use |
|---|---|---|---|
| **HTTP abuse guard** | `middleware/abuseguard` (wraps `security/abuse`) | Per-IP or per-key HTTP middleware | Protecting public HTTP endpoints from client abuse |
| **Token bucket primitive** | `x/resilience/ratelimit` | Global or per-key token bucket for extension packages | Wrapping internal provider calls, outbound integrations, messaging providers |
| **Pub/sub dispatch limiter** | `x/messaging/pubsub` (via `NewRateLimited`) | Global + per-topic + per-subscriber, optionally adaptive | Throttling message dispatch in an in-process pub/sub broker |
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
x/resilience/ratelimit.New(rate, burst)         → *TokenBucket   (one global bucket)
x/resilience/ratelimit.NewKeyed(rate, burst)    → *KeyedBuckets  (one bucket per key)
```

- Generic token bucket; no HTTP concerns
- `TokenBucket` exposes `Allow()`, `AllowN(n)`, `Wait(ctx)`, `WaitN(ctx, n)`,
  `Remaining()`, and `UpdateRate(rate, burst)`
- `KeyedBuckets` lazily creates a bucket per key: `Allow(key)`, `AllowN(key, n)`,
  `Delete(key)`
- Used by `x/messaging` as a type alias (`RateLimiter = ratelimit.TokenBucket`)
- Used by `x/ai/resilience` to wrap provider call budgets
- **Use this** anywhere inside an extension package that needs to throttle
  outbound calls, messaging providers, or internal work queues

### `x/messaging/pubsub` — pub/sub dispatch limiter

```
x/messaging/pubsub.DefaultRateLimitConfig()       → RateLimitConfig
x/messaging/pubsub.NewRateLimited(config, opts…)  → *RateLimitedPubSub
```

- Wraps a pub/sub broker with three simultaneous token-bucket scopes:
  `GlobalQPS/Burst`, `PerTopicQPS/Burst`, and `PerSubscriberQPS/Burst`
- `Publish` enforces the global and per-topic limits; `Subscribe` wraps each
  subscription with its own per-subscriber limiter
- Set `config.Adaptive = true` to let it adjust rates under sustained load
- `Publish` returns `ErrRateLimitExceeded` when a limit is hit;
  `RateLimitStats()` exposes counters
- Built on `x/resilience/ratelimit.TokenBucket` internally
- **Use this** when the throttling unit is a *message topic or subscriber*, not
  an HTTP client or a tenant

```go
config := pubsub.DefaultRateLimitConfig()
config.GlobalQPS = 10
config.PerTopicQPS = 50
config.PerSubscriberQPS = 20
config.Adaptive = true

rlps, err := pubsub.NewRateLimited(config)
// ...
defer rlps.Close()
if err := rlps.Publish("orders.created", msg); errors.Is(err, pubsub.ErrRateLimitExceeded) {
    // dispatch was throttled
}
```

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
| Hand-rolling per-topic throttling around a broker | Use `x/messaging/pubsub.NewRateLimited` |

## Import rules summary

```
security/abuse          ← stdlib only; no x/* imports
middleware/abuseguard   ← imports security/abuse and contract
x/resilience/ratelimit  ← imports contract and middleware only
x/messaging/pubsub      ← wraps x/resilience/ratelimit.TokenBucket
x/tenant/core           ← may import x/resilience (listed in allowed_imports)
x/ai/resilience         ← imports x/resilience; forbidden from x/tenant/**
```

See `specs/dependency-rules.yaml` for the machine-checked version.
