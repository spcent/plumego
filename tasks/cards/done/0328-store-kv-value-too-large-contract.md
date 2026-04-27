# Card 0328

Milestone:
Recipe: specs/change-recipes/refine-api.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:
- 0327-store-db-rollback-error-propagation

Goal:
Give oversized KV writes a stable, namespaced sentinel error.

Scope:
- Add an oversized-value sentinel for values exceeding configured memory.
- Wrap oversized write errors with the sentinel while preserving size details.
- Add focused tests for sentinel matching and message shape.

Non-goals:
- Do not change memory accounting, eviction, or storage format.
- Do not add WAL, shard, serializer, or durable-engine behavior.
- Do not change `Options`.

Files:
- store/kv/kv.go
- store/kv/kv_test.go

Tests:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv

Docs Sync:
- Not required.

Done Definition:
- Oversized writes match the new sentinel and remain clearly `kv:` namespaced.
- Existing rollback and eviction behavior remains unchanged.
- KV targeted tests and vet pass.

Outcome:
- Added `ErrValueTooLarge` for oversized KV writes.
- Wrapped oversized write failures with size and limit details.
- Added sentinel and message-shape coverage.

Validation:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv
