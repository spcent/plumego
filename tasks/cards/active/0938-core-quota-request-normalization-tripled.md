# Card 0938

Priority: P1
State: active
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/quota.go`
- `x/tenant/core/quota_window.go`
- `x/tenant/core/quota_sliding.go`
- `x/tenant/core/quota_limits.go`
Depends On:

Goal:
- Extract `QuotaRequest` normalization into a single helper and fix the UTC inconsistency in `SlidingWindowQuotaManager` so all three quota managers share one normalization path.

Problem:
- Three managers independently normalize the same two `QuotaRequest` fields before any business logic:

  **`FixedWindowQuotaManager.Allow`** (`quota.go:87-91`):
  ```go
  if req.Now.IsZero() { req.Now = time.Now().UTC() }
  if req.Requests <= 0 { req.Requests = 1 }
  ```

  **`WindowQuotaManager.Allow`** (`quota_window.go:39-43`):
  ```go
  if req.Now.IsZero() { req.Now = time.Now().UTC() }
  if req.Requests <= 0 { req.Requests = 1 }
  ```

  **`SlidingWindowQuotaManager.Allow`** (`quota_sliding.go:67-72`):
  ```go
  now := req.Now
  if now.IsZero() { now = time.Now() }   // ← missing .UTC()
  if req.Requests <= 0 { req.Requests = 1 }
  ```

- Three problems at once:
  1. The same normalization is copy-pasted in three places — any future change (e.g., adding a `Tokens` floor) must be applied three times.
  2. `SlidingWindowQuotaManager` calls `time.Now()` without `.UTC()`, while the other two managers call `time.Now().UTC()`. Time comparisons inside `SlidingWindowQuotaManager` (`cleanup`, `calculateRequestRetryAfter`) against the now-local `now` will silently produce wrong `RetryAfter` values in non-UTC deployments.
  3. `SlidingWindowQuotaManager` assigns the default time to a local `now` variable instead of mutating `req.Now`, creating a third code path where the caller-supplied `req.Now` is structurally honored but the field value used internally differs from what `FixedWindowQuotaManager` and `WindowQuotaManager` would store back.

Scope:
- Add a `normalizeQuotaRequest` function (or method on `QuotaRequest`) that:
  - Sets `req.Now = time.Now().UTC()` if zero.
  - Sets `req.Requests = 1` if `<= 0`.
- Replace the three inline normalization blocks with a single call.
- In `SlidingWindowQuotaManager`, eliminate the local `now` variable; use `req.Now` after normalization, consistent with the other managers.

Non-goals:
- Do not change the `QuotaRequest` struct fields or their types.
- Do not normalize `req.Tokens` (zero tokens is valid — it means "count requests only").
- Do not change `QuotaManager` interface.

Files:
- `x/tenant/core/quota_limits.go` (add `normalizeQuotaRequest`)
- `x/tenant/core/quota.go`
- `x/tenant/core/quota_window.go`
- `x/tenant/core/quota_sliding.go`
- `x/tenant/core/quota_test.go`
- `x/tenant/core/quota_sliding_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/core/...`
- `go vet ./x/tenant/core/...`

Docs Sync:
- None required.

Done Definition:
- A single `normalizeQuotaRequest` helper exists and is called by all three managers.
- `SlidingWindowQuotaManager` uses UTC time, consistent with the other managers.
- All tests pass.

Outcome:
- Pending.
