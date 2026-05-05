# 0742 - core Validation Precedence

State: active
Priority: P1
Primary Module: core

## Goal

Unify route and middleware validation precedence so app lifecycle state is
checked before argument validation for initialized app receivers.

## Scope

- Move route handler nil validation after app initialization and mutability
  checks.
- Add regression coverage for zero-value and prepared apps with nil route
  handlers.
- Keep nil app behavior and initialized mutable-app nil handler behavior
  unchanged.

## Non-goals

- Do not change router validation semantics.
- Do not change middleware behavior.
- Do not add exported errors.

## Files

- `core/routing.go`
- `core/app_test.go`
- `core/routing_test.go`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`

## Docs Sync

Not required.

## Done Definition

- Route and middleware entrypoints both prioritize app state over argument
  validation for initialized app receivers.
- Existing mutable-app nil handler wrapping remains intact.

