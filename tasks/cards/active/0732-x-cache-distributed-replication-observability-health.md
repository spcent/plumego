# 0732 - x/cache distributed replication observability health

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

Distributed async replication is best-effort but currently loses secondary
errors without any observable signal. Health checks are also fixed to a single
`Exists` probe.

## Scope

- Track replication lag and replication failure counts in metrics.
- Keep async replication best-effort but make failures observable.
- Add configurable health probe behavior without changing stable roots.
- Add tests for async replica failures and custom health probes.

## Out of Scope

- Durable repair queues.
- Consensus or rollback semantics.
- Provider-specific Redis health behavior.

## Files

- `x/cache/distributed/distributed.go`
- `x/cache/distributed/node.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Replication metrics include observable lag/failure state, async errors are not
silent to metrics, and health probes are caller-configurable.
