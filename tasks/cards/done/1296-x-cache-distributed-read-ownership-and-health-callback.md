# 1296 - x/cache distributed read ownership and health callback

Status: done
Priority: P0
Primary module: `x/cache`

## Problem

Distributed cache now owns caller byte slices on writes, but `Get` still returns
the wrapped node cache buffer directly on primary and failover paths. Health
checker failure callbacks also do not recover panics, unlike async replication
drop callbacks.

## Scope

- Copy bytes returned from distributed `Get` primary, failover, and retry paths.
- Add tests proving caller mutation of returned `Get` buffers does not mutate
  wrapped node cache values.
- Recover panics from health checker failure callbacks.
- Add focused tests for health callback panic isolation.
- Sync x/cache docs and evidence.

## Out of Scope

- Durable distributed repair.
- Changing failover selection semantics.
- Promoting distributed cache to stable.

## Files

- `x/cache/distributed/distributed.go`
- `x/cache/distributed/node.go`
- `x/cache/distributed/distributed_test.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`

## Validation

- `go test -race -timeout 60s ./x/cache/distributed`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Distributed read/write byte ownership is consistent, health callback panics do
not kill the checker goroutine, and both behaviors are tested and documented.

## Outcome

- Distributed `Get` now copies bytes on primary, failover, and retry returns.
- Health-check failure callback panics are recovered.
- Added regression coverage for read byte ownership, callback panic isolation,
  and async timeout metrics waiting.
- Validation passed:
  - `go test -race -timeout 60s ./x/cache/distributed`
  - `go test -timeout 20s ./x/cache/...`
  - `go vet ./x/cache/...`
