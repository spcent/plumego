# Card 0123

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
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

Validation:
