# 1309 - Core Drain Interval Contract

State: done
Priority: P1
Primary module: core

## Goal

Make the existing `DrainInterval <= 0` fallback behavior explicit and covered as a stable config contract.

## Scope

- Document that non-positive `DrainInterval` values use the default drain log interval.
- Add public or focused tests that pin this behavior without relying on brittle timing.
- Keep negative values accepted unless a blocker is discovered.

## Non-goals

- Do not reject negative drain intervals.
- Do not change shutdown/drain concurrency semantics.
- Do not introduce new configuration fields.

## Files

- `core/config.go`
- `core/lifecycle_test.go`
- `docs/modules/core/README.md`
- `tasks/cards/done/1309-core-drain-interval-contract.md`

## Tests

- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`

## Docs Sync

Update core module docs if public config behavior is documented there.

## Done Definition

- Public comments/docs describe `DrainInterval <= 0` fallback behavior.
- Core tests cover zero and negative interval fallback behavior.

## Validation

- `gofmt -w core/config.go core/lifecycle.go core/lifecycle_test.go`
- `go test -timeout 20s ./core/...`
- `go test -race -timeout 60s ./core/...`
