# Card 0318

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:
- 0317-store-kv-key-and-options-contract

Goal:
Preserve in-memory KV state when expired-key cleanup fails to persist during `Get`.

Scope:
- Snapshot KV data before deleting an expired key in `Get`.
- Restore the snapshot if persistence fails.
- Add focused coverage for failed expired-key cleanup rollback.

Non-goals:
- Do not change the on-disk JSON format.
- Do not add durable-engine features such as WAL or snapshots.
- Do not change eviction policy.

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
- `Get` no longer loses in-memory entries when expired cleanup cannot persist.
- Successful expired-key cleanup still removes the key and returns `ErrKeyExpired`.
- KV race and normal tests pass.

Outcome:
- Restored KV data when expired-key cleanup in `Get` fails to persist.
- Added regression coverage for the rollback path.

Validation:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv
