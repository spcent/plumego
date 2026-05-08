# 0756 - Core Server Prepared Helper Simplification

State: done
Priority: P3
Primary module: core

## Goal

Reduce the readability cost in the server-prepared helper while preserving the existing lifecycle contract.

## Scope

- Simplify `markServerPreparedIfInstalled` in `core/http_handler.go`.
- Keep behavior unchanged: if an HTTP server is already installed, mark preparation as `PreparationStateServerPrepared` and return true.
- Avoid public API changes.

## Non-goals

- Do not change prepare/shutdown semantics.
- Do not rename exported lifecycle types.
- Do not refactor unrelated core locking paths.

## Files

- `core/http_handler.go`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

## Docs Sync

No docs update required; behavior remains unchanged.

## Done Definition

- The helper has one clear lock path.
- Core normal, race, and vet checks pass.

## Validation

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`
