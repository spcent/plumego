# Card 0270

Priority: P1
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/ratelimit.go`
Depends On:

Goal:
- Rename `InMemoryRateLimitProvider` to `InMemoryRateLimitManager` so the type name matches its actual role: a mutable, stateful store, not a read-only provider.

Problem:
- The package uses a clear naming convention: types that only expose read methods are "Providers", stateful types that also mutate are "Managers".
- `InMemoryRateLimitProvider` (`ratelimit.go:50`) breaks this rule. It exposes `SetRateLimit()` (`ratelimit.go:63`), a write operation, making it a Manager.
- The read-only interface it satisfies is correctly named `RateLimitConfigProvider` (`ratelimit.go:39`). The concrete implementation should be named `InMemoryRateLimitManager` to match `InMemoryConfigManager` and `FixedWindowQuotaManager`.
- `NewInMemoryRateLimitProvider()` (`ratelimit.go:56`) follows the same wrong pattern.
- This inconsistency makes it unclear whether the type can be mutated and misleads callers about what they should inject.

Scope:
- Rename `InMemoryRateLimitProvider` → `InMemoryRateLimitManager` across the package.
- Rename `NewInMemoryRateLimitProvider()` → `NewInMemoryRateLimitManager()`.
- Update all references inside the package and in tests.
- No behavior change.

Non-goals:
- Do not rename the interface `RateLimitConfigProvider` (it is correctly named).
- Do not change `TokenBucketRateLimiter` (it is a specific algorithm name, not a Manager/Provider role).
- Do not touch anything outside `x/tenant/core`.

Files:
- `x/tenant/core/ratelimit.go`
- `x/tenant/core/ratelimit_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/core/...`
- `go vet ./x/tenant/core/...`

Docs Sync:
- None required.

Done Definition:
- No type or constructor named `InMemoryRateLimitProvider` or `NewInMemoryRateLimitProvider` exists in the package.
- `InMemoryRateLimitManager` and `NewInMemoryRateLimitManager` replace them with identical behavior.
- All tests pass.

Outcome:
- Pending.
