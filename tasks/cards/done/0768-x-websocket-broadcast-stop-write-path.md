# 0768 - x/websocket broadcast stop write path

Status: done
Priority: P1
Primary module: `x/websocket`

## Problem

Broadcast can pass the initial stopped check and enqueue after workers have
exited during `Stop`. Socket writes also lack a network write deadline, so a
slow client can pin a writer goroutine beyond send-queue timeout semantics.

## Scope

- Harden broadcast enqueue against concurrent stop.
- Ensure dropped broadcast facts are observable independent of the metrics flag.
- Add connection write deadline support for actual network writes.
- Add concurrency and timeout-focused tests.

## Out of Scope

- Full backpressure redesign.
- External metrics exporter wiring.

## Validation

- `go test -race -timeout 60s ./x/websocket/...`
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

## Outcome

- Added network write deadlines around frame writes using the connection send
  timeout.
- Hardened dispatch against stopped hubs and closed quit channels.
- Counted broadcast drops regardless of the metrics flag.
- Added focused tests for write deadlines and post-stop drop accounting.
