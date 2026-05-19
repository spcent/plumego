# Card 0040

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
Depends On:
- 0039-store-cache-miss-and-integer-codec

Goal:
Align `store/cache` configuration validation and comments with the actual default TTL behavior.

Scope:
- Reject negative `DefaultTTL` values during config validation.
- Reword `Set` documentation and stale tests/comments so zero or negative explicit TTL is documented as using `DefaultTTL` when configured.
- Remove unused private cache item fields that make the implementation look broader than it is.
- Add focused tests for negative `DefaultTTL`.

Non-goals:
- Do not change default TTL behavior.
- Do not add close-state enforcement or cache metrics.
- Do not add provider-specific behavior.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Inline package/test comments only.

Done Definition:
- Invalid negative default TTL cannot construct a `MemoryCache`.
- Comments describe implemented TTL behavior.
- Private cache item shape has no unused fields.

Outcome:
- Rejected negative `DefaultTTL` in `Config.Validate`.
- Updated TTL comments/tests so zero TTL behavior reflects `DefaultTTL`.
- Removed the unused private `cacheItem.key` field.

Validation:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache
