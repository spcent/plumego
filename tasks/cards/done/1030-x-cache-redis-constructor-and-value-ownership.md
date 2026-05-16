# 1030 - x/cache redis constructor and value ownership

Status: done
Priority: P1
State: done
Primary module: `x/cache`

## Problem

`NewAdapterWithOptions` accepts invalid option values without construction-time
feedback, and adapter value ownership is delegated to caller-provided clients.
Stable readiness needs a clearer option contract and copy semantics.

## Scope

- Add a validating Redis adapter constructor for new stable-oriented call sites.
- Keep existing constructors source-compatible.
- Reject invalid max-key and clear-prefix options in the validating path.
- Copy values on adapter `Get` and `Set` so adapter ownership does not depend
  on client behavior.
- Add regression tests for validation and slice aliasing.

## Out of Scope

- Removing exported compatibility fields.
- Adding a concrete Redis dependency.
- Full Redis integration matrix.

## Files

- `x/cache/redis/redis.go`
- `x/cache/redis/redis_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

New Redis adapter call sites have a validation-capable constructor and adapter
value ownership is independent of the wrapped client.

## Outcome

- Added `NewValidatedAdapterWithOptions` for construction-time client and
  option validation.
- Kept existing Redis adapter constructors source-compatible.
- Rejected nil clients, negative max-key lengths, and invalid explicit clear
  prefixes in the validating path.
- Copied byte slices on adapter `Set` and `Get`.
- Documented the validating constructor and adapter value ownership contract.

## Validation Run

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
