# Card 0723

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md
- docs/stable-api/snapshots/store-head.snapshot
Depends On: 0722

Goal:
Make `store/cache.MemoryCache.Close` a strict terminal lifecycle boundary with no writes after close returns.

Scope:
- Serialize public cache operations with close so operations that start after close fail with `ErrClosed`.
- Prevent expired-entry cleanup from mutating state after the cache is closed.
- Add focused close-race and post-close behavior tests.
- Update store docs and snapshot if exported surface changes.

Non-goals:
- Do not add cache metrics or introspection APIs.
- Do not add distributed or provider-specific cache behavior.
- Do not change cache key/value wire formats.

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
- Required for close lifecycle semantics.

Done Definition:
- Close waits for in-flight cache mutations and blocks later mutations.
- Public operations return `ErrClosed` once close wins the lifecycle boundary.
- Race tests cover close interaction.
- Targeted store tests, vet, and snapshot are updated.

Outcome:
- Serialized public cache operations with the close lifecycle boundary and rechecked `ErrClosed` after entering the write boundary.
- Made expired-entry cleanup avoid mutating state once the cache is closed.
- Added a close-boundary test that proves `Close` waits for the write boundary before returning.
- Synced the store module README and stable store API snapshot.
- Validation run: `go test -race -timeout 60s ./store/cache`; `go test -timeout 20s ./store/...`; `go vet ./store/...`; `go run ./internal/checks/dependency-rules`.
