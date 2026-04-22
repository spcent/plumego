# Card 2105: Harden WebSocket Handshake Error Contract

Priority: P1
State: active
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/websocket
Owned Files:
- x/websocket/server.go
- x/websocket/websocket.go
- x/websocket/errors.go
- x/websocket/server_config_test.go
- x/websocket/websocket_extended_test.go
Depends On:

## Goal

`x/websocket` has strong connection lifecycle coverage, but the handshake
transport error surface is still uneven:

- Several `contract.NewErrorBuilder()` calls use only `Type(...).Message(...)`
  without stable codes.
- `server.go` has a literal `"JOIN_DENIED"` code.
- Some config or join errors are returned as raw `err.Error()` messages.
- `websocket.go` has similar method/body error construction that should align
  with `server.go`.

The result is a non-uniform API for clients trying to distinguish method,
upgrade, auth, origin, room, and server capability failures.

## Scope

- Add package-local helpers/constants for handshake error responses.
- Use stable codes for method-not-allowed, bad upgrade request, missing or
  invalid key, forbidden origin, invalid token, room join denial, hijack
  unsupported, and internal read/setup failures.
- Keep fail-closed behavior for origin, room password, and token checks.
- Keep successful WebSocket upgrade behavior unchanged.
- Add table tests for the error status/code/message matrix.

## Non-goals

- Do not change frame parsing, hub lifecycle, or broadcast semantics.
- Do not weaken origin validation or token handling.
- Do not add dependencies.
- Do not change `ServeWSWithAuth` allow-all origin compatibility.

## Files

- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/errors.go`
- `x/websocket/server_config_test.go`
- `x/websocket/websocket_extended_test.go`

## Tests

```bash
go test -race -timeout 60s ./x/websocket/...
go test -timeout 20s ./x/websocket/...
go vet ./x/websocket/...
```

## Docs Sync

Update `docs/modules/x-websocket/README.md` only if any public error code or
documented handshake behavior changes.

## Done Definition

- Handshake errors use stable codes rather than ad hoc literals or type-only
  responses.
- Raw internal errors are not exposed as client-facing messages unless they are
  explicitly safe validation errors.
- Existing successful upgrade tests continue to pass.
- New negative-path tests assert the expected status and code for each
  handshake failure class.
- The listed validation commands pass.

## Outcome

