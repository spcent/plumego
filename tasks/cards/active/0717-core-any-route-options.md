# Card 0717

Milestone: M-002
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: active
Primary Module: core
Owned Files: core/routing.go, core/routing_test.go, docs/modules/core/README.md, docs/stable-api/snapshots/core-head.snapshot
Depends On: 0716-core-error-contract-stability

Goal:
Let `App.Any` use the same route metadata path as other route registration entrypoints.

Scope:
Extend `App.Any` to accept variadic `router.RouteOption` values while keeping existing calls source-compatible.
Add coverage for named ANY routes through `App.Any`.
Update stable API snapshot and docs if the signature change is reflected there.

Non-goals:
Do not export a new method constant unless required.
Do not change router `ANY` matching semantics.
Do not add named-route helper aliases.

Files:
core/routing.go
core/routing_test.go
docs/modules/core/README.md
docs/stable-api/snapshots/core-head.snapshot

Tests:
go test -timeout 20s ./core/...
go test -race -timeout 60s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
Update core docs and stable API snapshot when the public method signature changes.

Done Definition:
Existing `App.Any(path, handler)` calls still compile.
Named ANY routes can be registered without manually spelling `"ANY"`.
Core tests and dependency check pass.

Outcome:
