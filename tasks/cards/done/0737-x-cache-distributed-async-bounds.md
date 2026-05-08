# 0737 - x/cache distributed async bounds

Status: done
Priority: P0
Primary module: `x/cache`

## Problem

Async replication uses unbounded background goroutines with
`context.Background()`. Stable behavior needs explicit runtime bounds for
secondary writes.

## Scope

- Add async replication timeout configuration.
- Use bounded contexts for async secondary `Set`, `Incr`, and `Append`.
- Preserve best-effort caller behavior while recording timeout/failure metrics.
- Add tests for timeout-driven async failure metrics.

## Out of Scope

- Persistent queues.
- Retry policies.
- Worker pool sizing knobs beyond timeout bounding.

## Files

- `x/cache/distributed/distributed.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Async replica writes no longer run with an unbounded context, and timeout
failures are visible through existing replication metrics.

## Outcome

- Added `Config.AsyncReplicationTimeout` with a bounded default.
- Applied bounded contexts to async `Set`, `Incr`, and `Append` replication.
- Recorded timeout-driven async failures through existing replication metrics.
- Documented async replication timeout behavior.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
