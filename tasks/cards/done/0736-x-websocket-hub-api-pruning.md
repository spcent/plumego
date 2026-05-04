# 0736 - x/websocket hub API pruning

Status: done
Priority: P1
Primary module: `x/websocket`

## Goal

Remove silent hub APIs from the public surface so capacity, lifecycle, and
configuration failures are visible.

## Scope

- Remove `NewHub`, `NewHubWithConfig`, and `Join`.
- Migrate all callers to `NewHubE`, `NewHubWithConfigE`, and `TryJoin`.
- Update manifest, tests, examples, and docs.
- Follow AGENTS.md symbol-change protocol for every removed exported symbol.

## Non-goals

- Reworking worker internals.
- Changing broadcast semantics.

## Files

- `x/websocket/hub.go`
- `x/websocket/*_test.go`
- direct non-test callers found by symbol search
- websocket docs affected by examples

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`

## Docs Sync

Examples should use error-returning constructors and `TryJoin`.

## Done Definition

- Removed symbols no longer appear in `x/websocket` Go sources or docs as
  callable APIs.
- All call sites handle returned errors.
- Validation passes.

## Outcome

- Removed `NewHub`, `NewHubWithConfig`, and `Hub.Join`.
- Migrated hub construction to `NewHubE`/`NewHubWithConfigE` and room joins to
  `TryJoin` across tests, scaffold templates, docs, manifests, and API
  snapshot evidence.
- Updated the stopped-hub test to assert explicit `ErrHubStopped` from
  `TryJoin` instead of relying on the removed silent no-op behavior.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`
