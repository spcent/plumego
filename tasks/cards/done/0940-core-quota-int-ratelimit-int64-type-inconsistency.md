# Card 0940

Priority: P1
State: active
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/quota_limits.go`
- `x/tenant/core/quota.go`
- `x/tenant/core/quota_store.go`
- `x/tenant/core/ratelimit.go`
Depends On: 0925

Goal:
- Align the numeric types for request/token counts across the quota and rate limit subsystems: both should use the same integer width for the same semantic concept.

Problem:
- Quota types use `int` throughout:
  - `QuotaLimit.Requests int`, `QuotaLimit.Tokens int` (`quota_limits.go:19-20`)
  - `QuotaRequest.Requests int`, `QuotaRequest.Tokens int` (`quota.go:28-29`)
  - `QuotaResult.RemainingRequests int`, `QuotaResult.RemainingTokens int` (`quota.go:36-37`)
  - `QuotaStore.Reserve` `deltaRequests`, `deltaTokens`, `limitRequests`, `limitTokens` all `int` (`quota_store.go:22`)

- Rate limit types use `int64` throughout:
  - `RateLimitConfig.RequestsPerSecond int64`, `Burst int64` (`ratelimit.go:16-19`)
  - `RateLimitRequest.Tokens int64` (`ratelimit.go:25`)
  - `RateLimitResult.Remaining int64`, `Limit int64`, `Burst int64` (`ratelimit.go:33-36`)

- Both domains represent request and token counts. Using different integer widths for the same concept creates friction:
  - Code that converts quota remaining values to rate limit result values (e.g., in a combined middleware or hook) needs explicit casts: `int64(quotaResult.RemainingRequests)` vs the rate limit's `int64`.
  - `QuotaDecision.Tokens int` and `RateLimitDecision.Tokens int64` (`hooks.go:42` and `hooks.go:59`) already show this split at the hook layer — an observer comparing the two decision types must handle different widths.
  - On 32-bit platforms `int` is 32 bits (max ~2 billion), sufficient for any realistic quota. But using a different width from rate limits purely for historical reasons is a maintenance cost.

Scope:
- Align quota numeric types to match rate limits: change `int` → `int64` for all request/token count fields in quota types:
  - `QuotaLimit.Requests`, `QuotaLimit.Tokens`
  - `QuotaRequest.Requests`, `QuotaRequest.Tokens`
  - `QuotaResult.RemainingRequests`, `QuotaResult.RemainingTokens`
  - `QuotaUsage.Requests`, `QuotaUsage.Tokens`
  - `QuotaDecision.Tokens`, `QuotaDecision.Requests`, etc. in `hooks.go`
  - `QuotaStore.Reserve`/`Release` parameters (coordinate with card 0925)
  - The `remaining()` helper in `quota.go`
- Update all call sites within the package.

Non-goals:
- Do not change `RateLimitConfig` or `RateLimitResult` types.
- Do not change `QuotaWindow` or any window-boundary types.
- If card 0925 has already replaced QuotaStore.Reserve with a struct, update that struct's fields in this card.

Files:
- `x/tenant/core/quota_limits.go`
- `x/tenant/core/quota.go`
- `x/tenant/core/quota_store.go`
- `x/tenant/core/quota_window.go`
- `x/tenant/core/quota_sliding.go`
- `x/tenant/core/hooks.go`
- `x/tenant/core/quota_test.go`
- `x/tenant/core/quota_sliding_test.go`
- `x/tenant/core/quota_window_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/core/...`
- `go vet ./x/tenant/core/...`

Docs Sync:
- None required.

Done Definition:
- All request/token count fields in quota types use `int64`.
- `QuotaDecision` and `RateLimitDecision` in `hooks.go` use the same integer width for token counts.
- All tests pass.

Outcome:
- Pending.
