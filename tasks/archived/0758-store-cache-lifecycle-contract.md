# Card 0758

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md
- docs/stable-api/snapshots/store-head.snapshot
Depends On: 0716

Goal:
Make the in-memory cache lifecycle explicit so `Close` means closed rather than only stopping cleanup.

Scope:
- Add a cache closed sentinel error.
- Make cache operations return the closed sentinel after `Close`.
- Keep `Close` idempotent.
- Add focused tests for post-close behavior.
- Update store docs and the stable API snapshot.

Non-goals:
- Do not add distributed cache behavior, metrics ownership, or HTTP response caching.
- Do not change cache key, TTL, integer, or append semantics except for closed lifecycle.
- Do not introduce new dependencies.

Files:
- `store/cache/cache.go`
- `store/cache/cache_test.go`
- `docs/modules/store/README.md`
- `docs/stable-api/snapshots/store-head.snapshot`

Tests:
- `go test -race -timeout 60s ./store/cache`
- `go test -timeout 20s ./store/...`
- `go vet ./store/...`

Docs Sync:
- Required for cache lifecycle behavior.

Done Definition:
- Closed caches reject Get/Set/Delete/Exists/Clear/Incr/Decr/Append with `ErrClosed`.
- Repeated `Close` remains nil.
- Targeted store tests, vet, and the store API snapshot are updated.

Outcome:
- Added `cache.ErrClosed` and a closed lifecycle flag to `MemoryCache`.
- Made `Get`, `Set`, `Delete`, `Exists`, `Clear`, `Incr`, `Decr`, and `Append` reject post-close operations with `ErrClosed`.
- Kept `Close` idempotent.
- Added focused post-close operation coverage.
- Synced the store module primer and stable store API snapshot.
- Validation run: `go test -race -timeout 60s ./store/cache`; `go test -timeout 20s ./store/...`; `go vet ./store/...`.
