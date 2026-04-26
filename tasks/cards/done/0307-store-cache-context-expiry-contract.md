# Card 0307

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
Depends On:
- 0306-store-cache-lifecycle-atomicity

Goal:
Make the `context.Context` parameter and expiration checks in `store/cache` meaningful and consistent.

Scope:
- Return caller-owned context cancellation/deadline errors before mutating or reading cache state.
- Normalize cache expiration checks so boundary-time expiry is treated consistently across reads, existence checks, cleanup, and atomic operations.
- Add focused tests for canceled contexts and zero/expired timestamp behavior.

Non-goals:
- Do not introduce timeout policy or background deadline ownership in `store/cache`.
- Do not change the stable cache interface.
- Do not add observability or metrics hooks.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Not required; this aligns implementation with the existing context-bearing API.

Done Definition:
- Canceled or expired contexts are returned by cache operations before state changes.
- Expired cache items are handled through one canonical predicate.
- Existing cache behavior remains compatible for active contexts.

Outcome:
- Added a shared context error guard to cache operations before reads or mutations.
- Normalized expiration through `expiredAt`, treating exact expiration time as expired.
- Added canceled-context coverage across cache operations and boundary coverage for expiration checks.

Validation:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache
