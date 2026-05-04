# 0733 - x/cache leaderboard lifecycle limits range

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

`x/cache/leaderboard` still has non-idempotent close behavior, soft
`MaxLeaderboards` enforcement, unused range errors, and incomplete metric
coverage.

## Scope

- Make `MemoryLeaderboardCache.Close` idempotent and nil-safe.
- Enforce `MaxLeaderboards` without a concurrent over-admission window.
- Use `ErrInvalidRange` for explicitly invalid score/rank ranges.
- Count `ZIncrBy` as a mutation metric or add a clear metric field.
- Add regression tests for lifecycle, concurrent limits, invalid ranges, and
  metrics.

## Out of Scope

- Distributed leaderboard support.
- Broad skiplist algorithm rewrite.
- Redis sorted-set compatibility expansion.

## Files

- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `x/cache/leaderboard/test_helpers_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard lifecycle and limit behavior are deterministic, invalid ranges have
a public error contract, and metrics cover documented mutations.
