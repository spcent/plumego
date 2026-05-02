# Card 0765

Milestone: M-003
Recipe: specs/change-recipes/security-hardening.yaml
Priority: P0
State: todo
Primary Module: x/websocket
Owned Files:
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0764

Goal:
- Remove secret-bearing room credentials from URL query handling.

Problem:
`room_password` is read from the URL query string, which can leak through access logs, browser history, and referrers. Stable defaults should avoid URL-carried secrets.

Scope:
- Stop reading room passwords from `?room_password=`.
- Move built-in room credentials to a header-based policy or require custom room authorization to read credentials from the request.
- Keep `AllowQueryToken` explicitly opt-in for token transport if retained.
- Add negative tests proving URL room passwords are rejected or ignored.

Non-goals:
- Do not remove the room concept.
- Do not add request-body parsing to the WebSocket handshake.
- Do not weaken existing token defaults.

Files:
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for credential transport and security caveats.

Done Definition:
- Default websocket setup does not read room secrets from URL query parameters.
- Header/custom room authorization paths are documented.
- Tests cover URL secret rejection and allowed non-secret room selection.

Outcome:
-
