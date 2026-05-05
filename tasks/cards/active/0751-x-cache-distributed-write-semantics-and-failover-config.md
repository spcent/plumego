# 0751 - x/cache distributed write semantics and failover config

Status: active
Priority: P0
Primary module: `x/cache`

## Problem

Distributed replication wording and retry behavior still read stronger than the
implementation can guarantee. Synchronous replication can return an error after
some replicas already accepted a write, and destructive operations can partially
apply. Failover retry attempts and backoff are also hard-coded.

## Scope

- Replace strong-consistency wording with explicit synchronous best-effort
  replica write semantics.
- Document partial side effects for `Set`, `Delete`, and `Clear`.
- Add configurable failover retry attempts and backoff, with conservative
  defaults matching existing behavior.
- Add validation and regression tests for retry attempts, backoff, and partial
  write/destructive-operation errors.
- Sync x/cache module docs.

## Out of Scope

- Distributed transactions or rollback.
- Durable repair queues.
- Promoting `x/cache` to stable.

## Files

- `x/cache/distributed/distributed.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Distributed cache write semantics no longer overclaim strong consistency, and
failover retry behavior is configurable, validated, tested, and documented.
