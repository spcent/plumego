# Card 0743

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/writer.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0742

Goal:
- Guarantee WebSocket shutdown and blocking writes cannot hang indefinitely.

Problem:
Hub workers drain queued jobs on quit and call `WriteMessage`; when a connection uses blocking send without a timeout, shutdown can wait forever on queue space. Separately, `enqueueBlocking` uses a 1ms ticker retry loop, which creates avoidable latency and CPU churn.

Scope:
- Change worker shutdown so queued fanout cannot block indefinitely during `Stop` or `Shutdown`.
- Add or enforce bounded write/enqueue deadlines for shutdown drain behavior.
- Replace the 1ms polling enqueue loop with a direct channel/select model that handles close, timeout, and queue availability.
- Add regression tests for `SendBlock` with no timeout, full queues, closed connections, and Hub shutdown.
- Document hard close versus graceful close-frame behavior accurately.

Non-goals:
- Do not change public broadcast authorization.
- Do not add network dependencies or external goroutine-management libraries.
- Do not implement full graceful drain guarantees beyond the documented bounded behavior.

Files:
- `x/websocket/hub.go`
- `x/websocket/writer.go`
- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for shutdown semantics, write modes, queue-full behavior, and close-frame caveats.

Done Definition:
- Hub stop/shutdown tests prove no indefinite wait when client queues are full.
- Blocking send uses channel/select control flow rather than ticker polling.
- Closed connections and queue timeouts return deterministic errors.
- Race tests pass for the WebSocket package.

Outcome:
- Replaced `enqueueBlocking` ticker polling with direct channel/select enqueue and cancellation.
- Added a hub worker fallback send deadline for `SendBlock` connections without their own timeout.
- Added regression tests for blocked queue release, close while blocked, and hub stop with full send queues.
- Documented bounded stop/write behavior and hard shutdown semantics.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go test -race -timeout 60s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go build ./...`
