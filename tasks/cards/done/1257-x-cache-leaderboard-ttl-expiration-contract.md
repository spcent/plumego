# 1257 - x/cache leaderboard ttl and expiration contract

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

Leaderboard TTL semantics are not expressive enough for stable use. A zero
`DefaultTTL` means "use default", and negative values are also normalized to the
default, so callers have no explicit no-expiration contract. Expiration behavior
is lazy plus cleanup based and needs a clear concurrency contract.

## Scope

- Add an explicit no-expiration `DefaultTTL` contract.
- Keep zero `DefaultTTL` as "use default" for compatibility.
- Add tests for no-expiration leaderboards.
- Document lazy cleanup and eventually-expired concurrency semantics.
- Sync x/cache evidence.

## Out of Scope

- Per-call TTL APIs for sorted-set writes.
- Strongly consistent expiration snapshots.
- Redis sorted-set compatibility.

## Files

- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `x/cache/leaderboard/doc.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard callers can intentionally create non-expiring sorted sets through a
documented config value, while zero-value default behavior remains preserved.

## Outcome

- Added `leaderboard.NoExpirationTTL` as the explicit no-expiration
  `DefaultTTL` contract.
- Preserved `DefaultTTL == 0` as "use the package default".
- Rejected ambiguous negative TTL values other than `NoExpirationTTL`.
- Added tests for invalid negative TTL and non-expiring leaderboards.
- Documented lazy plus cleanup based expiration as an eventually-expired
  contract.

## Validation Run

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
