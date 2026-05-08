# Card 0733

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/server.go`
- `x/websocket/hub.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0732

Goal:
- Make WebSocket join, handshake, and shutdown lifecycle behavior deterministic and client-visible.

Problem:
The server checks Hub capacity before sending `101 Switching Protocols`, but performs the real `TryJoin` after the upgrade response. A capacity race can accept the WebSocket upgrade and then immediately close the TCP connection without a useful HTTP error or close frame. Hub shutdown also calls `Conn.Close`, while comments imply a close frame is sent.

Scope:
- Remove or close the capacity race between pre-handshake capacity checks and post-handshake join.
- Ensure join failure after upgrade produces a clear WebSocket close frame, or make post-upgrade join failure unreachable.
- Align shutdown comments and implementation: either send close frames before hard close or document hard close explicitly.
- Add tests for capacity-full handshake behavior and shutdown close behavior.

Non-goals:
- Do not redesign Hub room ownership or add business authorization.
- Do not change JWT or origin policy beyond what card 0732 owns.
- Do not add reconnect, backoff, or client session recovery.

Files:
- `x/websocket/server.go`
- `x/websocket/hub.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required if shutdown or join failure behavior changes.

Done Definition:
- Capacity-full connections fail before upgrade, or receive an explicit close frame after upgrade by design.
- Hub shutdown behavior matches docs and tests.
- No goroutine or connection lifecycle regression is introduced under race tests.

Outcome:
- Moved the real `Hub.TryJoin` before the `101 Switching Protocols` response so capacity races fail before WebSocket upgrade.
- Added a raw HTTP error response path for post-hijack, pre-upgrade join denial.
- Clarified `Hub.Shutdown` as a hard-close path and added tests for shutdown closing registered connections.
- Added coverage for post-hijack capacity race denial returning HTTP 503 instead of `101`.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go test -race -timeout 60s ./x/websocket/...`
  - `go vet ./x/websocket/...`
