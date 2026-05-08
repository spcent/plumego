# Card 0715

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: core
Owned Files: core/app_helpers.go, core/routing.go, core/middleware.go, core/app_test.go, core/routing_test.go
Depends On: 0714-core-prepare-failure-atomicity

Goal:
Make route and middleware mutation atomic with app preparation.

Scope:
Keep mutable-state checks and the corresponding route or middleware mutation under one synchronization boundary.
Add regression coverage for concurrent preparation versus registration where practical.
Preserve the existing public API and frozen-state errors.

Non-goals:
Do not make middleware implementations themselves concurrency-safe.
Do not add new lifecycle states.
Do not change router matching internals.

Files:
core/app_helpers.go
core/routing.go
core/middleware.go
core/app_test.go
core/routing_test.go

Tests:
go test -race -timeout 60s ./core/...
go test -timeout 20s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
None expected unless the mutation/preparation contract changes.

Done Definition:
No app preparation can freeze between a successful mutable-state check and the corresponding mutation.
Concurrent registration/preparation coverage passes under `-race`.
Core tests and dependency check pass.

Outcome:
Completed.

Changes:

- Kept `Use` mutable-state checks and middleware chain mutation under one app lock.
- Kept route mutable-state checks and router mutation under one app lock.
- Added race-focused regression tests for concurrent `Use`/`Prepare` and route registration/`Prepare`.

Validation:

- `go test -timeout 20s ./core/...` passed.
- `go test -race -timeout 60s ./core/...` passed.
- `go run ./internal/checks/dependency-rules` passed.
