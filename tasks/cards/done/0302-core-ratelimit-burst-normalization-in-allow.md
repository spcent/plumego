# Card 0302

Priority: P1
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/ratelimit.go`
Depends On:

Goal:
- Move `RateLimitConfig.Burst` defaulting logic out of `TokenBucketRateLimiter.Allow` and into the config validation boundary so callers see consistent effective values and the hot path is free of config mutation.

Problem:
- `TokenBucketRateLimiter.Allow` (`ratelimit.go:131-133`) mutates the config it fetched from the provider:
  ```go
  if cfg.Burst <= 0 {
      cfg.Burst = cfg.RequestsPerSecond
  }
  ```
- This normalization runs on every request, on every call into `Allow`. The provider returns a fresh `RateLimitConfig` struct each time; the mutation is thrown away after the call returns.
- The mutation silently changes the effective burst capacity from 0 to `RequestsPerSecond`. Any tooling that reads `RateLimitConfig.Burst` (logs, hooks, introspection) will see 0, not the actual effective value.
- `RateLimitResult.Burst` (`ratelimit.go:190`) is populated from `cfg.Burst` AFTER the mutation (`return RateLimitResult{..., Burst: cfg.Burst}`), so the result reflects the normalized value. But the stored config does not. This inconsistency can mislead dashboards and `RateLimitDecision` observers.
- There is no parallel behavior in `QuotaConfig`: if `Limits` is empty, quota managers return "allow all" directly — they don't mutate the config struct. The ratelimit approach is inconsistent with this pattern.

Scope:
- Add a `Validate()` or `Effective()` method (or unexported helper) to `RateLimitConfig` that returns the effective burst:
  - If `Burst > 0`, use it as-is.
  - If `Burst <= 0` and `RequestsPerSecond > 0`, effective burst = `RequestsPerSecond`.
  - If both are 0, unlimited (no enforcement).
- Use this helper inside `TokenBucketRateLimiter.Allow` instead of mutating `cfg`.
- Update `RateLimitResult.Burst` to reflect the effective value.
- The `InMemoryRateLimitManager` `SetRateLimit` method (after 0923 rename) can call `Validate()` to reject clearly invalid configs at storage time.

Non-goals:
- Do not change `RateLimitConfig` field names.
- Do not change the `RateLimitConfigProvider` interface.
- Do not change `RateLimitRequest.Tokens` normalization (stays in `Allow` — it is request-time, not config-time).

Files:
- `x/tenant/core/ratelimit.go`
- `x/tenant/core/ratelimit_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/core/...`
- `go vet ./x/tenant/core/...`

Docs Sync:
- None required.

Done Definition:
- `TokenBucketRateLimiter.Allow` does not mutate the `RateLimitConfig` struct.
- The effective burst value is derived from a single helper and reflected consistently in `RateLimitResult.Burst`.
- All tests pass.

Outcome:
- Pending.
