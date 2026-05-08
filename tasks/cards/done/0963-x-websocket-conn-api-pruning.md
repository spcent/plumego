# 0963 - x/websocket conn API pruning

Status: done
Priority: P1
Primary module: `x/websocket`

## Goal

Remove the nil-returning connection constructor from the public API.

## Scope

- Remove `NewConn`.
- Migrate all callers and examples to `NewConnE`.
- Update manifest and tests.
- Follow AGENTS.md symbol-change protocol.

## Non-goals

- Changing frame read/write behavior.
- Changing server-side hijack construction.

## Files

- `x/websocket/conn.go`
- `x/websocket/*_test.go`
- direct non-test callers found by symbol search
- websocket docs affected by examples

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`

## Docs Sync

Examples should use `NewConnE` and handle constructor errors.

## Done Definition

- `NewConn` is removed from code and manifest.
- All callers handle `NewConnE` errors.
- Validation passes.

## Outcome

- Removed the nil-returning `NewConn` wrapper and migrated direct test callers
  to `NewConnE`.
- Updated connection examples, docs, module manifest, and the x/websocket API
  snapshot to expose only `NewConnE`.
- Preserved the server hijack path via `newConnFromHijack` and clarified its
  allocation comment.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go build ./...`
- `go run ./internal/checks/module-manifests`
