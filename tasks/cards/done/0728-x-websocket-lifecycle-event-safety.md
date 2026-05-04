# 0728 - x/websocket lifecycle event safety

Status: done
Priority: P0
Primary module: `x/websocket`

## Problem

`Hub.Stop` can still block while draining jobs into a full `SendBlock` queue, and
security event handlers are called synchronously from code paths that may hold
the hub lock.

## Scope

- Make hub stop/drain bounded and unable to block forever on connection send
  queues.
- Ensure user-provided security event handlers are not called while the hub lock
  is held.
- Clarify shutdown behavior and health after stop.
- Add focused tests for stop under full queues and event handler reentrancy.

## Out of Scope

- Broadcast API return-value redesign.
- Admin broadcast route defaults.

## Validation

- `go test -race -timeout 60s ./x/websocket/...`
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

## Outcome

- Added bounded hub worker writes so `SendBlock` connections with no explicit
  timeout cannot block hub workers forever.
- Added bounded stop-time drain behavior and drop accounting for jobs that
  cannot be drained before the stop deadline.
- Moved hub capacity security event handler calls outside the hub lock.
- Marked server health unhealthy when the hub has been stopped.
- Added focused tests for stop under a full send queue and reentrant security
  event handlers.
- Validation passed:
  - `go test -race -timeout 60s ./x/websocket/...`
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
