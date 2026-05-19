# 0833 - x/websocket secure route defaults

Status: done
Priority: P0
State: done
Primary module: `x/websocket`

## Problem

`Server.RegisterRoutes` still wires anonymous websocket access, origin allow-all,
and an enabled admin broadcast route from defaults. Stable route wiring must fail
closed unless the application explicitly opts into weaker behavior.

## Scope

- Make the admin broadcast route opt-in by default.
- Add explicit server-level fields for websocket anonymous access, allowed
  origins, room auth, token auth, and broadcast authorization.
- Stop using the JWT signing secret as the broadcast admin token.
- Validate route registration inputs such as nil registrar and empty route path.
- Preserve direct `ServeWSWithConfig` behavior for callers that opt in
  explicitly.

## Out of Scope

- RFC6455 frame validation.
- Constructor API redesign outside route registration.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/module-manifests`

## Outcome

- Made `DefaultWebSocketConfig` keep the admin broadcast route disabled by
  default.
- Added explicit route-wiring fields for room auth, token auth, anonymous
  access, query-token support, allowed origins, and a dedicated broadcast
  secret.
- Stopped using the JWT signing secret as the admin broadcast bearer token.
- Added visible route-registration errors for nil registrars and empty paths.
- Updated focused websocket tests and the `reference/with-websocket` demo
  wiring for explicit anonymous/origin policy.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go run ./internal/checks/module-manifests`
  - `go test -timeout 20s ./reference/with-websocket/...`
