# 0938 - x/websocket explicit server config

Status: done
Priority: P1
State: done
Primary module: `x/websocket`

## Goal

Remove implicit and ignored server configuration from the stable-facing
`Server` constructor path.

## Scope

- Remove `WS_SECRET` environment reads from `DefaultWebSocketConfig`.
- Change `New` to accept only `WebSocketConfig`.
- Update all direct callers and tests.
- Update docs that mention constructor shape or default secret loading.

## Non-goals

- Changing route behavior.
- Changing auth policy beyond making configuration explicit.
- Promotion to beta or stable.

## Files

- `x/websocket/websocket.go`
- `x/websocket/websocket_test.go`
- direct callers found by symbol search
- websocket docs affected by constructor examples

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`

## Docs Sync

Document that applications read environment variables themselves and pass
secrets explicitly.

## Done Definition

- `DefaultWebSocketConfig` has no environment dependency.
- `New` has no ignored parameters.
- Old constructor call sites are migrated.
- Validation passes.

## Outcome

- Removed `WS_SECRET` reads from `DefaultWebSocketConfig`.
- Changed `New` to accept only `WebSocketConfig`.
- Migrated direct callers and docs to pass secrets explicitly from application code.
- Added test coverage that default config leaves `Secret` empty.

Completed validations:

- `go test -timeout 20s ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`
