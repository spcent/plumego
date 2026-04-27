# Card 0045

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:
- 0044-store-cache-ttl-key-contract

Goal:
Make `store/kv` reject invalid empty keys consistently and namespace option validation errors.

Scope:
- Add a stable invalid-key sentinel for empty KV keys.
- Reject empty keys in key-specific operations.
- Prefix option validation errors with the `kv:` package namespace.
- Add focused tests for empty-key rejection and option error messages.

Non-goals:
- Do not change storage format.
- Do not add tenant, shard, WAL, serializer, or provider configuration behavior.
- Do not change `Keys`, `Size`, or stats APIs.

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
- Empty keys fail before mutation or lookup-side miss accounting.
- Option validation errors are clearly attributable to `store/kv`.
- Existing persistence and TTL tests still pass.

Outcome:
- Added `ErrInvalidKey` for empty KV keys.
- Rejected empty keys in `Set`, `Get`, `Delete`, and `Exists`.
- Namespaced options validation errors with the `kv:` package prefix.
- Added focused tests for invalid keys and options errors.

Validation:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv
