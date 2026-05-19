# Card 0044

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
Depends On:
- 0043-store-db-resource-guardrails

Goal:
Align `store/cache` TTL and key-validation contract text with the actual stable cache behavior.

Scope:
- Make `Set` TTL documentation match the existing non-positive TTL fallback to `DefaultTTL`.
- Centralize expiration calculation used by `Set`.
- Remove tenant-oriented wording from stable cache key validation comments.
- Add focused coverage for negative per-entry TTL fallback to `DefaultTTL`.

Non-goals:
- Do not add tenant-aware cache behavior.
- Do not change the `Cache` interface.
- Do not change cache close semantics or add provider-specific features.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Not required; package comments only.

Done Definition:
- TTL comments and implementation use one clear helper path.
- Key validation comments remain stable-store scoped.
- Cache tests cover default TTL fallback for a negative explicit TTL.

Outcome:
- Centralized cache expiration calculation in `expirationForTTL`.
- Updated `Set` and key-validation comments to match stable cache behavior.
- Added negative-TTL default fallback coverage.

Validation:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache
