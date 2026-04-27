# Card 0039

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
Depends On:

Goal:
Converge `store/cache` miss and integer handling so the stable in-memory cache has one canonical miss sentinel and one local integer codec path.

Scope:
- Preserve `ErrCacheMiss` as an exported compatibility name but make it resolve to the canonical `ErrNotFound` sentinel.
- Extract local integer encode/decode helpers for `Incr` and `Decr` instead of open-coding gob handling.
- Add focused tests for `ErrCacheMiss` compatibility and integer value round trips.

Non-goals:
- Do not remove exported symbols.
- Do not change the `Cache` interface.
- Do not move integer operations to an extension package.

Files:
- store/cache/cache.go
- store/cache/cache_test.go

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Not required unless comments need to mention the compatibility alias.

Done Definition:
- `errors.Is(ErrCacheMiss, ErrNotFound)` and `errors.Is(ErrNotFound, ErrCacheMiss)` both work.
- Integer operations use one helper path and still reject non-integer bytes.
- Existing cache behavior remains compatible.

Outcome:
- Kept `ErrCacheMiss` exported while making it a compatibility alias for `ErrNotFound`.
- Extracted `encodeInt64` and `decodeInt64` helpers for memory-cache integer operations.
- Added tests for sentinel compatibility and integer payload round trips.

Validation:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache
