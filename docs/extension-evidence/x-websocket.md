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

## API Inventory State

Current stable-candidate public symbols are recorded in
`x/websocket/module.yaml`. The manifest now includes the server, hub,
connection, auth, validation, opcode/close-code, and documented error surfaces
that are still exported.

API cleanup recorded before stable:

- `New` no longer accepts unused `debug` or `logger` parameters; caller-owned
  logging remains on `HubConfig.Logger`.
- `HubConfig.EnableMetrics` was removed because runtime counters are always
  recorded as facts.
- The unused private `Hub.metrics` field was removed.
- Internal hub security events are no longer exported as `SecurityEvent`.
- `ContainsDangerousPatterns` was removed from the transport package because
  heuristic XSS/SQL scanning is not part of the websocket transport contract.
- `ServeWSWithConfig` is now the transport serve path with caller-provided
  `MessageHandler`; the previous room fanout behavior is exposed explicitly via
  `ServeRoomFanoutWS`.
- Token authentication and room authorization are split through
  `TokenAuthenticator` and `RoomAuthorizer`. Anonymous mode no longer requires
  a JWT secret; the built-in compact HS256 verifier is exposed as
  `SimpleHS256TokenAuth`.
- Built-in room-password credentials are read from the
  `X-WebSocket-Room-Password` header. `room_password` query parameters are
  ignored.
- Capacity naming now uses `MaxRoomRegistrations` for connection-room
  registrations. Unique active connections remain reported separately as
  `HubMetrics.ActiveConnections`.
- Hub lifecycle now treats `Shutdown(nil)` as `context.Background()` and
  serializes `Stop` with broadcast enqueue.
- Connection writes now apply a configurable write deadline. `WriteClose` is
  documented as best-effort close-frame delivery followed by TCP close.
- `ReadMessageReader` is explicitly documented as a bounded buffered reader,
  not a zero-copy or unbounded streaming API. `ReadMessage` is documented as a
  full in-memory read with an owned payload copy.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/websocket` API changes.

Release refs:

- none recorded

Stable-governance requirements:

- `older_minor_release_ref`: a real tag or immutable release ref that resolves
  to a git commit and already includes the stable-candidate API surface.
- `newer_minor_release_ref`: the next real minor release ref after
  `older_minor_release_ref`; it must also resolve to a git commit.
- `api_delta`: the exported `x/websocket` API must be unchanged between those
  two release refs, except for explicitly documented non-breaking additions.
- `runtime_gate`: the stable-readiness runtime gates from cards 0761-0772 must
  be complete before the release refs are evaluated for promotion.

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

Required release snapshot refs:

- `docs/extension-evidence/snapshots/<older_minor_release_ref>/x-websocket.snapshot`
- `docs/extension-evidence/snapshots/<newer_minor_release_ref>/x-websocket.snapshot`

Current-head snapshots must remain clearly labeled as development baselines.
They must not be moved into release snapshot slots until they are generated
from real release refs.

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

Required sign-off record:

- Owner: `realtime`
- Scope: exported API surface, runtime stable-readiness gates, documentation
  caveats, and remaining operational limits.
- Location: this evidence file or a linked release artifact checked into
  `docs/extension-evidence/`.

## Blockers

Runtime stable-readiness hardening has been recorded in task cards 0739-0760.
Additional stable-readiness cleanup is tracked in cards 0761-0772. The
remaining governance blockers are:

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/websocket` remains `experimental`.
