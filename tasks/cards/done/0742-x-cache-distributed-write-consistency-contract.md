# 0742 - x/cache distributed write consistency contract

Status: done
Priority: P0
Primary module: `x/cache`

## Problem

Synchronous `Incr` and `Append` can mutate the primary before a secondary
replica fails. Async replication is bounded, but failure visibility remains
metrics-only. Stable readiness needs an explicit write-consistency contract and
tests for partial-write behavior.

## Scope

- Make partial synchronous mutation behavior explicit in code and docs.
- Add tests proving primary-visible partial writes when secondary replication
  fails.
- Tighten async timeout behavior so internally invalid timeout state does not
  create unbounded background work.
- Document async failure visibility and remaining repair-hook blocker.

## Out of Scope

- Two-phase commit or rollback across replicas.
- Durable repair queues.
- Public callback API design.

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

Distributed write consistency is no longer implicit: partial success and async
failure visibility are tested and documented as current experimental behavior.

## Outcome

- Added regression coverage for synchronous `Incr` and `Append` partial
  outcomes when primary mutation succeeds and secondary replication fails.
- Made the async replication context defensively fall back to the default
  timeout even if internal state is invalid.
- Documented partial synchronous mutation visibility and async metrics-only
  failure reporting.
- Updated x/cache evidence with fourth-pass distributed coverage.

## Validation Run

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
