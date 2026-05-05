# 0751 - x/websocket hub input guards

Status: active
Priority: P0
Primary module: `x/websocket`

## Goal

Make public hub entrypoints fail visibly on invalid inputs before stable API
freeze.

## Scope

- Reject invalid room names in `Hub.TryJoin` and `Hub.CanJoin`.
- Reject nil connections in `Hub.TryJoin`.
- Reject negative `MaxRoomRegistrations`, `MaxRoomConnections`, and
  `MaxConnectionRate` in `NewHubWithConfigE`.
- Add tests for direct Hub API invalid room, nil conn, and negative config
  values.

## Non-goals

- Changing handshake room validation.
- Adding new capacity concepts.
- Promotion or release evidence.

## Files

- `x/websocket/hub.go`
- `x/websocket/hub_lifecycle_test.go`
- `x/websocket/security_test.go`
- `docs/modules/x-websocket/README.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document that direct Hub APIs use the same room-name validation as the
handshake.

## Done Definition

- Invalid direct Hub inputs return explicit errors.
- Negative hub limits fail construction.
- Validation passes.
