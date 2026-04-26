# Card 0308

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:
- 0307-store-cache-context-expiry-contract

Goal:
Fix `store/kv` persistence and capacity edge cases while keeping the stable embedded KV surface small.

Scope:
- Reject a single value that cannot fit within the configured max memory instead of accepting the write and evicting it immediately.
- Make state-file persistence use a unique temp file, cleanup on failure, and sync written state before rename.
- Normalize `DataDir` validation so whitespace-only paths do not create surprising directories.
- Add focused tests for oversized values and persistence temp-file cleanup.

Non-goals:
- Do not add WAL, snapshots, serializers, compression, sharding, or durable-engine tuning.
- Do not add tenant-aware storage policy.
- Do not change the concrete `KVStore` API.

Files:
- store/kv/kv.go
- store/kv/kv_test.go

Tests:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv

Docs Sync:
- Not required unless a new public symbol is introduced.

Done Definition:
- Oversized writes return an error and do not silently disappear.
- State persistence leaves no stale deterministic temp path after failure.
- Existing TTL, key scan, stats, and reopen tests still pass.

Outcome:
- Trimmed whitespace-only `DataDir` values through the defaulting path.
- Rejected values larger than the configured memory ceiling before mutating store state.
- Replaced deterministic temp-state writes with unique temp files, fsync, rename, and cleanup-on-failure.
- Added tests for defaulting, oversized writes, and temp-file cleanup on persist failure.

Validation:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv
