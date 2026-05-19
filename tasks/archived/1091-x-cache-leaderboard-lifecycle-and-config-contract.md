# 1091 - x/cache leaderboard lifecycle and config contract

Status: done
Priority: P1
State: done
Primary module: `x/cache`

## Problem

Leaderboard config validation mutates caller-owned config, and operations after
`Close` do not have an explicit package-level contract. Stable readiness needs
clear constructor normalization and lifecycle behavior.

## Scope

- Normalize leaderboard config on a private copy instead of mutating caller
  input.
- Add closed-state guards for leaderboard-specific operations.
- Keep `Close` nil-safe and idempotent.
- Document closed-cache and config-normalization behavior.
- Add tests for caller config immutability and post-close operations.

## Out of Scope

- Changing stable `store/cache` lifecycle behavior.
- Redis sorted-set compatibility decisions.

## Files

- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard construction no longer mutates caller config and post-close
leaderboard operations fail with a documented stable cache lifecycle error.

## Outcome

- Moved leaderboard config defaulting to a private normalized constructor copy.
- Kept caller-owned `LeaderboardConfig` values unchanged.
- Added `leaderboard.ErrClosed` and closed-state checks for leaderboard-specific
  operations and leaderboard `Clear`.
- Documented config normalization and post-close behavior.
- Added regression tests for caller config immutability and post-close errors.

## Validation Run

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
