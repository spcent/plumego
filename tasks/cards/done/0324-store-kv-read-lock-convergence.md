# Card 0324

Milestone:
Recipe: specs/change-recipes/refactor.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:
- 0323-store-kv-load-key-validation

Goal:
Use read locks for KV read-only inspection paths that no longer mutate expired entries.

Scope:
- Switch `Exists` and `Keys` to `RLock`.
- Keep `Get` write-locked because it may delete expired entries and persist cleanup.
- Add focused race-safe coverage around concurrent read-only inspection.

Non-goals:
- Do not change expiry cleanup semantics.
- Do not change persistence or eviction behavior.
- Do not introduce new synchronization primitives.

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
- Read-only `Exists` and `Keys` use read locks.
- Race tests cover concurrent read-only inspection.
- KV targeted tests and vet pass.

Outcome:
- Switched `Exists` and `Keys` from write locks to read locks.
- Added concurrent read-only inspection coverage for `Exists`, `Keys`, and `Size`.

Validation:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv
