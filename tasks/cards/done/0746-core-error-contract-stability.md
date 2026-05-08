# Card 0746

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
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
Completed.

Changes:

- Replaced Go map formatting in core operation errors with deterministic sorted `key=value` output.
- Added coverage proving nil-handler route errors still unwrap to `contract.ErrHandlerNil`.
- Updated tests that depended on the previous map-shaped parameter text.

Validation:

- `go test -timeout 20s ./core/...` passed.
- `go test -race -timeout 60s ./core/...` passed.
- `go run ./internal/checks/dependency-rules` passed.
