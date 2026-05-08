# Card 0758

Milestone: M-003
Recipe: specs/change-recipes/refactor.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files:
- `x/websocket/websocket.go`
- `x/websocket/security.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`
Depends On: 0757

Goal:
- Consolidate runtime configuration so stable callers can predict defaults and metrics names.

Problem:
`New(WebSocketConfig, ...)` does not apply the same defaults as `DefaultWebSocketConfig`, and `SecurityConfig` contains fields that belong to hub/runtime behavior. `SecurityMetrics.InvalidJWTSecrets` counts JWT verification failures, not bad secrets.

Scope:
- Normalize zero-valued `WebSocketConfig` runtime fields through the same defaulting rules as `DefaultWebSocketConfig`.
- Remove or relocate ineffective runtime fields from `SecurityConfig`.
- Rename security metrics to reflect JWT verification failures accurately.
- Update tests and docs for the consolidated configuration contract.
- Follow `AGENTS.md §7.1` for exported field renames/removals.

Non-goals:
- Do not change `HubConfig` ownership of connection limits or queue behavior.
- Do not add new metrics backends.
- Do not promote module maturity.

Files:
- `x/websocket/websocket.go`
- `x/websocket/security.go`
- `x/websocket/auth.go`
- `x/websocket/*_test.go`
- `docs/modules/x-websocket/README.md`

Tests:
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

Docs Sync:
- Required for config defaulting and security metrics names.

Done Definition:
- A minimal valid `WebSocketConfig` gets deterministic queue, timeout, path, and broadcast body defaults.
- `SecurityConfig` no longer exposes runtime knobs that do not affect `SecureRoomAuth`.
- Security metric names describe what they actually count.

Outcome:
- Normalized minimal `WebSocketConfig` values through deterministic defaults for workers, queues, timeout, routes, and broadcast body limits.
- Removed hub/runtime fields from `SecurityConfig`; those knobs remain owned by `HubConfig`.
- Renamed `SecurityMetrics.InvalidJWTSecrets` to `JWTVerificationFailures`.
- Updated tests and docs for the new defaults and metric naming.
- Validation passed:
  - `go test -timeout 20s ./x/websocket/...`
  - `go vet ./x/websocket/...`
  - `go build ./...`
