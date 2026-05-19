# Card 0962

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0736

Goal:
- Converge the WebSocket public API, metrics, and logging shape before maturity promotion.

Problem:
The package exposes multiple construction and join paths with different failure semantics. `Join` bypasses Hub capacity checks while `TryJoin` enforces them. Metrics names are ambiguous because active connection counts can represent room registrations rather than unique TCP connections. Hub also defaults to a stderr logger instead of a caller-owned logger or no-op logger.

Scope:
- Prefer explicit error-returning public constructors for stable paths.
- Mark bypass-style helpers as legacy, unsafe, or compatibility-only if they must remain.
- Clarify `Join` versus `TryJoin` semantics in code, tests, and docs.
- Rename or document metrics so unique connections and room registrations cannot be confused.
- Replace default stderr logging with injected or no-op logging behavior.
- Follow `AGENTS.md` symbol-change protocol for any exported symbol rename or removal.

Non-goals:
- Do not widen WebSocket into a generic pubsub/event-bus module.
- Do not move WebSocket setup into stable roots.
- Do not complete release evidence in this card.

Files:
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/server.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go test -race -timeout 60s ./x/websocket/...`
- `go vet ./x/websocket/...`

Docs Sync:
- Required for public API, metrics, logger, or compatibility behavior changes.

Done Definition:
- Stable-path constructors and join methods have consistent error semantics.
- Metrics names or docs make unique connection counts versus room registrations explicit.
- Hub no longer writes to stderr by default unless explicitly configured.
- Exported symbol changes, if any, follow the full caller enumeration protocol.

Outcome:
- Added `NewHubWithConfigE` as the error-returning stable hub constructor path.
- Kept `NewHubWithConfig` as the legacy defaulting constructor and documented that distinction.
- Marked `Join` as compatibility-only and kept `TryJoin` as the capacity-enforced stable join path.
- Changed the default hub logger to discard output unless `HubConfig.Logger` is provided.
- Split metrics semantics: `ActiveConnections` is unique open connections; `RoomRegistrations` is connection-room registrations.
- Updated module docs and manifest for the new exported constructor surface.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go test -race -timeout 60s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go run ./internal/checks/module-manifests`
