# 1264 - Core Error Contract Test Convergence

State: done
Priority: P2
Primary module: core

## Goal

Reduce brittle complete-string assertions while keeping the stable core error contract covered.

## Scope

- Introduce unexported operation constants for core error operation names.
- Reuse those constants in implementation and tests.
- Update focused tests to assert operation prefix, key message fragments, and wrapped causes instead of full string equality where practical.

## Non-goals

- Do not export lifecycle sentinel errors.
- Do not change error messages intentionally.
- Do not change HTTP error response shape.

## Files

- `core/app_helpers.go`
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/middleware.go`
- `core/routing.go`
- `core/*_test.go`
- `tasks/cards/done/1264-core-error-contract-test-convergence.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

## Docs Sync

No docs update required; the documented error contract remains unchanged.

## Done Definition

- Core error operation names are centralized internally.
- Tests no longer over-freeze complete error strings where operation/message/cause assertions are sufficient.

## Validation

- `go test -timeout 20s ./core/...`
- `go vet ./core/...`
