# Card 1398

Milestone: v1-cleanup-phase-4
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/ipc
Owned Files:
- x/ipc/ipc.go
- x/ipc/client.go
- x/ipc/server.go
- x/ipc/reconnect.go
- x/ipc/ipc_test.go
Depends On:
- 1397

Goal:
- Split `x/ipc` client, server, and reconnect responsibilities into separate files without changing behavior.

Scope:
- Move client implementation and client interface wiring into `client.go`.
- Move server construction and accept/close paths into `server.go`.
- Move reconnect configuration and reconnecting client behavior into `reconnect.go`.
- Keep exported symbol names, comments, and behavior intact.

Non-goals:
- Do not change connection timeout, keepalive, permission, or reconnect defaults.
- Do not modify Unix or Windows transport implementations unless a compile error requires a local import adjustment.
- Do not change errors or wrapping semantics.

Files:
- x/ipc/ipc.go
- x/ipc/client.go
- x/ipc/server.go
- x/ipc/reconnect.go
- x/ipc/ipc_test.go

Tests:
- go test -timeout 20s ./x/ipc
- go vet ./x/ipc
- go run ./internal/checks/dependency-rules

Docs Sync:
- None expected unless package examples or comments are moved incorrectly.

Done Definition:
- `ipc.go` keeps package overview, shared config, and high-level public contracts.
- Client, server, and reconnect implementation each have clear file ownership.
- Existing `x/ipc` tests pass without behavior changes.

Outcome:
- Moved client interfaces and dial helpers into `x/ipc/client.go`.
- Moved server interface and construction into `x/ipc/server.go`.
- Moved reconnect configuration, constructor, reconnect loop, and delegated client methods into `x/ipc/reconnect.go`.
- Left heartbeat, pool, stream, rate-limit, platform-specific transport, and protocol behavior unchanged.
- Validation:
  - `go test -timeout 20s ./x/ipc`
  - `go vet ./x/ipc`
  - `go run ./internal/checks/dependency-rules`
