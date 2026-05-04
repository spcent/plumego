# Card 0722

Milestone: M-002
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: active
Primary Module: router
Owned Files: router/router.go, router/router_test.go, core/routing.go, core/routing_test.go, docs/stable-api/snapshots/router-head.snapshot, docs/stable-api/snapshots/core-head.snapshot
Depends On: 0721-core-mutation-helper-and-tests

Goal:
Replace duplicated app/router `"ANY"` method sentinel strings with one router-owned exported constant.

Scope:
Add a router-owned exported method constant for catch-all method registration.
Update core to use the router-owned sentinel.
Update tests and stable API snapshots affected by the exported router symbol.

Non-goals:
Do not change ANY matching behavior.
Do not introduce additional route helper aliases.
Do not make core own route matching internals.

Files:
router/router.go
router/router_test.go
core/routing.go
core/routing_test.go
docs/stable-api/snapshots/router-head.snapshot
docs/stable-api/snapshots/core-head.snapshot

Tests:
go test -timeout 20s ./core/... ./router/...
go test -race -timeout 60s ./core/... ./router/...
go run ./internal/checks/dependency-rules

Docs Sync:
Regenerate the core and router stable API snapshots.

Done Definition:
Core no longer duplicates the ANY method sentinel string.
Router exposes the canonical catch-all method sentinel.
Core/router tests and dependency check pass.

Outcome:
