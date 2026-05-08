# Card 1089

Milestone:
Recipe: specs/change-recipes/feature.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md

Goal:
Add a minimal stable stats snapshot for `store/cache.MemoryCache`.

Scope:
- Add a `Stats` value type and `Stats()` method for entries, memory usage, and closed state.
- Keep the existing `Cache` interface unchanged.
- Add tests proving snapshots are point-in-time values.
- Sync store docs.

Non-goals:
- Do not add metrics exporters or instrumentation.
- Do not add provider-specific cache stats.
- Do not change eviction behavior.

Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/cache
- go vet ./store/cache
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required for new public method.

Done Definition:
- `MemoryCache` exposes stable entries/memory/closed snapshot.
- `Cache` interface remains narrow and unchanged.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Added `cache.Stats` and `MemoryCache.Stats()` for entries, tracked payload bytes, and closed lifecycle state.
- Kept the `Cache` interface unchanged.
- Documented the snapshot semantics in store docs.

Validation:
- `go test -timeout 20s ./store/cache`
- `go vet ./store/cache`
- `go run ./internal/checks/dependency-rules`
