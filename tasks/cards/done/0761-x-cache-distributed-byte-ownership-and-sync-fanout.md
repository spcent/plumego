# 0761 - x/cache distributed byte ownership and sync fanout

Status: active
Priority: P0
Primary module: `x/cache`

## Problem

Distributed cache write ownership is not uniform. Async secondary replication
copies byte slices before scheduling, but primary and synchronous replica paths
pass caller-owned slices directly to underlying caches. Synchronous `Set` also
starts one goroutine per selected replica with no local concurrency bound.

## Scope

- Make distributed `Set` and `Append` own caller byte slices before crossing
  node cache boundaries.
- Keep async secondary byte copies intact.
- Bound synchronous `Set` replica fanout using existing replication config.
- Add tests proving caller mutation after write does not affect stored values.
- Add tests proving synchronous fanout respects the configured concurrency
  bound.
- Sync x/cache docs and evidence.

## Out of Scope

- Durable async repair.
- Changing best-effort synchronous partial-write semantics.
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

Distributed cache owns caller byte slices consistently, sync fanout has a
bounded execution path, and the behavior is covered by tests and docs.

## Outcome

- Copied caller-owned byte slices before distributed `Set` and `Append` pass
  data to node caches.
- Replaced sync `Set` per-replica goroutine fanout with a worker pool bounded by
  the configured replication concurrency value.
- Added regression tests with a cache that intentionally retains caller slices.
- Added regression coverage for bounded synchronous `Set` fanout.
- Updated x/cache module docs and evidence.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
