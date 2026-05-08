# 0974 - x/websocket lifecycle close contract

Status: done
Priority: P0
Primary module: `x/websocket`

## Goal

Make `Stop`, `Shutdown`, room cleanup, and close-frame behavior internally
consistent and documented.

## Scope

- Make `Shutdown` use best-effort close frames before closing connections.
- Clear hub room registrations during shutdown.
- Clarify `Stop` as worker shutdown only, or rename behavior in docs/comments.
- Add tests for room cleanup, close-frame emission, and post-shutdown metrics.

## Non-goals

- Full WebSocket closing handshake wait.
- New dependencies.

## Files

- `x/websocket/hub.go`
- `x/websocket/conn.go`
- `x/websocket/hub_lifecycle_test.go`
- `x/websocket/writer_pump_test.go`
- websocket docs/comments as needed

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

## Docs Sync

Document exact difference between `Stop`, `Shutdown`, `Close`, and
`WriteClose`.

## Done Definition

- Shutdown does not leave stale rooms/registrations.
- Close-frame behavior is tested and documented.
- Validation passes.

## Outcome

- `Shutdown` now stops workers, clears hub room registrations, resets the
  room-registration metric, and best-effort sends `CloseGoingAway` frames before
  closing registered connections.
- `Stop` documentation now explicitly describes worker shutdown only and points
  callers to `Shutdown` for room cleanup and close-frame emission.
- Added lifecycle coverage for nil contexts, canceled shutdown cleanup,
  post-shutdown room metrics, and close-frame emission.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`
