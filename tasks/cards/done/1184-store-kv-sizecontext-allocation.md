# Card 1184

Milestone:
Recipe: specs/change-recipes/refactor.yaml
Priority: P3
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go

Goal:
Avoid unnecessary key allocation and sorting in `KVStore.SizeContext`.

Scope:
- Make `SizeContext` count non-expired keys under the store read lock without calling `KeysContext`.
- Preserve error behavior for nil, closed, invalid context, and normal stores.
- Add or adjust tests to cover size behavior without relying on sorted key allocation.

Non-goals:
- Do not change `KeysContext` ordering.
- Do not change stats semantics.
- Do not change persistence behavior.

Files:
- store/kv/kv.go
- store/kv/kv_test.go

Tests:
- go test -timeout 20s ./store/kv
- go vet ./store/kv
- go run ./internal/checks/dependency-rules

Docs Sync:
- Not required for behavior-preserving optimization.

Done Definition:
- `SizeContext` no longer allocates or sorts a key slice.
- Existing public behavior remains unchanged.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- `KVStore.SizeContext` now counts non-expired keys directly under the read lock.
- `KeysContext` keeps sorted key enumeration semantics, while size no longer depends on key allocation or sorting.
- Tests cover live size consistency and expired-key exclusion.

Validation:
- `go test -timeout 20s ./store/kv`
- `go vet ./store/kv`
- `go run ./internal/checks/dependency-rules`
