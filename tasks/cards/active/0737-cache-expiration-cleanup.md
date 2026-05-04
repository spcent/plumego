# Card 0737: Cache Expiration Cleanup

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P3
State: active
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md
- docs/stable-api/snapshots/store-head.snapshot
Depends On:
- 0736

Goal:
Make memory cache expiration cleanup predictable enough for stable in-process use.

Scope:
- Avoid cleanup behavior that can indefinitely miss expired entries.
- Keep cleanup bounded or explicitly configurable without broadening cache ownership.
- Add tests for large expired sets and memory accounting cleanup.
- Refresh store API snapshot if the public surface changes.

Non-goals:
- Do not add cache metrics or introspection ownership to stable store.
- Do not add HTTP response cache helpers.
- Do not add external dependencies.

Files:
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md
- docs/stable-api/snapshots/store-head.snapshot

Tests:
- go test -timeout 20s ./store/cache
- go test -race -timeout 60s ./store/cache
- go vet ./store/cache

Docs Sync:
- Required if cleanup configuration or behavior is documented.

Done Definition:
- Expired entries cannot remain indefinitely due only to cleanup scan limits.
- Memory accounting is covered by tests.
- Targeted tests, race tests, vet, and snapshot refresh if needed pass.

Outcome:
