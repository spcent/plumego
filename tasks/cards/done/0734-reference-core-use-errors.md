# 0734 - reference Core Use Errors

State: done
Priority: P1
Primary Module: core

## Goal

Make canonical and extension reference apps honor the `core.Use` error contract
instead of silently ignoring middleware registration failures.

## Scope

- Update reference app constructors that call `Use` without checking errors.
- Return wrapped constructor errors on middleware registration failure.
- Keep middleware order unchanged.

## Non-goals

- Do not change route wiring.
- Do not alter middleware behavior.
- Do not add new logging or dependencies.

## Files

- `reference/standard-service/internal/app/app.go`
- `reference/with-gateway/internal/app/app.go`
- `reference/with-messaging/internal/app/app.go`
- `reference/with-webhook/internal/app/app.go`

## Tests

- `go test -timeout 20s ./reference/...`
- `go run ./internal/checks/reference-layout`

## Docs Sync

Not required.

## Done Definition

- Listed reference constructors return errors from `core.Use`.
- Reference tests and layout checks pass.

## Outcome

- Updated listed reference constructors to return wrapped middleware
  registration errors from `core.Use`.
- Preserved middleware registration order.
- Verified with `go test -timeout 20s ./reference/...` and
  `go run ./internal/checks/reference-layout`.
