# Card 0767

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md
- docs/stable-api/snapshots/store-head.snapshot
Depends On: 0717

Goal:
Add context-aware KV operations and keep read-path expiry checks free of persistence side effects.

Scope:
- Add context-aware KV methods for Set, Get, Delete, Exists, Keys, and Size.
- Keep existing method names as source-compatible convenience calls.
- Make expired-key reads return expiry/not-found semantics without attempting disk persistence on the read path.
- Add focused tests for canceled contexts and read-only expired access.
- Update store docs and the stable API snapshot.

Non-goals:
- Do not add WAL, snapshots, serializer selection, compression, or shard tuning.
- Do not introduce a new durable KV engine.
- Do not add tenant-aware storage policy.

Files:
- `store/kv/kv.go`
- `store/kv/kv_test.go`
- `docs/modules/store/README.md`
- `docs/stable-api/snapshots/store-head.snapshot`

Tests:
- `go test -race -timeout 60s ./store/kv`
- `go test -timeout 20s ./store/...`
- `go vet ./store/...`

Docs Sync:
- Required for KV context and read-path semantics.

Done Definition:
- Context-aware KV methods reject canceled contexts before taking the store lock.
- Existing KV methods remain source-compatible.
- Expired reads do not persist cleanup as a side effect.
- Targeted store tests, vet, and the store API snapshot are updated.

Outcome:
- Added context-aware `SetContext`, `GetContext`, `DeleteContext`, `ExistsContext`, `KeysContext`, and `SizeContext` methods.
- Kept existing non-context methods as source-compatible convenience wrappers.
- Changed expired `Get` behavior to return `ErrKeyExpired` without deleting or persisting cleanup on the read path.
- Added canceled-context coverage for all new context-aware methods and updated expired-read coverage.
- Synced the store module primer and stable store API snapshot.
- Validation run: `go test -race -timeout 60s ./store/kv`; `go test -timeout 20s ./store/...`; `go vet ./store/...`; `go run ./internal/checks/dependency-rules`.
