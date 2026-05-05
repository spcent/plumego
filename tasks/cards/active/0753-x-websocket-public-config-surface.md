# 0753 - x/websocket public config surface

Status: active
Priority: P1
Primary module: `x/websocket`

## Goal

Remove confusing stable-surface API before promotion work.

## Scope

- Export `RouteRegistrar` and use it in `Server.RegisterRoutes`.
- Remove Hub-only fields from `SecurityConfig`: `RejectOnQueueFull` and
  `MaxConnectionRate`.
- Rename `EnableSecurityMetrics` to `EnableSecurityEvents` in `HubConfig` and
  `WebSocketConfig`; metrics stay always collected.
- Update all callers, tests, docs, manifest, and current-head API snapshot.

## Non-goals

- Full stable promotion.
- New security event API.
- Changing metrics field names in `HubMetrics`.

## Files

- `x/websocket/hub.go`
- `x/websocket/security.go`
- `x/websocket/websocket.go`
- `x/websocket/module.yaml`
- `docs/modules/x-websocket/README.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`

## Docs Sync

Update docs and API inventory evidence for renamed/removed exported fields.

## Done Definition

- Public route registrar is named.
- SecurityConfig contains auth/security helper fields only.
- Security event flag name matches behavior.
- Validation passes.
