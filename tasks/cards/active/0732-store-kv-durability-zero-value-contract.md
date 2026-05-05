# Card 0732

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md
Depends On:
- 0731

Goal:
Clarify KV durability guarantees and zero-value object behavior for stable store primitives.

Scope:
- Sync the KV parent directory after atomic state replacement when supported.
- Document constructor-only behavior for `MemoryCache` and `KVStore` zero values.
- Add tests ensuring zero-value objects fail closed instead of panicking where practical.

Non-goals:
- Do not add WAL, snapshots, compression, or durable-engine tuning.
- Do not make zero-value stores fully usable.
- Do not introduce non-stdlib dependencies.

Files:
- store/kv/kv.go
- store/kv/kv_test.go
- store/cache/cache.go
- store/cache/cache_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/cache ./store/kv
- go vet ./store/cache ./store/kv

Docs Sync:
- Update store README with atomic-replace durability and constructor-only object guidance.

Done Definition:
- KV persistence syncs the parent directory when possible.
- Zero-value cache/KV objects do not panic in documented common methods.
- Documentation names stable durability limits.

Outcome:

Validation:
