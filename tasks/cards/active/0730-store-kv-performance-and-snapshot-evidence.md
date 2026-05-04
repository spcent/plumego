# Card 0730: Store KV Performance And Snapshot Evidence

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md
- docs/stable-api/snapshots/store-head.snapshot
Depends On:
- 0729

Goal:
Make the `store/kv` performance envelope measurable and keep stable API evidence synchronized after the store changes.

Scope:
- Add focused benchmarks for `Set`, `Get`, `Delete`, and eviction under small embedded datasets.
- Tune defaults or documentation if benchmarks show the current defaults misrepresent the intended small-state-file usage.
- Refresh `docs/stable-api/snapshots/store-head.snapshot`.
- Run store and boundary checks.

Non-goals:
- Do not add WAL, snapshots, serializers, compression, sharding, or cross-process locks to stable `store/kv`.
- Do not change the public `KVStore` method set unless required by prior cards.

Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md
- docs/stable-api/snapshots/store-head.snapshot

Tests:
- go test -timeout 20s ./store/...
- go test -race -timeout 60s ./store/...
- go run ./internal/checks/extension-api-snapshot -module ./store/... -out docs/stable-api/snapshots/store-head.snapshot

Docs Sync:
- Required for KV performance envelope and stable API evidence.

Done Definition:
- KV has benchmarks documenting the cost model.
- Store API snapshot matches current code.
- Store tests, race tests, vet, and boundary checks pass.

Outcome:
