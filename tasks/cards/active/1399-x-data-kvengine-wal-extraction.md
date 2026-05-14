# Card 1399

Milestone: v1-cleanup-phase-4
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/kvengine
Owned Files:
- x/data/kvengine/kv.go
- x/data/kvengine/wal.go
- x/data/kvengine/kv_test.go
- x/data/kvengine/serializer_test.go
Depends On:
- 1398

Goal:
- Extract `x/data/kvengine` WAL responsibilities from the large store implementation file without changing persistence semantics.

Scope:
- Move WAL entry encoding/decoding, WAL init, append, replay, flush, and close-related helpers into `wal.go`.
- Preserve WAL sync mode behavior, CRC validation, auto-detection behavior, and close-time flush guarantees.
- Keep public store APIs and errors unchanged.

Non-goals:
- Do not change snapshot format.
- Do not change serializer implementations.
- Do not introduce compaction, rotation, or new WAL features.
- Do not change concurrency semantics around write operations.

Files:
- x/data/kvengine/kv.go
- x/data/kvengine/wal.go
- x/data/kvengine/kv_test.go
- x/data/kvengine/serializer_test.go

Tests:
- go test -timeout 30s ./x/data/kvengine
- go vet ./x/data/kvengine
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless package-level WAL documentation changes.

Done Definition:
- WAL-specific logic lives in `wal.go`.
- Existing persistence, recovery, serializer, and close-path tests pass.
- No public API or file-format change is introduced.

Outcome:

