# Card 0721

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: core
Owned Files: core/app_helpers.go, core/app_test.go, core/routing_test.go, core/introspection_test.go
Depends On: 0720-core-shutdown-contract

Goal:
Remove stale mutation helper paths and reduce tests that assert behavior by directly mutating internal state.

Scope:
Delete unused mutable-state helper code.
Replace direct `preparationState` mutation tests with public behavior tests where possible.
Keep current public API and runtime behavior unchanged.

Non-goals:
Do not add new lifecycle states.
Do not change `Prepare`, `ServeHTTP`, route matching, or middleware ordering behavior.
Do not broaden test coverage outside `core`.

Files:
core/app_helpers.go
core/app_test.go
core/routing_test.go
core/introspection_test.go

Tests:
go test -timeout 20s ./core/...
go test -race -timeout 60s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
None expected.

Done Definition:
Stale mutation helper code is removed.
Core tests verify prepared-state behavior through public entrypoints rather than manually setting `preparationState`.
Core tests and dependency check pass.

Outcome:
