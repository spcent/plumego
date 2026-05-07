# Card 0123

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:
- 0122

Goal:
Make KV error semantics and startup persistence behavior stable and explicit.

Scope:
- Add error-returning variants for `Exists`, `Keys`, `Size`, and stats where needed while preserving compatibility wrappers.
- Avoid mandatory startup persistence when load only prunes expired in memory.
- Treat invalid persisted keys as load corruption instead of silently dropping them.
- Align key validation with cache by rejecting control characters.

Non-goals:
- Do not add WAL, snapshots, or sharding.
- Do not change the single JSON state-file storage model.
- Do not move durable engines into stable store.

Files:
- store/kv/kv.go
- store/kv/kv_test.go

Tests:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv

Docs Sync:
- Public comments only.

Done Definition:
- Context-aware/read APIs surface close/cancel errors.
- Compatibility wrappers remain available.
- Corrupt persisted keys fail load loudly.
- Opening an existing store does not require an immediate write unless caller mutates state.

Outcome:
- Added `GetStatsContext` so stats callers can receive context cancellation and closed-store errors while keeping `GetStats` as a compatibility wrapper.
- Clarified compatibility wrappers for `Exists`, `Keys`, `Size`, and `GetStats` as error-collapsing conveniences.
- Removed mandatory startup persistence after load-time pruning/eviction; startup normalization is now in-memory until a caller-initiated write occurs.
- Made invalid persisted keys fail startup loudly instead of being silently skipped.
- Aligned KV key validation with cache by rejecting ASCII control characters.
- Updated store module docs and stable API snapshot.

Validation:
- go test -timeout 20s ./store/kv
- go vet ./store/kv
- go test -race -timeout 60s ./store/kv
