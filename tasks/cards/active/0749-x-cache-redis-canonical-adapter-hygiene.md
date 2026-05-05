# 0749 - x/cache redis canonical adapter hygiene

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

Redis adapter behavior is split across legacy and validated constructors, and
`Append` still passes caller byte slices directly to the wrapped client.

## Scope

- Make the validated constructor the documented canonical path.
- Copy byte slices on `Append`, matching `Set`.
- Add a public adapter capability check for atomic/cache extension operations.
- Add tests for append copy ownership and capability reporting.
- Document legacy constructor compatibility boundaries.

## Out of Scope

- Removing legacy constructors or exported compatibility fields.
- Adding a concrete Redis dependency.
- Producing a real Redis client matrix.

## Files

- `x/cache/redis/redis.go`
- `x/cache/redis/redis_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/redis`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Redis adapter byte ownership is consistent across mutation methods and new call
sites have a clearer canonical constructor/capability contract.

## Outcome

Pending.
