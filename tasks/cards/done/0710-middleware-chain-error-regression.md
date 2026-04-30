# Card 0710

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/middleware.go, middleware/auth/auth.go, middleware/recovery/recover.go, middleware/timeout/timeout.go, middleware/*_test.go
Depends On: 0706-stable-root-api-inventory

Goal:
Lock down middleware chain ordering, exactly-once next behavior, and transport error short-circuit regressions.

Scope:
Add or tighten focused tests for root `middleware.Chain` behavior and one or two representative transport middleware error paths without changing public APIs.

Non-goals:
Do not add tenant, persistence, or business DTO behavior to middleware.
Do not introduce new middleware lifecycle abstractions.
Do not add dependencies.

Files:
middleware/middleware.go
middleware/auth/auth.go
middleware/recovery/recover.go
middleware/timeout/timeout.go
middleware/*_test.go

Tests:
go test -race -timeout 60s ./middleware/...
go test -timeout 20s ./middleware/...
go run ./internal/checks/dependency-rules

Docs Sync:
None unless behavior changes.

Done Definition:
Middleware order and short-circuit behavior have regression coverage.
No new public API is added.
Middleware tests and dependency boundary check pass.

Outcome:
Completed.

Changes:

- Added a root `middleware.Chain` regression test proving `Snapshot()` returns a copy and cannot mutate the chain.
- No runtime code or public API changed.

Validation:

- `go test -race -timeout 60s ./middleware/...` passed.
- `go test -timeout 20s ./middleware/...` passed.
- `go run ./internal/checks/dependency-rules` passed.

