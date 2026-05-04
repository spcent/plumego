# 0727 - core Shutdown Drain Retry

State: active
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

- A first shutdown with an already-canceled context returns an error but leaves
  drain startup retryable.
- A later shutdown with a live context can start drain logging once.

