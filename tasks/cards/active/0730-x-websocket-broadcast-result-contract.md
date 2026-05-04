# 0730 - x/websocket broadcast result contract

Status: active
Priority: P1
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

