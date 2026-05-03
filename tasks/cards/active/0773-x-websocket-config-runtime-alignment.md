# Card 0773

Milestone: M-003
Recipe: specs/change-recipes/implementation.yaml
Priority: P1
State: active
Primary Module: x/websocket
Owned Files:
- `x/websocket/websocket.go`
- `x/websocket/hub.go`
- `x/websocket/server_config_test.go`
- `x/websocket/websocket_test.go`
- `x/websocket/module.yaml`

Goal:
- Align the top-level `WebSocketConfig` constructor path with the actual `HubConfig` runtime behavior and public API inventory.

Scope:
- Pass all top-level hub-facing runtime knobs through `New`.
- Remove conflicting default behavior between `DefaultWebSocketConfig` and `NewHubWithConfigE` where the top-level constructor owns the values.
- Clarify public comments for exported fields whose behavior is enqueue-level or runtime-level.
- Keep the module manifest aligned with any new or renamed public entrypoints.

Non-goals:
- Do not promote `x/websocket` out of experimental status.
- Do not fabricate release evidence.
- Do not add non-standard-library dependencies.

Files:
- `x/websocket/websocket.go`
- `x/websocket/hub.go`
- `x/websocket/server_config_test.go`
- `x/websocket/websocket_test.go`
- `x/websocket/module.yaml`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required only for public config semantics or manifest changes.

Done Definition:
- `New(WebSocketConfig)` can configure the hub behavior it exposes.
- Tests cover the constructor-to-hub config path.
- Manifest checks pass.
