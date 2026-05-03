# 0727 - x/cache distributed replication failover

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

`x/cache/distributed` advertises replication modes, failover strategies, node
weights, and metrics, but several options are unused or only partially enforced.
Stable users need replication and failover behavior that matches the API.

## Scope

- Implement or remove unused distributed failover strategy branches.
- Make sync replication fail when no healthy replica accepts a write.
- Align `Incr`, `Decr`, and `Append` with the replication contract.
- Track meaningful replication/failover metrics or stop exposing misleading
  values.
- Add tests for unhealthy replicas, strategy differences, and mutation failover.

## Out of Scope

- Cross-process consensus or durable rebalancing.
- Redis adapter atomicity.
- Public status promotion.

## Files

- `x/cache/distributed/distributed.go`
- `x/cache/distributed/hashring.go`
- `x/cache/distributed/distributed_test.go`
- `x/cache/distributed/distributed_bench_test.go` only if benchmark semantics
  need updates

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Distributed replication and failover options are either implemented or no longer
present as misleading public contract, with tests covering the chosen behavior.
