# Card 0755

Milestone: M-003
Recipe: specs/change-recipes/bugfix.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/errors.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0754

Goal:
- Make hub join, capacity, and shutdown semantics precise and panic-free.

Problem:
`TryJoin` accepts nil connections and can later panic during broadcast. Duplicate joins are checked after hub capacity, so an already-joined connection can be rejected when the hub is full. `Shutdown` closes connections but does not clear manually registered room state.

Scope:
- Reject nil `*Conn` values at public hub boundaries.
- Make duplicate `TryJoin` idempotency win over capacity checks.
- Clear hub room state and registration counters during `Shutdown`.
- Add focused tests for nil joins, duplicate-at-capacity joins, and post-shutdown metrics.

Non-goals:
- Do not introduce tenant/session ownership.
- Do not change broadcast delivery guarantees.
- Do not add new hub-level dependencies.

Files:
- `x/websocket/hub.go`
- `x/websocket/errors.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for join and shutdown lifecycle semantics.

Done Definition:
- Nil connections cannot be registered or iterated into broadcast paths.
- Rejoining the same connection in the same room returns nil even when capacity is reached.
- `Shutdown` leaves room counts and registration metrics at zero.

Outcome:
- Added `ErrNilConn` and rejected nil hub join inputs before registration.
- Made duplicate `TryJoin` calls idempotent before rate limiting and capacity checks.
- Hardened metrics, range, broadcast, leave, and remove paths against nil connection entries.
- Made successful `Shutdown` clear room maps and room-registration counters after closing connections.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go test -race -timeout 60s ./x/websocket/...`
  - `go vet ./x/websocket/...`
