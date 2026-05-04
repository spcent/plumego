# 0727 - x/websocket constructor setter validation

Status: active
Priority: P0
Primary module: `x/websocket`

## Problem

`NewHubWithConfig`, `NewHub`, `NewConn`, and runtime setters still accept invalid
inputs through silent defaulting or panic-prone paths. Stable APIs must expose
configuration errors visibly.

## Scope

- Add error-returning constructors for hub and connection setup.
- Validate nil connections, negative queue sizes, invalid send behavior, and
  invalid hub worker/queue sizes.
- Make read limit, ping period, and pong wait setters reject invalid values.
- Migrate internal and test callers to the error-returning constructors where
  behavior matters.
- Keep or remove legacy convenience wrappers consistently with the manifest.

## Out of Scope

- Changing route registration defaults.
- Protocol-level frame validation.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

