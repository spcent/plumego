# Card 0770

Milestone: M-003
Recipe: specs/change-recipes/security-hardening.yaml
Priority: P0
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/server.go`
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0765, 0766

Goal:
- Add a stable room-name policy to prevent unbounded or unsafe room identifiers.

Problem:
Room names are read directly from query parameters and used as map keys, log values, and metrics dimensions without length or character validation.

Scope:
- Add a default room-name policy with maximum length and allowed characters.
- Allow callers to provide a custom room-name validator when needed.
- Apply the same policy to WebSocket joins and admin broadcast room targeting.
- Keep the default room behavior explicit and documented.

Non-goals:
- Do not add tenant or business-channel logic.
- Do not change hub storage away from room maps.
- Do not introduce regex-heavy hot paths if a simple validator is enough.

Files:
- `x/websocket/server.go`
- `x/websocket/hub.go`
- `x/websocket/websocket.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for default room policy and custom validation.

Done Definition:
- Empty, oversized, control-character, and unsafe room names are rejected before hub registration.
- Admin broadcast room targeting uses the same validator.
- Tests cover accepted and rejected room names.

Outcome:
- Added the default room-name policy, custom `RoomNameValidator`, handshake
  validation, hub join validation, and admin broadcast room-target validation.
- Documented the default allowed character set and maximum length.
- Verified with `go test -timeout 20s ./x/websocket/...`, `go vet
  ./x/websocket/...`, `go build ./...`, and `go run
  ./internal/checks/module-manifests`.
