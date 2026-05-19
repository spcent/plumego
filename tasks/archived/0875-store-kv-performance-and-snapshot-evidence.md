# Card 0875: Store KV Performance And Snapshot Evidence

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P1
State: done
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
- Added focused `store/kv` benchmarks for overwrite `Set`, existing-key `Get`,
  small-dataset eviction, and existing-key `Delete`.
- Reduced embedded KV defaults from 100000 entries / 200 MiB to 4096 entries /
  32 MiB, keeping the stable root aligned with small local state usage.
- Documented the KV envelope in `docs/modules/store/README.md`, including the
  recommendation to use `x/data/kvengine` or explicit measured limits for
  larger datasets.
- Refreshed `docs/stable-api/snapshots/store-head.snapshot` after the store
  API changes.

Benchmark Evidence:
- `BenchmarkKVStoreSetOverwrite-10`: 15394987 ns/op, 2951 B/op, 26 allocs/op.
- `BenchmarkKVStoreGetExisting-10`: 69.72 ns/op, 8 B/op, 1 alloc/op.
- `BenchmarkKVStoreSetEvictSmallDataset-10`: 12554063 ns/op, 20824 B/op,
  212 allocs/op.
- `BenchmarkKVStoreDeleteExisting-10`: 19563574 ns/op, 1577 B/op, 17 allocs/op.

Validation:
- `go test -timeout 20s ./store/...`
- `go test -race -timeout 60s ./store/...`
- `go vet ./store/...`
- `go test -run '^$' -bench 'BenchmarkKVStore' ./store/kv`
- `go run ./internal/checks/extension-api-snapshot -module ./store/... -out docs/stable-api/snapshots/store-head.snapshot`
- `go run ./internal/checks/extension-api-snapshot -compare docs/stable-api/snapshots/store-head.snapshot docs/stable-api/snapshots/store-head.snapshot`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
