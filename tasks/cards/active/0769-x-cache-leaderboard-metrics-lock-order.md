# 0769 - x/cache leaderboard metrics lock order

Status: active
Priority: P0
Primary module: `x/cache`

## Problem

Leaderboard mutation paths hold `sortedSet.mu` and then take `metrics.mu`, while
`GetLeaderboardMetrics` takes `metrics.mu` and then each `sortedSet.mu`. This
reverse lock ordering can deadlock under concurrent metrics collection and
mutation.

## Scope

- Remove the reverse lock ordering in leaderboard metrics paths.
- Keep metrics semantics approximate and operational.
- Add a concurrent regression test for metrics collection during mutation.
- Sync x/cache docs and evidence if the snapshot semantics change.

## Out of Scope

- Changing exported leaderboard APIs.
- Adding strongly consistent metrics snapshots.
- Optimizing score range scans.

## Files

- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard metrics collection cannot deadlock with sorted-set mutation, and
the concurrency behavior is covered by tests.
