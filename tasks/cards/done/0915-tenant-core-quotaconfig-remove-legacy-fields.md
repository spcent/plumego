# Card 0915

Priority: P2
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/quota.go`
- `x/tenant/core/quota_limits.go`
- `x/tenant/core/quota_sliding.go`
Depends On: —

Goal:
- Remove the legacy `RequestsPerMinute` / `TokensPerMinute` shorthand fields from
  `QuotaConfig` and require callers to use `Limits []QuotaLimit` exclusively,
  eliminating the dual-path confusion.

---

## Problem

**File:** `x/tenant/core/quota.go` lines 12–22

```go
type QuotaConfig struct {
    RequestsPerMinute int   // ← legacy shorthand
    TokensPerMinute   int   // ← legacy shorthand
    // When non-empty, Limits take precedence over RequestsPerMinute/TokensPerMinute.
    Limits []QuotaLimit
}
```

`QuotaConfig` exposes two ways to express the same per-minute quota:

```go
// Style A — shorthand (legacy)
QuotaConfig{RequestsPerMinute: 100, TokensPerMinute: 5000}

// Style B — explicit (current)
QuotaConfig{Limits: []QuotaLimit{{Window: QuotaWindowMinute, Requests: 100, Tokens: 5000}}}
```

`normalizeQuotaLimits` bridges them, but:
1. A caller who sets both gets silently different precedence than expected ("Limits wins").
2. `InMemoryQuotaManager` documents that it "enforces only the first valid limit window"
   but accepts the full `Limits` slice — callers who set hour/day/month limits and use
   `InMemoryQuotaManager` get silent single-window enforcement.
3. The shorthand only covers minutes; it cannot express hour/day/month limits at all,
   so any caller who needs multi-window quotas must switch to `Limits` anyway.

The shorthand fields exist only because the package evolved iteratively. They add no
capability that `Limits` does not already provide, and they create a misleading dual-path
that requires reading `normalizeQuotaLimits` to understand the precedence rules.

---

## Fix

**Remove `RequestsPerMinute` and `TokensPerMinute` from `QuotaConfig`:**

```go
type QuotaConfig struct {
    // Limits defines per-tenant quota windows. Zero values mean unlimited for that dimension.
    // Supported windows: QuotaWindowMinute, QuotaWindowHour, QuotaWindowDay, QuotaWindowMonth.
    Limits []QuotaLimit
}
```

**Simplify `normalizeQuotaLimits`** (`quota_limits.go`):

```go
func normalizeQuotaLimits(cfg QuotaConfig) []QuotaLimit {
    limits := make([]QuotaLimit, 0, len(cfg.Limits))
    for _, limit := range cfg.Limits {
        if isValidQuotaWindow(limit.Window) {
            limits = append(limits, limit)
        }
    }
    return limits
}
```

**Migrate all internal callers** of the shorthand fields:
- `grep -rn 'RequestsPerMinute\|TokensPerMinute' . --include='*.go'` to find every reference.
- Update test fixtures, config examples, and any in-package construction.

**Simplify `slidingWindowEffectiveLimits`** (`quota_sliding.go`):

The special-case path for `cfg.RequestsPerMinute / cfg.TokensPerMinute` in
`slidingWindowEffectiveLimits` can be removed; the function becomes a straight scan
of `cfg.Limits` for `QuotaWindowMinute`.

---

Scope:
- Delete `RequestsPerMinute` and `TokensPerMinute` fields.
- Simplify `normalizeQuotaLimits` and `slidingWindowEffectiveLimits`.
- Migrate all test fixtures and callers to `Limits`.
- No behavior change for correctly-migrated callers.

Non-goals:
- Do not add new quota window types.
- Do not change `QuotaLimit`, `QuotaWindow`, or any manager interface.
- Do not change manager selection guidance (InMemory vs Window vs Sliding).

Files:
- `x/tenant/core/quota.go`
- `x/tenant/core/quota_limits.go`
- `x/tenant/core/quota_sliding.go`
- `x/tenant/core/quota_test.go`
- `x/tenant/core/quota_sliding_test.go`
- `x/tenant/core/quota_window_test.go`
- `x/tenant/core/config.go` (if Config.Quota is populated with shorthand fields in tests)
- `x/tenant/core/config_test.go`

Tests:
- `grep -rn 'RequestsPerMinute\|TokensPerMinute' . --include='*.go'` returns empty after migration.
- `go build ./...`
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`

Docs Sync:
- Update any inline comments in `quota.go` that describe the shorthand fallback logic.

Done Definition:
- `QuotaConfig` has only `Limits []QuotaLimit`.
- `normalizeQuotaLimits` has no shorthand fallback branch.
- `slidingWindowEffectiveLimits` scans `cfg.Limits` directly.
- `grep -rn 'RequestsPerMinute\|TokensPerMinute' . --include='*.go'` returns empty.
- All tests pass.

Outcome:
