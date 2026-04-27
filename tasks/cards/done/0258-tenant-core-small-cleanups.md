# Card 0258

Priority: P3
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/context.go`
- `x/tenant/core/quota_sliding.go`
Depends On: —

Goal:
- Fix two small but concrete issues in `x/tenant/core` that create either a latent panic
  or unnecessary noise relative to the rest of the codebase.

---

## Issue A — `tenantIDContextKeyVar` is an unnecessary singleton

**File:** `x/tenant/core/context.go` lines 8–10

```go
type tenantIDContextKey struct{}

var tenantIDContextKeyVar tenantIDContextKey  // ← redundant
```

Every other context key in the repo (e.g. `contract/request_id.go`, `contract/trace.go`,
`contract/auth.go`) uses the zero-value struct literal directly:

```go
context.WithValue(ctx, requestIDKey{}, id)
ctx.Value(requestIDKey{})
```

`x/tenant/core` deviates by storing the key in a package-level var and using the var
everywhere. The var adds nothing — the struct type is already the unique key. Keeping it
implies the pattern is intentional or required, which misleads readers.

**Fix:**
- Delete `var tenantIDContextKeyVar tenantIDContextKey`.
- Replace every `tenantIDContextKeyVar` reference with `tenantIDContextKey{}`.
- Confirm `grep -n 'tenantIDContextKeyVar' x/tenant/core/context.go` returns empty.

---

## Issue B — `SlidingWindowQuotaManager.Allow` has no nil receiver or nil provider guard

**File:** `x/tenant/core/quota_sliding.go` lines 48–52

```go
func (m *SlidingWindowQuotaManager) Allow(ctx context.Context, tenantID string, req QuotaRequest) (QuotaResult, error) {
    cfg, err := m.provider.QuotaConfig(ctx, tenantID)  // ← panics if m == nil or m.provider == nil
```

Every other `Allow` implementation in the package guards against this:

| Type | Guard |
|------|-------|
| `InMemoryQuotaManager` | `if m == nil \|\| m.provider == nil { return QuotaResult{Allowed: true}, nil }` |
| `WindowQuotaManager` | `if m == nil \|\| m.provider == nil \|\| m.store == nil { ... }` |
| `TokenBucketRateLimiter` | `if l == nil \|\| l.provider == nil { return RateLimitResult{Allowed: true}, nil }` |
| `ConfigPolicyEvaluator` | `if e == nil \|\| e.Provider == nil { return PolicyResult{Allowed: true}, nil }` |

`SlidingWindowQuotaManager` is the only one that skips the guard. A nil provider
(misconfigured wiring) causes a nil-pointer panic instead of the consistent "allow when
unconfigured" behavior that other managers implement.

**Fix:** Add a nil guard at the top of `SlidingWindowQuotaManager.Allow`:

```go
func (m *SlidingWindowQuotaManager) Allow(ctx context.Context, tenantID string, req QuotaRequest) (QuotaResult, error) {
    if m == nil || m.provider == nil {
        return QuotaResult{Allowed: true}, nil
    }
    ...
```

---

Scope:
- Apply both fixes; they are independent and each touches a handful of lines.
- No behavior change for correctly-wired callers.

Non-goals:
- Do not change QuotaManager interface.
- Do not change the "allow when unconfigured" policy itself.
- Do not add context key helpers to a shared package.

Files:
- `x/tenant/core/context.go`
- `x/tenant/core/quota_sliding.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./x/tenant/...`
- `go test -race -timeout 60s ./x/tenant/...`
- `go vet ./x/tenant/...`

Docs Sync: —

Done Definition:
- `var tenantIDContextKeyVar tenantIDContextKey` is deleted; all usages replaced with `tenantIDContextKey{}`.
- `SlidingWindowQuotaManager.Allow` returns `QuotaResult{Allowed: true}, nil` when receiver or provider is nil.
- `grep -n 'tenantIDContextKeyVar' x/tenant/core/context.go` returns empty.
- All tests pass.

Outcome:
