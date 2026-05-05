# x/websocket Beta Evidence

Module: `x/websocket`

Owner: `realtime`

Current status: `experimental`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Hub lifecycle coverage includes stop idempotency, shutdown paths, connection
  joins, leaves, iteration, context cancellation, bounded stop drains, close
  frame emission, room cleanup, and event-handler reentrancy.
- Capacity behavior covers `ErrHubFull`, `ErrRoomFull`, and `ErrHubStopped`.
- Broadcast behavior covers positive paths, stopped-hub no-op behavior,
  result-returning `TryBroadcast*` APIs, and queue-full drop accounting.
- Security and server setup coverage includes config validation, room-password
  validation, method rejection, bad requests, invalid config rejection, weak
  HS256 secrets, explicit Origin policy, bounded admin broadcast bodies, and
  malformed `exp` claims.
- RFC 6455 negative coverage includes RSV bits, unknown opcodes, invalid
  continuation state, non-minimal payload length encodings, and invalid close
  payloads. Fragmented message coverage verifies cumulative read-limit
  enforcement across continuation frames.

## Primer And Boundary State

- Primer: `docs/modules/x-websocket/README.md`
- Manifest: `x/websocket/module.yaml`
- Boundary state: documented and aligned with explicit websocket transport
  wiring outside stable roots.
- Current package status remains `experimental`; the cleanup below does not
  constitute release evidence.

## Current Cleanup State

As of 2026-05-05, the websocket stable-readiness cleanup passes have landed the
following code, test, documentation, and governance work:

- Split transport message handling from the room fanout helper so
  `ServeWSWithConfig` no longer bakes in product broadcast behavior.
- Split room authorization from token authentication, made query tokens
  opt-in, and documented the HS256 helper as a lightweight built-in verifier.
- Replaced query room passwords with the `X-Room-Password` header and added
  room-name validation.
- Renamed capacity and metrics semantics from total connection language to
  room-registration language.
- Hardened broadcast, stop, and write paths with stopped-hub checks, drop
  accounting, and network write deadlines.
- Clarified best-effort close-frame behavior and bounded-reader stream
  semantics.
- Removed unused server/logger/metrics fields, made security event handling
  explicit, and kept metric collection unconditional.
- Tightened secret ownership and log sanitization behavior.
- Required `Sec-WebSocket-Version: 13` during handshake.
- Changed admin broadcast to opt-in, moved it to a separate
  `BroadcastSecret`, and kept URL secrets out of the admin path.
- Added `BroadcastMaxBodyBytes` to bound admin broadcast request bodies.
- Required browser requests with `Origin` to match explicit `AllowedOrigins`;
  non-browser requests without `Origin` continue to skip the origin check.
- Added explicit route-registration errors for nil registrar, nil hub, empty
  websocket path, and invalid broadcast setup.
- Added error-returning constructors for connections and hubs, and made setter
  validation return errors.
- Removed nil-returning/silent constructor and join helpers (`NewConn`,
  `NewHub`, `NewHubWithConfig`, and `Hub.Join`) so setup and capacity failures
  remain visible.
- Made application code responsible for reading websocket secrets and passing
  them through explicit config rather than hidden environment access.
- Added result-returning `TryBroadcastRoom` and `TryBroadcastAll` APIs for
  accepted/dropped job counts.
- Removed non-core public helper API (`Outbound` and
  `ContainsDangerousPatterns`) from the websocket transport surface.
- Hardened `NewHS256TokenAuth` to reject weak secrets and malformed `exp`
  claims while documenting that issuer, audience, `nbf`, and `iat` remain
  outside the built-in helper.
- Removed default stderr writes from the hub by adding caller-provided logging
  with a no-op default.
- Moved `SecurityEventHandler` execution out of event producer hot paths.
- Changed `Shutdown` to stop workers, clear room registrations, reset
  room-registration metrics, and best-effort emit close frames before closing
  registered connections.
- Enforced complete-message read limits across fragmented messages and capped
  retained pooled buffers so large message buffers are discarded.
- Made route-registered `WebSocketConfig` own cloned `Secret`,
  `BroadcastSecret`, and `AllowedOrigins` values after construction.
- Propagated top-level read-limit, message-validation, logging, queue,
  rate-limit, metric, and security-event settings into the owned server and hub
  runtime.
- Added a 64 MiB hard cap for connection, server, top-level, and auth-derived
  read limits.
- Capped retained broadcast snapshot slices so large room fanout snapshots do
  not stay in the hub pool.
- Required built-in room password setters to validate room names.
- Made `SecureRoomAuth` enforce room password strength by default, with
  `SecurityConfig.AllowWeakRoomPasswords` as the explicit opt-out path.
- Updated module manifest, primer docs, and English/Chinese website docs to
  match implemented security defaults, lifecycle semantics, room-registration
  language, and experimental maturity.
- Refreshed the current-head development API snapshot at
  `docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`.

These items reduce technical risk but do not replace release-history,
release-snapshot, or owner-approval evidence.

## Public API Inventory

Current-head public API inventory is recorded at
`docs/extension-evidence/x-websocket-public-api-inventory.md`.

The inventory classifies the exported surface into stable transport candidates,
built-in helper APIs that need explicit owner scope approval, and helper or
diagnostic symbols that need review before a stable promise. It is a freeze
review input only; it does not clear release-history, release-snapshot, or owner
sign-off requirements.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/websocket` API changes.

Release refs:

- none recorded

Required external inputs:

- Older minor release ref that contains `x/websocket`.
- Newer consecutive minor release ref that contains `x/websocket`.
- Confirmation that both refs are immutable release tags or otherwise approved
  release identifiers.

## API Snapshot Evidence

One current-head baseline snapshot is recorded and was refreshed from the
working tree on 2026-05-05. It is useful for comparing the candidate surface
during development, but it is not release evidence and does not clear
`api_snapshot_missing` by itself.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/websocket/... -out /tmp/plumego-x-websocket-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/first-batch/x-websocket-head.snapshot`

Required external inputs:

- Release API snapshot generated from the older minor release ref.
- Release API snapshot generated from the newer consecutive minor release ref.
- Release comparison output checked into the evidence tree or linked from a
  stable artifact location.

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

Required external inputs:

- Named `realtime` owner approval.
- Approval date.
- Approval scope, at minimum: public API surface, security defaults, lifecycle
  semantics, release evidence, and known remaining risks.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/websocket` remains `experimental`.
