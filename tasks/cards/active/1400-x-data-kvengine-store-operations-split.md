# Card 1400

Milestone: v1-cleanup-phase-4
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/operations.go
- x/data/kvengine/lru.go
- x/data/kvengine/kv_test.go
Depends On:
- 1399

Goal:
- Split `x/data/kvengine` store operations and LRU helpers into focused files without changing behavior.

Scope:
- Move Set/Get/Delete/Exists-style store operations into `operations.go`.
- Move private LRU chain and eviction helpers into `lru.go`.
- Preserve locking order, metrics observations, TTL behavior, capacity checks, and error values.
- Keep constructor and option/default handling in `kv.go`.

Non-goals:
- Do not change public method names or return contracts.
- Do not change TTL expiry behavior.
- Do not change memory accounting semantics.
- Do not add new storage features.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/operations.go
- x/data/kvengine/lru.go
- x/data/kvengine/kv_test.go

Tests:
- go test -timeout 30s ./x/data/kvengine
- go vet ./x/data/kvengine
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless comments moved during the split require package doc cleanup.

Done Definition:
- Store operation code is separated from construction and WAL code.
- LRU helpers have clear private ownership.
- Existing store behavior and tests remain unchanged.

Outcome:

