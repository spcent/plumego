# 0761 - x/websocket security event lifecycle

Status: active
Priority: P1
Primary module: `x/websocket`

## Goal

Make security event handler lifecycle bounded and explicit.

## Scope

- Prevent Stop/Shutdown from trying to drain into a permanently blocking user
  handler.
- Keep producer delivery best-effort and bounded.
- Update docs/tests for event drop and shutdown behavior.

## Non-goals

- Guaranteed delivery of all security events.
- A public event subscription API.
- Changing metrics collection.

## Files

- `x/websocket/hub.go`
- `x/websocket/hub_lifecycle_test.go`
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document that handler delivery is best-effort and may drop events during
shutdown.

## Done Definition

- Security event dispatcher cannot remain blocked in shutdown drain logic.
- Panic recovery remains covered.
- Validation passes.
