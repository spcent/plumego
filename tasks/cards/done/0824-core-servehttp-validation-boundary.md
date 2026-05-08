# 0824 - core ServeHTTP Validation Boundary

State: done
Priority: P1
Primary Module: core

## Goal

Make the split between handler preparation and server preparation concrete:
`ServeHTTP` must stay `net/http` compatible and ignore server-only config
validation, while `Prepare` must continue to reject invalid server config before
server allocation.

## Scope

- Add focused regression coverage for invalid server config with `ServeHTTP`.
- Remove stale handler preparation commentary if touched.
- Document the server-validation boundary in the core module README.

## Non-goals

- Do not make `ServeHTTP` allocate or validate `http.Server`.
- Do not change route matching behavior.
- Do not change public API.

## Files

- `core/app.go`
- `core/app_test.go`
- `docs/modules/core/README.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go run ./internal/checks/dependency-rules`

## Docs Sync

Required in `docs/modules/core/README.md`.

## Done Definition

- Tests prove `ServeHTTP` can serve a request even when server-only config would
  make `Prepare` fail.
- Mutation is still frozen after `ServeHTTP`.
- `Prepare` still returns the canonical invalid-config error.

## Outcome

- Added regression coverage for `ServeHTTP` serving with invalid server-only
  config while `Prepare` still rejects the same app config.
- Removed stale handler-once testing commentary.
- Updated core module docs to describe `ServeHTTP` as handler-only preparation
  and `Prepare` as the server-validation path.
- Verified with `go test -timeout 20s ./core/...` and
  `go run ./internal/checks/dependency-rules`.
