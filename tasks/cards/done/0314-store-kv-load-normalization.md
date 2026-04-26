# Card 0314

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go
Depends On:
- 0313-store-kv-persist-rollback

Goal:
Normalize `store/kv` loaded state and sentinel error presentation.

Scope:
- Recompute entry size when loading persisted state so stale or zero size fields cannot corrupt stats or capacity decisions.
- Apply existing capacity eviction after load, not only after writes.
- Prefix KV sentinel error messages with `kv:` for consistency with other stable store subpackages.
- Add focused tests for load-time size repair, load-time max-entry eviction, and sentinel strings.

Non-goals:
- Do not change sentinel identities.
- Do not add new exported errors.
- Do not add durable-engine features.

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
- Loaded entries have correct memory sizes even from older state files.
- Stores opened with tighter capacity converge immediately.
- Existing callers matching sentinels by identity or `errors.Is` continue to work.

Outcome:
- Namespaced stable KV sentinel error messages with `kv:`.
- Recomputed persisted entry sizes at load time.
- Applied existing capacity eviction during store open.
- Added tests for sentinel strings, load-time size repair, and load-time max-entry convergence.

Validation:
- go test -timeout 20s ./store/kv
- go test -race -timeout 60s ./store/kv
- go vet ./store/kv
