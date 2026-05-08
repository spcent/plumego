# Card 0783

Milestone: M-003
Recipe: specs/change-recipes/implementation.yaml
Priority: P1
State: active
Primary Module: x/websocket
Owned Files:
- `x/websocket/websocket.go`
- `x/websocket/hub.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
- `x/websocket/module.yaml`

Goal:
- Rename misleading hub and server config knobs before API freeze.

Scope:
- Rename `MaxConnectionRate` to a room-registration/join-rate name.
- Rename `EnableSecurityMetrics` to an event-monitoring name while keeping metrics counters always-on.
- Rename `RejectOnQueueFull` to reflect queue-full reporting/drop accounting behavior.
- Update all in-repo callers, tests, docs, and comments in the same commit.

Non-goals:
- Do not change broadcast delivery guarantees.
- Do not add durable queues or acknowledgements.

Files:
- `x/websocket/websocket.go`
- `x/websocket/hub.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
- `x/websocket/module.yaml`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for renamed public config fields.

Done Definition:
- Config names match actual runtime behavior and no old names remain in Go code.
