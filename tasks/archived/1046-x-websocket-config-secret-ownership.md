# 1046 - x/websocket config and secret ownership

Status: done
Priority: P1
State: done
Primary module: `x/websocket`

## Goal

Make top-level server config ownership and propagation explicit enough for a
stable API.

## Scope

- Clone `Secret` and `BroadcastSecret` on `New`.
- Propagate top-level read/message validation settings to `ServerConfig`.
- Expose top-level hub logging/rate/security-event settings or clearly route
  them into `HubConfig`.
- Add tests proving caller slice mutation after `New` cannot change auth
  behavior.

## Non-goals

- New observability exporter.
- New token policy engine.
- Renaming existing config types.

## Files

- `x/websocket/websocket.go`
- `x/websocket/websocket_test.go`
- `x/websocket/server_config_test.go`
- `docs/modules/x-websocket/README.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document any newly exposed top-level config fields.

## Done Definition

- Secrets are owned by the server after construction.
- Top-level config can express stable read/message/hub settings.
- Validation passes.

## Outcome

- `WebSocketConfig` now owns cloned `Secret`, `BroadcastSecret`, and
  `AllowedOrigins` values after `New`.
- Top-level read/message validation settings are passed into registered
  handlers.
- Top-level hub logging, queue, rate-limit, metric, and security event settings
  are passed into the owned hub.
- API snapshot and module docs were refreshed for the newly exposed fields.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`
- `go build ./...`
