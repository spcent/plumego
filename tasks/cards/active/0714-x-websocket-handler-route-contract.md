# 0714 - x/websocket handler route contract

Status: active
Priority: P1
Primary module: `x/websocket`

## Problem

`ServeWSWithConfig` currently mixes transport setup with chat-style room fanout:
client messages are automatically broadcast back to the room. `Server.RegisterRoutes`
also hides this behavior by routing through the compatibility helper. This makes
the stable transport API unclear because applications cannot distinguish a low
level websocket handler from a product-specific fanout helper.

## Scope

- Add an explicit message handler contract for `ServeWSWithConfig`.
- Add a named room-fanout helper for the existing broadcast-back behavior.
- Make `Server.RegisterRoutes` wire either the configured message handler or the
  room-fanout helper explicitly.
- Keep route registration visible and structured-error based.
- Update focused tests and `docs/modules/x-websocket/README.md`.

## Out of Scope

- Authentication model changes.
- Hub capacity or shutdown changes.
- Public promotion from experimental.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

