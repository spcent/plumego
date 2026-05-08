# Card 0778

Milestone: M-003
Recipe: specs/change-recipes/implementation.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/server.go`
- `x/websocket/errors.go`
- `x/websocket/server_config_test.go`
- `x/websocket/ws_test.go`
- `x/websocket/module.yaml`

Goal:
- Let application message handlers choose client-facing close semantics without weakening validation.

Scope:
- Add a typed handler close error or equivalent explicit mechanism.
- Keep validation/read/protocol errors mapped to existing close codes.
- Update comments so `ServeWSWithConfig` is documented as bounded whole-message handling.
- Keep `ServeRoomFanoutWS` clearly documented as the fanout helper.

Non-goals:
- Do not implement true streaming callbacks in this card.
- Do not add application-level retry or acknowledgement semantics.

Files:
- `x/websocket/server.go`
- `x/websocket/errors.go`
- `x/websocket/server_config_test.go`
- `x/websocket/ws_test.go`
- `x/websocket/module.yaml`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for new public error type and handler semantics.

Done Definition:
- Handler errors can map to policy/client/server close codes intentionally.
- Tests cover the default and typed handler-error paths.

Outcome:
- Added `CloseError` and `NewCloseError` for message handlers that need to
  choose a close code and reason.
- Mapped typed handler errors to their requested close frame while preserving
  the default 1011 server-error close for ordinary errors.
- Added integration tests for typed and default handler close behavior.
- Verified with `go test -timeout 20s ./x/websocket/...`, `go vet
  ./x/websocket/...`, and `go run ./internal/checks/module-manifests`.
