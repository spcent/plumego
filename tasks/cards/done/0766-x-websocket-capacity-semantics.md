# Card 0766

Milestone: M-003
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`
Depends On: 0762

Goal:
- Make connection capacity names match the implementation semantics.

Problem:
`MaxConnections` is described as total connections in some places, but the hub counter is room registrations: one connection joined to N rooms contributes N. That name is misleading for stable users and metrics consumers.

Scope:
- Rename registration-based limits and metrics to `MaxRoomRegistrations` or equivalent.
- Add a separate `MaxActiveConnections` only if the stable API needs a true unique-connection cap.
- Update `CanJoin`, metrics, docs, and tests to use one consistent vocabulary.
- Follow the exported-symbol completeness protocol for every rename.

Non-goals:
- Do not change unrelated hub broadcast behavior.
- Do not leave deprecated wrappers behind.
- Do not promote maturity state.

Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for capacity fields, metrics, and migration notes.

Done Definition:
- No stable-facing comment describes room registrations as total unique connections.
- Metrics, config, and tests use the same terminology.
- Old exported field names are removed or intentionally absent.

Outcome:
- Renamed exported `MaxConnections` config and metrics fields to
  `MaxRoomRegistrations`.
- Renamed internal counters to use room-registration terminology.
- Updated websocket tests, docs, examples, and evidence notes to distinguish
  unique active connections from connection-room registrations.
- Did not add a separate unique-connection cap in this card.
