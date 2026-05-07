# Card 0738

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md

Goal:
Make stable KV filesystem writes explicit by requiring callers to choose a data directory.

Scope:
- Stop defaulting an empty `Options.DataDir` to the relative `data` directory in `NewKVStore`.
- Return a clear configuration error when `DataDir` is empty.
- Add tests proving `Options{}` does not create relative directories implicitly.
- Sync store docs and examples with explicit `DataDir` construction.

Non-goals:
- Do not add WAL, snapshots, serializer selection, compression, or sharding topology.
- Do not introduce provider-specific durable engine behavior into stable store.
- Do not add non-stdlib dependencies.

Files:
- store/kv/kv.go
- store/kv/kv_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/kv
- go vet ./store/kv
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update store module docs and examples to require explicit `DataDir`.

Done Definition:
- Empty `DataDir` fails before filesystem creation.
- Tests prove no implicit relative `data` directory creation for `Options{}`.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Changed `NewKVStore` to require an explicit `Options.DataDir` before filesystem creation.
- Kept `MaxEntries` and `MaxMemoryMB` zero-value defaults.
- Added tests proving empty or whitespace-only `DataDir` fails and `Options{}` does not create a relative `data` directory.
- Updated store module docs with the explicit `DataDir` contract.

Validation:
- `gofmt -w store/kv/kv.go store/kv/kv_test.go`
- `go test -timeout 20s ./store/kv`
- `go vet ./store/kv`
- `go run ./internal/checks/dependency-rules`
