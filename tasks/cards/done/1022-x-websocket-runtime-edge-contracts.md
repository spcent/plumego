# 1022 - x/websocket runtime edge contracts

Status: done
Priority: P0
State: done
Primary module: `x/websocket`

## Goal

Remove remaining runtime edge-case panics and misleading close results.

## Scope

- Make `WriteClose` return close-frame write errors while still closing TCP.
- Make `WriteMessageContext(nil, ...)` safe and deterministic.
- Prevent `SetPongWait` / `pongMonitor` from creating zero-duration tickers.
- Add focused tests for each edge case.

## Non-goals

- Full WebSocket closing handshake wait.
- Changing public close-code constants.
- Reworking writer pump architecture.

## Files

- `x/websocket/conn.go`
- `x/websocket/writer.go`
- `x/websocket/writer_pump_test.go`
- `x/websocket/websocket_extended_test.go`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Update comments where close or runtime setter behavior changes.

## Done Definition

- `WriteClose` reports write failures.
- Nil write contexts do not panic.
- Pong wait cannot create invalid tickers.
- Validation passes.

## Outcome

- `WriteClose` now returns close-frame write errors after still closing TCP.
- `WriteMessageContext(nil, ...)` treats nil context as `context.Background()`.
- `SetPongWait` rejects durations that would create zero-duration monitor
  tickers, and `pongMonitor` defensively falls back to the default wait if the
  stored value is invalid.
- Added focused regression tests for close-frame write failure, nil context,
  and tiny pong waits.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
