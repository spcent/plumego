# 0727 - core Shutdown Drain Retry

State: done
Priority: P0
Primary Module: core

## Goal

Ensure a canceled shutdown context does not permanently consume the active
connection drain-once latch before drain work can actually start.

## Scope

- Update connection drain startup to reject already-canceled contexts without
  setting the drain-started latch.
- Add regression coverage for canceled-context shutdown followed by a live
  shutdown attempt.
- Keep `Shutdown` error wrapping unchanged.

## Non-goals

- Do not change `http.Server.Shutdown` delegation semantics.
- Do not add new exported errors or lifecycle states.
- Do not change logging message shape.

## Files

- `core/lifecycle.go`
- `core/lifecycle_test.go`

## Tests

- `go test -timeout 20s ./core/...`
- `go run ./internal/checks/dependency-rules`

## Docs Sync

Not required unless behavior text changes.

## Done Definition

- A first shutdown with an already-canceled context leaves drain startup
  retryable.
- A later shutdown with a live context can start drain logging once.

## Outcome

- `connectionTracker.startDrain` now refuses already-canceled contexts before
  setting the drain-started latch.
- Drain now releases the latch when context cancellation ends the drain while
  active connections remain, allowing a later shutdown to retry.
- Added shutdown regression coverage for canceled-context drain startup followed
  by a live shutdown context.
- Verified with `go test -timeout 20s ./core/...` and
  `go run ./internal/checks/dependency-rules`.
