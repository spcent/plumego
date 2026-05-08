# 0735 - core Shutdown Drain Contract

State: done
Priority: P1
Primary Module: core

## Goal

Align shutdown drain documentation and tests with the current retryable drain
semantics.

## Scope

- Document that one drain attempt is active at a time, while canceled drain with
  active connections remains retryable.
- Add focused regression coverage for drain cancellation releasing the latch.
- Keep runtime behavior unchanged.

## Non-goals

- Do not change `http.Server.Shutdown` delegation.
- Do not change log message shape.
- Do not add public API.

## Files

- `core/lifecycle_test.go`
- `docs/modules/core/README.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go run ./internal/checks/module-manifests`

## Docs Sync

Required in `docs/modules/core/README.md`.

## Done Definition

- Core docs no longer claim unconditional drain once-only behavior.
- Tests prove canceled drain attempts with active connections can be retried.

## Outcome

- Updated the core module lifecycle matrix to describe one active drain attempt
  at a time with retry after cancellation while connections remain active.
- Added regression coverage for canceled drain attempts releasing the start
  latch and allowing retry with a live context.
- Verified with `go test -timeout 20s ./core/...` and
  `go run ./internal/checks/module-manifests`.
