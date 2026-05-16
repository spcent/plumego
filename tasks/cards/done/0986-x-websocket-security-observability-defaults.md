# 0986 - x/websocket security and observability defaults

Status: done
Priority: P0
State: done
Primary module: `x/websocket`

## Goal

Tighten security defaults and remove library-owned stderr logging from the hub.

## Scope

- Add a max body limit to the admin broadcast endpoint.
- Make origin policy explicit enough for stable docs/tests.
- Replace default stderr hub logger with caller-injected or no-op logging.
- Make security event handler execution semantics non-blocking or explicitly
  named/documented.

## Non-goals

- Changing token algorithms.
- Building a full observability exporter.

## Files

- `x/websocket/websocket.go`
- `x/websocket/server.go`
- `x/websocket/hub.go`
- `x/websocket/websocket_test.go`
- `x/websocket/hub_lifecycle_test.go`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/dependency-rules`

## Docs Sync

Document admin broadcast body limits, origin policy, and logging/event hook
semantics.

## Done Definition

- Admin broadcast cannot read unbounded bodies.
- Hub has no default stderr writes.
- Event hook behavior cannot unexpectedly block hot paths.
- Validation passes.

## Outcome

- Added `BroadcastMaxBodyBytes` with a 1 MiB default and enforced it with
  `http.MaxBytesReader` on the admin broadcast endpoint.
- Changed browser Origin handling so requests with `Origin` require explicit
  `AllowedOrigins`; non-browser requests without `Origin` still skip the check.
- Added caller-injected `HubConfig.Logger` with a no-op default, removing hub
  writes to stderr.
- Moved `SecurityEventHandler` invocation onto the security monitor goroutine
  so event producers cannot block on caller code.
- Updated websocket docs, website guides, and API snapshot evidence.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/dependency-rules`
- `go build ./...`
- `go run ./internal/checks/module-manifests`
