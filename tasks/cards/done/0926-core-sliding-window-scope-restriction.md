# Card 0926

Priority: P1
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/quota_sliding.go`
Depends On:

Goal:
- Make `SlidingWindowQuotaManager`'s per-minute-only constraint visible at the interface boundary so callers cannot silently misconfigure it with non-minute windows.

Problem:
- `SlidingWindowQuotaManager` implements `QuotaManager` and is passed `QuotaConfig` which can contain any combination of `QuotaWindowMinute`, `QuotaWindowHour`, `QuotaWindowDay`, `QuotaWindowMonth` limits.
- In practice, the manager only inspects the `QuotaWindowMinute` entry (`slidingWindowEffectiveLimits`, `quota_sliding.go:192`). All other window types are silently ignored and the call returns `Allowed: true`.
- The sliding window cleanup cutoff is hardcoded at `now.Add(-time.Minute)` (`quota_sliding.go:78`), and retry-after calculations also add `time.Minute` (`quota_sliding.go:156`, `quota_sliding.go:181`). This hardcoding makes the window duration impossible to change without rewriting the algorithm.
- A caller who configures `QuotaConfig{Limits: []QuotaLimit{{Window: QuotaWindowHour, Requests: 1000}}}` and uses `SlidingWindowQuotaManager` will receive `Allowed: true` for every request, with no error or warning.
- The `QuotaManager` interface comment at `quota.go:42` gives no indication that a concrete implementation may only support a subset of window types.

Scope:
- In `SlidingWindowQuotaManager.Allow()`, if the provider config contains limits for windows other than `QuotaWindowMinute`, return an error (e.g., `ErrUnsupportedQuotaWindow`) rather than silently allowing.
- Add `ErrUnsupportedQuotaWindow` to the package sentinel errors (see card 0922).
- Document in the `SlidingWindowQuotaManager` type comment that it only enforces `QuotaWindowMinute` and that `WindowQuotaManager` should be used for hour/day/month limits.
- The type-level doc block already mentions the minute-only scope (`quota_sliding.go:14`); strengthen it to be explicit about the error behavior when other windows are configured.

Non-goals:
- Do not extend `SlidingWindowQuotaManager` to support hour/day/month windows; use `WindowQuotaManager` for that.
- Do not change `WindowQuotaManager` or `FixedWindowQuotaManager`.
- Do not change `QuotaManager` interface.

Files:
- `x/tenant/core/quota_sliding.go`
- `x/tenant/core/errors.go` (add `ErrUnsupportedQuotaWindow` — coordinate with card 0922)
- `x/tenant/core/quota_sliding_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/core/...`
- `go vet ./x/tenant/core/...`

Docs Sync:
- None required.

Done Definition:
- `SlidingWindowQuotaManager.Allow()` returns `ErrUnsupportedQuotaWindow` when the tenant config includes any window other than `QuotaWindowMinute`.
- The type comment clearly states the minute-only scope and the error behavior.
- A test covers the misconfiguration path.

Outcome:
- Pending.
