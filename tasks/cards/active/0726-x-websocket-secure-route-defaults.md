# 0726 - x/websocket secure route defaults

Status: active
Priority: P0
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

