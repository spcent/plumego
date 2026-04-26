# x/websocket Beta Evidence

Module: `x/websocket`

Owner: `realtime`

Current status: `experimental`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Hub lifecycle coverage includes stop idempotency, shutdown paths, connection
  joins, leaves, iteration, and context cancellation.
- Capacity behavior covers `ErrHubFull`, `ErrRoomFull`, and `ErrHubStopped`.
- Broadcast behavior covers positive paths and stopped-hub no-op behavior.
- Security and server setup coverage includes config validation, room-password
  validation, method rejection, bad requests, and invalid config rejection.

## Primer And Boundary State

- Primer: `docs/modules/x-websocket/README.md`
- Manifest: `x/websocket/module.yaml`
- Boundary state: documented and aligned with explicit websocket transport
  wiring outside stable roots.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/websocket` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

Missing. Generate snapshots with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out /tmp/plumego-x-websocket-api.snapshot
```

Snapshot refs:

- none recorded

## Owner Sign-Off

Missing. The `realtime` owner must confirm the beta criteria before any
`module.yaml` status change.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/websocket` remains `experimental`.
