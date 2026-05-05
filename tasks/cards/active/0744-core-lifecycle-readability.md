# 0744 - core Lifecycle Readability

State: active
Priority: P2
Primary Module: core

## Goal

Clean up minor core lifecycle readability issues without changing behavior.

## Scope

- Rename the state-mutating `serverAlreadyPrepared` helper to reflect its side
  effect.
- Keep prepare idempotency behavior unchanged.
- Normalize import grouping in lifecycle tests.

## Non-goals

- Do not change public API.
- Do not alter shutdown or prepare semantics.
- Do not refactor unrelated helper code.

## Files

- `core/http_handler.go`
- `core/lifecycle_test.go`

## Tests

- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

## Docs Sync

Not required.

## Done Definition

- Helper naming no longer implies a pure query when state is mutated.
- Core lifecycle tests keep standard import grouping.
- Core tests and vet pass.

