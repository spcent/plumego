# Card 0914

Priority: P2
State: active
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/ratelimit.go`
Depends On: —

Goal:
- Eliminate the silent-allow inconsistency between `InMemoryRateLimitProvider` and the
  other in-memory providers so callers get the same "not found" signal regardless of
  which provider they use.

---

## Problem

**File:** `x/tenant/core/ratelimit.go` lines 76–84

```go
func (p *InMemoryRateLimitProvider) RateLimitConfig(ctx context.Context, tenantID string) (RateLimitConfig, error) {
    ...
    // Unknown tenant → zero config (unlimited), no error.
    return p.configs[tenantID], nil
}
```

This returns a zero-value `RateLimitConfig` (unlimited, no error) when the tenant has
no explicit configuration. Every other in-memory provider in the package returns
`ErrTenantNotFound` for unknown tenants:

| Provider | Unknown tenant behavior |
|----------|------------------------|
| `InMemoryConfigManager.GetTenantConfig` | returns `ErrTenantNotFound` |
| `InMemoryRoutePolicyStore.RoutePolicy` | returns `ErrRoutePolicyNotFound` |
| `InMemoryRateLimitProvider.RateLimitConfig` | **returns zero config, nil** ← inconsistent |

The inconsistency means:
- A caller who switches from `InMemoryConfigManager` (via `RateLimitConfigProviderFromConfig`)
  to `InMemoryRateLimitProvider` gets silently different behavior for unconfigured tenants.
- There is no way for `TokenBucketRateLimiter` to distinguish "this tenant has no config"
  from "this tenant explicitly has unlimited rate limit" — both look like zero-value config.

The silent-allow default is already enforced inside `TokenBucketRateLimiter.Allow` itself
(lines 119–127): when `cfg.RequestsPerSecond <= 0`, it returns `Allowed: true` regardless.
`InMemoryRateLimitProvider` does not need to bake in the unlimited assumption.

---

## Fix

Return `ErrTenantNotFound` when the tenant has no config, matching the other providers.
`TokenBucketRateLimiter.Allow` already handles this correctly: it returns
`RateLimitResult{Allowed: false}, err` when `RateLimitConfig` returns an error.

Wait — that means fixing the provider will change behavior: currently unknown tenants are
**allowed** (unlimited), after the fix they will be **denied** (error from limiter).

To preserve "allow when unconfigured" at the *limiter* level without baking it into the
provider, `TokenBucketRateLimiter.Allow` must treat `ErrTenantNotFound` as "no config →
allow":

```go
cfg, err := l.provider.RateLimitConfig(ctx, tenantID)
if err != nil {
    if errors.Is(err, ErrTenantNotFound) {
        return RateLimitResult{Allowed: true}, nil
    }
    return RateLimitResult{Allowed: false}, err
}
```

This separates concerns cleanly:
- The **provider** answers "does this tenant have a rate limit config?" with a clear
  `ErrTenantNotFound` when it does not.
- The **limiter** decides what to do with an unconfigured tenant (allow, based on the
  "zero means unlimited" policy).

---

Scope:
- Fix `InMemoryRateLimitProvider.RateLimitConfig` to return `(RateLimitConfig{}, ErrTenantNotFound)` for absent tenants.
- Add the `errors.Is(err, ErrTenantNotFound)` branch to `TokenBucketRateLimiter.Allow`.
- Update `InMemoryRateLimitProvider` tests to assert `ErrTenantNotFound` for unknown tenants.
- Verify existing limiter tests for unknown tenants still pass (should now hit the new branch in `Allow`).

Non-goals:
- Do not change `RateLimiter` or `RateLimitConfigProvider` interfaces.
- Do not change `RateLimitConfig` zero-value semantics elsewhere.
- Do not change `InMemoryConfigManager.RateLimitConfig` behavior.

Files:
- `x/tenant/core/ratelimit.go`
- `x/tenant/core/ratelimit_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/...`
- `go test -race -timeout 60s ./x/tenant/...`
- `go vet ./x/tenant/...`

Docs Sync: —

Done Definition:
- `InMemoryRateLimitProvider.RateLimitConfig` returns `ErrTenantNotFound` for absent tenants.
- `TokenBucketRateLimiter.Allow` treats `ErrTenantNotFound` as "allow" (unlimited).
- Tests confirm: unknown tenant via `InMemoryRateLimitProvider` → allowed; limiter error from provider → denied.
- All tests pass.

Outcome:
