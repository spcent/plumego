# 0733 - x/websocket doc sync stable readiness

Status: done
Priority: P2
Primary module: `x/websocket`

## Problem

Source comments and website guides still contain stale maturity, default-safety,
and room-registration language. Documentation must match implemented defaults
before any stable decision.

## Scope

- Remove remaining production-ready language from source comments.
- Update English and Chinese website guides for security defaults, admin
  broadcast opt-in, room registrations, shutdown behavior, and simple HS256
  verifier limits.
- Sync module primer and root docs only for implemented behavior.

## Out of Scope

- Code behavior changes.
- Promotion to beta or stable.

## Validation

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go test -timeout 20s ./x/websocket/...`

## Outcome

- Removed stale "production-ready" source language from `Hub` comments while keeping the module experimental.
- Updated English and Chinese WebSocket guides for explicit origins, disabled admin broadcast by default, separate broadcast secret, query-token rejection, room-password headers, `TryBroadcast*` result APIs, room-registration semantics, shutdown behavior, and simple HS256 verifier limits.
- Updated module primers and the local module README to match implemented security and broadcast semantics.

Completed validations:

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go test -timeout 20s ./x/websocket/...`
