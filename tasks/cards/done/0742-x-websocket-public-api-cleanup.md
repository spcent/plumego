# Card 0742

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`
Depends On: 0741

Goal:
- Remove legacy public APIs and leave one clear stable path for constructors, joins, and connection counts.

Problem:
The package currently exposes compatibility paths with weaker semantics: `NewHubWithConfig` defaults invalid config instead of returning errors, `NewConn` returns nil on invalid input, `Join` bypasses capacity checks, and `GetTotalCount` says total connections while returning room registrations.

Scope:
- Delete `NewHubWithConfig`; use `NewHubWithConfigE` or rename it to the canonical error-returning constructor.
- Delete `NewConn`; use `NewConnE` or rename it to the canonical error-returning constructor.
- Delete `Join`; require `TryJoin` or a renamed capacity-enforcing join API.
- Delete `GetTotalCount`; expose unambiguous methods such as `GetActiveConnectionCount` and `GetRoomRegistrationCount`, or require `Metrics()`.
- Update all internal callers, tests, docs, and module manifest exported-symbol inventory.
- Follow `AGENTS.md §7.1` before and after removing each exported symbol.

Non-goals:
- Do not preserve deprecated wrappers.
- Do not change Hub capacity policy except to remove bypass access.
- Do not add cross-module abstractions.

Files:
- `x/websocket/hub.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for exported API deletions and migration notes.

Done Definition:
- `rg -n --glob '*.go' 'NewHubWithConfig|NewConn|Join|GetTotalCount' .` has no stale production call sites for deleted names.
- There is one canonical error-returning constructor path per public type.
- All join paths enforce Hub capacity and closed-state checks.
- Connection-count methods and docs distinguish unique connections from room registrations.

Outcome:
- Deleted `NewHubWithConfig`, `NewConn`, `Join`, and `GetTotalCount`.
- Migrated tests and examples to `NewHubWithConfigE`, `NewConnE`, `TryJoin`, and `GetRoomRegistrationCount`.
- Added `GetActiveConnectionCount` for unique active connection counts.
- Updated module manifest, docs, and website docs to remove the deleted symbols.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go build ./...`
  - `go run ./internal/checks/module-manifests`
