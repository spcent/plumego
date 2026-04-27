# Card 0041

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:
- 0040-store-cache-config-and-doc-contract

Goal:
Keep `store/kv` memory and disk state consistent when persistence fails.

Scope:
- Roll back `Set` mutations if state persistence fails.
- Roll back `Delete` mutations if state persistence fails.
- Remove best-effort persistence side effects from read-only `Exists` and `Keys`.
- Add focused tests for failed `Set`, failed `Delete`, and read-only expired-key checks.

Non-goals:
- Do not add WAL, snapshots, retry loops, or durable-engine tuning.
- Do not change the `KVStore` public method set.
- Do not add tenant-aware behavior.

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
- Failed persistence leaves in-memory key/value state as it was before the mutation.
- `Exists` and `Keys` no longer hide persistence errors behind read-only calls.
- Existing TTL, stats, and reopen behavior still pass.

Outcome:
- Added map cloning and rollback for `Set` and `Delete` when persistence fails.
- Stopped `Exists` and `Keys` from mutating expired entries or ignoring persistence failures.
- Added targeted tests for failed write rollback, failed delete rollback, and read-only expired checks.

Validation:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv
