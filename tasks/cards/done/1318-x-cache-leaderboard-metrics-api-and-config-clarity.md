# 1318 - x/cache leaderboard metrics API and config clarity

Status: done
Priority: P1
Primary module: `x/cache`

## Problem

`LeaderboardMetrics` is exported but includes an unexported mutex, making it a
poor stable DTO. `EnableMetrics` is a bool defaulting to true in the default
config, but a custom config literal gets the zero value false, which is easy to
misread.

## Scope

- Separate internal metrics counters from the exported metrics snapshot shape.
- Keep `GetLeaderboardMetrics` returning the exported snapshot type.
- Document `EnableMetrics` custom-config semantics clearly.
- Add tests proving metrics can be disabled intentionally.
- Refresh current-head leaderboard API snapshot if the exported shape changes.

## Out of Scope

- Strongly consistent metrics snapshots.
- Removing metrics support.
- Promoting leaderboard to stable.

## Files

- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`
- `docs/extension-evidence/snapshots/x-cache/x-cache-leaderboard-head.snapshot`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard metrics use a clean exported snapshot type, internal locking is not
part of the public DTO, and metrics enable/disable behavior is explicit.

## Outcome

- Split internal metrics locking/counters into an unexported type while keeping
  `LeaderboardMetrics` as the exported snapshot DTO.
- Documented that `EnableMetrics` defaults to true through
  `DefaultLeaderboardConfig`; custom config literals keep the zero value and
  intentionally disable operation counters.
- Added disabled-metrics behavior coverage and refreshed the leaderboard
  current-head API snapshot.
- Validation passed:
  - `go test -race -timeout 60s ./x/cache/leaderboard`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/leaderboard -out docs/extension-evidence/snapshots/x-cache/x-cache-leaderboard-head.snapshot`
  - `go test -timeout 20s ./x/cache/...`
  - `go vet ./x/cache/...`
