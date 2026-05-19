# Card 0049

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
Depends On:
- 0048-store-file-type-contract-docs

Goal:
Preserve existing cache entry expiration when mutation operations update an unexpired value.

Scope:
- Ensure `Append` preserves an existing unexpired entry's expiration.
- Ensure `Incr` preserves an existing unexpired entry's expiration.
- Keep missing-key mutation behavior unchanged, including default TTL application.
- Add focused tests for expiration preservation.

Non-goals:
- Do not change the `Cache` interface.
- Do not add distributed cache, metrics, or HTTP response-cache behavior.
- Do not change close semantics.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Not required.

Done Definition:
- Mutating an existing unexpired cache key does not silently extend or replace its expiration.
- Missing-key `Append` and `Incr` still use the default TTL path.
- Cache targeted tests and vet pass.

Outcome:
- Added an internal expiration-preserving store helper.
- Updated `Incr` and `Append` to preserve existing unexpired entry expirations.
- Added focused tests for `Incr` and `Append` expiration preservation.

Validation:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache
