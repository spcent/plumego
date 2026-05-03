# Card 0716

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: core
Owned Files: core/app_helpers.go, core/app_test.go, core/routing_test.go
Depends On: 0715-core-mutation-preparation-sync

Goal:
Stabilize core operation error formatting without changing public entrypoints.

Scope:
Replace map-string error parameter output with deterministic key ordering.
Keep wrapped causes available through Go error unwrapping.
Update focused tests that assert error messages.

Non-goals:
Do not introduce an exported error type.
Do not change `contract` error construction.
Do not change router error strings.

Files:
core/app_helpers.go
core/app_test.go
core/routing_test.go

Tests:
go test -timeout 20s ./core/...
go test -race -timeout 60s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
None expected; this is internal error formatting cleanup.

Done Definition:
Core operation errors with parameters are deterministic.
Existing callers can still unwrap the underlying cause.
Core tests and dependency check pass.

Outcome:
