# 0752 - x/websocket security event dispatcher

Status: done
Priority: P0
Primary module: `x/websocket`

## Goal

Keep security event handlers observable without unbounded goroutines or process
panic risk.

## Scope

- Replace per-event `go handler(event)` with bounded handler dispatch.
- Recover handler panics and log only through the caller-provided logger.
- Keep producers and `Stop`/`Shutdown` nonblocking on user handlers.
- Add tests for handler panic recovery and bounded handler dispatch.

## Non-goals

- New external observability exporter.
- Guaranteed delivery of all security events.
- Changing security event payload shape.

## Files

- `x/websocket/hub.go`
- `x/websocket/hub_lifecycle_test.go`
- `docs/modules/x-websocket/README.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document best-effort bounded event delivery and panic recovery semantics.

## Done Definition

- Security event handlers are bounded and recovered.
- Stop/shutdown remain nonblocking on handlers.
- Validation passes.

## Outcome

- Replaced per-event handler goroutines with a bounded internal handler queue
  and dispatcher.
- Recovered handler panics and logged them only through caller-provided loggers.
- Kept producers and hub stop/shutdown nonblocking on user handlers.
- Documented best-effort bounded delivery.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
