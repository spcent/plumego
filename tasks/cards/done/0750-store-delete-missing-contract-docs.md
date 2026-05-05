# Card 0750

Milestone:
Recipe: specs/change-recipes/doc-sync.yaml
Priority: P3
State: done
Primary Module: store
Owned Files:
- store/cache/cache.go
- store/kv/kv.go
- docs/modules/store/README.md

Goal:
Document stable delete-miss semantics for cache and KV primitives.

Scope:
- Clarify that cache delete is idempotent for missing keys.
- Clarify that KV delete returns `ErrKeyNotFound` for missing keys.
- Ensure code comments and store docs agree.

Non-goals:
- Do not change delete behavior.
- Do not modify tests unless comments expose drift.
- Do not add new errors.

Files:
- store/cache/cache.go
- store/kv/kv.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/cache ./store/kv
- go vet ./store/cache ./store/kv
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required.

Done Definition:
- Delete miss semantics are explicit in comments and module docs.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Documented idempotent missing-key delete semantics for `MemoryCache.Delete`.
- Documented `kv.ErrKeyNotFound` missing-key delete semantics for `KVStore.Delete` and `DeleteContext`.
- Synced store module docs.

Validation:
- `go test -timeout 20s ./store/cache ./store/kv`
- `go vet ./store/cache ./store/kv`
- `go run ./internal/checks/dependency-rules`
