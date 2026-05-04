# 0739 - core Lifecycle Test Readiness

State: active
Priority: P2
Primary Module: core

## Goal

Remove fixed startup sleeps from core lifecycle tests in favor of observable
server readiness.

## Scope

- Replace fixed `time.Sleep` startup waits with listener-based serving and an
  HTTP readiness probe.
- Keep test behavior and production runtime unchanged.
- Preserve network-test skip behavior.

## Non-goals

- Do not change core lifecycle implementation.
- Do not alter reference app startup behavior.
- Do not add external dependencies.

## Files

- `core/lifecycle_test.go`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`

## Docs Sync

Not required.

## Done Definition

- Core lifecycle tests no longer rely on fixed sleeps for server startup.
- Normal and race core tests pass.

