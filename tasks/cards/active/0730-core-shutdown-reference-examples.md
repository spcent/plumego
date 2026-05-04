# 0730 - core Shutdown Reference Examples

State: active
Priority: P2
Primary Module: core

## Goal

Make reference applications consistently handle or explicitly discard core
shutdown errors.

## Scope

- Update reference app shutdown defers that currently ignore the returned error.
- Keep changes local to application shutdown examples.
- Prefer explicit `_ =` discards where the surrounding example has no logger or
  recovery path.

## Non-goals

- Do not change reference app startup, routing, or dependency wiring.
- Do not change core runtime behavior.
- Do not add logging dependencies.

## Files

- `reference/standard-service/internal/app/app.go`
- `reference/production-service/internal/app/app.go`
- `reference/with-gateway/internal/app/app.go`
- `reference/with-messaging/internal/app/app.go`
- `reference/with-webhook/internal/app/app.go`

## Tests

- `go test -timeout 20s ./reference/...`
- `go run ./internal/checks/reference-layout`

## Docs Sync

Not required.

## Done Definition

- The listed reference apps do not silently ignore core `Shutdown` return values.
- Reference tests and layout checks pass.

