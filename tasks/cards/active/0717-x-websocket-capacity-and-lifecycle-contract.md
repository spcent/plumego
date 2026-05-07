# 0717 - x/websocket capacity and lifecycle contract

Status: active
Priority: P1
Primary module: `x/websocket`

## Problem

`MaxConnections` and `GetTotalCount` describe total connections but actually
count room registrations. `Shutdown(ctx)` assumes a non-nil context and
context-cancel semantics do not clearly state whether rooms are cleared.

## Scope

- Rename or document room-registration capacity so the metric and config names
  match behavior.
- Add nil-context handling for shutdown.
- Clarify and test shutdown behavior on context cancellation.
- Update metric JSON/doc fields only if the public contract is intentionally
  changed in this card.

## Out of Scope

- Unique connection caps.
- Worker queue redesign.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

