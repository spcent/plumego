# Card 0718

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: core
Owned Files: core/app.go, core/config.go, core/lifecycle.go, core/app_test.go, core/lifecycle_test.go, docs/modules/core/README.md
Depends On: 0717-core-any-route-options

Goal:
Align connection-tracking names, comments, tests, and docs with actual core behavior.

Scope:
Clarify that core tracks active HTTP connections during graceful shutdown.
Remove stale test names and unused lifecycle scaffolding from core tests.
Keep runtime behavior unchanged.

Non-goals:
Do not add websocket-specific behavior to core.
Do not change shutdown semantics.
Do not add public APIs.

Files:
core/app.go
core/config.go
core/lifecycle.go
core/app_test.go
core/lifecycle_test.go
docs/modules/core/README.md

Tests:
go test -timeout 20s ./core/...
go test -race -timeout 60s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
Update `docs/modules/core/README.md` if comments clarify public behavior.

Done Definition:
Core comments and tests no longer imply websocket or panic behavior that is not owned by core.
Runtime behavior remains unchanged.
Core tests and dependency check pass.

Outcome:
Completed.

Changes:

- Updated core comments and docs to describe active HTTP connection tracking instead of websocket or generic in-flight semantics.
- Renamed the stale panic-oriented middleware test to match the current error-return behavior.
- Removed unused lifecycle test scaffolding from `core/app_test.go`.

Validation:

- `go test -timeout 20s ./core/...` passed.
- `go test -race -timeout 60s ./core/...` passed.
- `go run ./internal/checks/dependency-rules` passed.
