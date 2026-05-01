# x/websocket Beta Evidence

Module: `x/websocket`

Owner: `realtime`

Current status: `experimental`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Hub lifecycle coverage includes stop idempotency, shutdown paths, connection
  joins, leaves, iteration, bounded worker shutdown writes, and context
  cancellation.
- Capacity behavior covers `ErrHubFull`, `ErrRoomFull`, and `ErrHubStopped`.
- Broadcast behavior covers positive paths, stopped-hub no-op behavior,
  result-returning fanout, partial delivery, total rejection, and queue-full
  drop accounting independent of the metrics toggle.
- Security and server setup coverage includes config validation, room-password
  validation, method rejection, RFC6455 version checks, bad requests, invalid
  config rejection, explicit query-token policy, and separately authorized admin
  broadcast.
- Large-message reads are bounded by `ReadLimit`; `ReadMessageReader` exposes a
  bounded reader for one message and does not claim true unbounded streaming.

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

One current-head baseline snapshot is recorded. It is useful for comparing the
candidate surface during development, but it is not release evidence and does
not clear `api_snapshot_missing` by itself.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out /tmp/plumego-x-websocket-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/websocket/... \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-websocket-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until the
recorded refs and snapshot files come from real releases.

## Owner Sign-Off

Missing. The `realtime` owner must confirm the beta criteria before any
`module.yaml` status change.

## Blockers

Runtime stable-readiness hardening has been recorded in task cards 0739-0745.
The remaining blockers are external release-governance evidence only:

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/websocket` remains `experimental`.
