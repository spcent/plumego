# 0763 - Core Connection Tracker Open Count

State: done
Priority: P2
Primary module: core

## Goal

Align connection tracker naming with actual `ConnState` semantics.

## Scope

- Rename internal `connectionTracker.active` semantics to open connections.
- Rename internal helpers and test assertions accordingly.
- Use an `open_connections` log field for drain logging.

## Non-goals

- Do not change public APIs.
- Do not change when the tracker increments or decrements.
- Do not add request-level activity tracking.

## Files

- `core/lifecycle.go`
- `core/lifecycle_test.go`
- `tasks/cards/active/0763-core-connection-tracker-open-count.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`

## Docs Sync

No docs update required unless public wording mentions active connection counts.

## Done Definition

- Internal names and log field describe open connection count.
- Existing behavior remains covered and core tests pass.

## Validation

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`
