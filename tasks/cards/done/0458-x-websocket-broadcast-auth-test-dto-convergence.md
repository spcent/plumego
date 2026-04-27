# Card 0458: x/websocket Broadcast Auth Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: active
Primary Module: x/websocket
Owned Files:
- `x/websocket/websocket_test.go`
Depends On: none

Goal:
Make the broadcast authorization case-insensitivity test use a typed error
response DTO for its failure diagnostics.

Problem:
`TestBroadcastAuthCaseInsensitive` decodes non-204 responses into
`map[string]any`. The payload is an HTTP error envelope with a known shape, so
the map decode keeps the expected response contract implicit and inconsistent
with nearby DTO-oriented tests.

Scope:
- Replace the debug-only `map[string]any` decode with a small local typed
  response struct.
- Preserve the existing assertion that lowercase `bearer` authorization succeeds
  with HTTP 204.

Non-goals:
- Do not change websocket auth behavior, route registration, or public APIs.
- Do not add dependencies.

Files:
- `x/websocket/websocket_test.go`

Tests:
- `go test -race -timeout 60s ./x/websocket/...`
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
No docs change required; this is a test-only DTO convergence.

Done Definition:
- The broadcast auth case-insensitivity test no longer decodes fixed response
  payloads through `map[string]any`.
- The listed validation commands pass.

Outcome:
