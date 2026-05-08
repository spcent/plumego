# 1189 - x/websocket registration validation

Status: done
Priority: P0
Primary module: `x/websocket`

## Goal

Make websocket route setup fail before starting runtime goroutines or mutating
the route registrar.

## Scope

- Validate `WSRoutePath`, broadcast path, and broadcast secret requirements in
  `New`.
- Keep `RegisterRoutes` as a final defensive check, but make it validate all
  static config before the first `AddRoute`.
- Add tests that failed setup does not partially register routes.

## Non-goals

- Adding rollback support to router registration.
- Changing default route paths.
- Changing auth semantics.

## Files

- `x/websocket/websocket.go`
- `x/websocket/websocket_test.go`
- `docs/modules/x-websocket/README.md`
- `tasks/cards/active/README.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document that route config errors are constructor-time failures.

## Done Definition

- Invalid route config fails before a hub is created.
- `RegisterRoutes` performs no partial registration on static config errors.
- Validation passes.

## Outcome

- Added constructor-time route configuration validation for websocket path and
  enabled broadcast route requirements before hub creation.
- Made `RegisterRoutes` repeat static route validation before any `AddRoute`
  call.
- Added tests for constructor rejection and no partial registration on static
  config errors.
- Documented constructor-time route validation in the module README.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
