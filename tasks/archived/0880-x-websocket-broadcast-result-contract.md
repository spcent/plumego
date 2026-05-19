# 0880 - x/websocket broadcast result contract

Status: done
Priority: P1
State: done
Primary module: `x/websocket`

## Problem

`BroadcastRoom` and `BroadcastAll` are fire-and-forget even when the hub is
stopped or the job queue drops messages. `RejectOnQueueFull` reads like a hard
rejection but callers cannot observe the result.

## Scope

- Add result-returning broadcast APIs that report sent and dropped counts.
- Preserve simple fire-and-forget wrappers only if the stable surface documents
  them clearly.
- Align `RejectOnQueueFull` naming or behavior with the observable result.
- Update tests for stopped hub, empty room, queue full, and successful fanout.

## Out of Scope

- Changing wire format or message payload schemas.
- Route admin broadcast authorization.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Outcome

- Added `BroadcastResult` plus `TryBroadcastRoom` and `TryBroadcastAll` so callers can observe accepted and dropped send jobs.
- Kept `BroadcastRoom` and `BroadcastAll` as fire-and-forget wrappers over the result-returning APIs.
- Counted dropped dispatch jobs independently of metrics flags and covered stopped, empty, sent, and queue-full paths in tests.

Completed validations:

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`
