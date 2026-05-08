# 0829 - x/cache distributed constructor lifecycle

Status: done
Priority: P1
Primary module: `x/cache`

## Problem

`x/cache/distributed` exposes public configuration and lifecycle entrypoints,
but construction silently ignores invalid nodes and health-check lifecycle calls
are not idempotent. Stable users need deterministic constructor errors and safe
`Close` behavior.

## Scope

- Add explicit distributed config validation without changing stable roots.
- Make constructor failures observable.
- Preserve existing `New` compatibility if a safer constructor is added.
- Make health checker and distributed cache close paths idempotent.
- Add negative-path tests for invalid config, nil/duplicate/empty nodes, and
  repeated lifecycle calls.

## Out of Scope

- Full failover strategy semantics.
- Redis or leaderboard behavior.
- Extension status promotion.

## Files

- `x/cache/distributed/distributed.go`
- `x/cache/distributed/node.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md` if public constructor semantics change

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go vet ./x/cache/distributed`
- `go run ./internal/checks/dependency-rules`

## Done Definition

Invalid distributed cache construction is test-covered, lifecycle methods can be
called safely more than once, and boundary checks still pass.

## Outcome

- Added `distributed.NewWithConfig` and `Config.Validate` for observable
  construction failures.
- Kept `distributed.New` as a compatibility helper that returns `nil` on
  construction error.
- Made distributed cache close and health checker start/stop paths idempotent.
- Added negative-path and lifecycle regression tests.

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go vet ./x/cache/distributed`
