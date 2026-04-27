# Card 0275

Priority: P1
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/quota_store.go`
- `x/tenant/core/quota_window.go`
Depends On:

Goal:
- Replace the 8-parameter `QuotaStore.Reserve()` signature and 6-parameter `Release()` signature with dedicated request structs to make the interface safe and readable to call.

Problem:
- `QuotaStore.Reserve()` (`quota_store.go:22`) takes 8 positional parameters:
  `(ctx, tenantID, window, windowStart, deltaRequests, deltaTokens, limitRequests, limitTokens int)`
- `QuotaStore.Release()` (`quota_store.go:24`) takes 6 positional parameters:
  `(ctx, tenantID, window, windowStart, deltaRequests, deltaTokens int)`
- Positional integer parameters at this count are extremely easy to swap silently. `deltaRequests` and `limitRequests` look identical at call sites.
- `WindowQuotaManager.Allow()` (`quota_window.go:63`) calls `Reserve` with all 8 arguments spread across 8 lines; a single transposition is undetectable by the compiler.
- Every other decision API in the package (policy, quota, ratelimit) uses a request struct: `PolicyRequest`, `QuotaRequest`, `RateLimitRequest`. The `QuotaStore` is the only surface that does not follow this pattern.

Scope:
- Introduce `QuotaReserveRequest` struct with fields: `TenantID`, `Window`, `WindowStart`, `DeltaRequests`, `DeltaTokens`, `LimitRequests`, `LimitTokens`.
- Introduce `QuotaReleaseRequest` struct with fields: `TenantID`, `Window`, `WindowStart`, `DeltaRequests`, `DeltaTokens`.
- Update `QuotaStore` interface to use these structs:
  `Reserve(ctx, req QuotaReserveRequest) (QuotaUsage, bool, error)`
  `Release(ctx, req QuotaReleaseRequest) error`
- Update `InMemoryQuotaStore` implementation accordingly.
- Update `WindowQuotaManager.Allow()` call sites to construct the structs.

Non-goals:
- Do not change `QuotaUsage` return shape.
- Do not change `QuotaManager` interface or `FixedWindowQuotaManager`.
- Do not change `SlidingWindowQuotaManager` (it does not use `QuotaStore`).

Files:
- `x/tenant/core/quota_store.go`
- `x/tenant/core/quota_window.go`

Tests:
- `go test -timeout 20s ./x/tenant/core/...`
- `go vet ./x/tenant/core/...`

Docs Sync:
- None required.

Done Definition:
- `QuotaStore.Reserve` and `QuotaStore.Release` accept request structs instead of positional parameters.
- `InMemoryQuotaStore` and `WindowQuotaManager` are updated with no behavior change.
- All tests pass.

Outcome:
- Pending.
