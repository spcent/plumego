# Card 1412

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/ipc
Owned Files:
- x/ipc/ipc.go
- x/ipc/heartbeat.go
- x/ipc/pool.go
- x/ipc/stream.go
- x/ipc/ipc_test.go
Depends On:
- 1411

Goal:
- Continue reducing `x/ipc` edit radius by splitting heartbeat, pool, and stream wrappers.

Scope:
- Move heartbeat client and config helpers into `heartbeat.go`.
- Move connection pool types and helpers into `pool.go`.
- Move stream client and stream config helpers into `stream.go`.
- Preserve remote address, timeout, pool cleanup, and stream chunk behavior.

Non-goals:
- Do not change framing, client, server, or reconnect files.
- Do not change platform-specific Unix or Windows transports.
- Do not change compatibility helpers.

Files:
- x/ipc/ipc.go
- x/ipc/heartbeat.go
- x/ipc/pool.go
- x/ipc/stream.go
- x/ipc/ipc_test.go

Tests:
- go test -timeout 20s ./x/ipc
- go vet ./x/ipc
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless package examples move.

Done Definition:
- `ipc.go` no longer owns heartbeat, pool, and stream wrapper implementations.
- Existing IPC tests pass.
- No wire protocol or platform behavior change is introduced.

Outcome:

