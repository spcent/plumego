# 1153 - x/websocket public config surface

Status: done
Priority: P1
Primary module: `x/websocket`

## Goal

Remove confusing stable-surface API before promotion work.

## Scope

- Export `RouteRegistrar` and use it in `Server.RegisterRoutes`.
- Remove Hub-only fields from `SecurityConfig`: `RejectOnQueueFull` and
  `MaxConnectionRate`.
- Rename the security-event opt-in flag to `EnableSecurityEvents` in
  `HubConfig` and `WebSocketConfig`; metrics stay always collected.
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

## Outcome

- Exported `RouteRegistrar` and updated `Server.RegisterRoutes` to use it.
- Removed Hub-only `RejectOnQueueFull` and `MaxConnectionRate` fields from
  `SecurityConfig`.
- Renamed the security event opt-in flag to `EnableSecurityEvents` in
  `HubConfig` and `WebSocketConfig`; metrics remain always collected.
- Updated tests, docs, module manifest, and current-head API snapshot.

## Validations

- `rg -n --glob '*.go' 'RejectOnQueueFull' .`
- `rg -n --glob '*.go' 'MaxConnectionRate' .`
- `rg -n --glob '*.go' 'EnableSecurityMetrics' .`
- `rg -n --glob '*.go' 'routeRegistrar|RegisterRoutes' x/websocket .`
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `go run ./internal/checks/module-manifests`
- `go build ./...`
