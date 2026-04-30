# Card 0707

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: core
Owned Files: core/app.go, core/routing.go, core/lifecycle.go, core/http_handler.go, core/*_test.go
Depends On: 0706-stable-root-api-inventory

Goal:
Strengthen `core` lifecycle and route-registration regression coverage without changing stable public APIs.

Scope:
Add focused tests for preparation state, route registration failure propagation, middleware attachment order, and server preparation behavior using existing `core` APIs.

Non-goals:
Do not add public APIs.
Do not change route matching internals owned by `router`.
Do not introduce dependencies.

Files:
core/app.go
core/routing.go
core/lifecycle.go
core/http_handler.go
core/*_test.go

Tests:
go test -race -timeout 60s ./core/...
go test -timeout 20s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
None unless behavior changes. If a behavior change is required, record it in Outcome before touching docs.

Done Definition:
Core lifecycle and route-registration edge cases have positive and negative tests.
No stable public API changes are introduced.
Core package tests and dependency boundary check pass.

Outcome:
Completed.

Changes:

- Added a core regression test proving middleware wrappers execute in registration order around the route handler.
- Added a route-registration regression test proving real `Prepare()` freezes route mutation while existing routes remain available.
- No runtime code or public API changed.

Validation:

- `go test -race -timeout 60s ./core/...` passed.
- `go test -timeout 20s ./core/...` passed.
- `go run ./internal/checks/dependency-rules` passed.

