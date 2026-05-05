# 0764 - x/cache leaderboard randomness and performance evidence

Status: active
Priority: P2
Primary module: `x/cache`

## Problem

Leaderboard skiplist level selection uses package-level `math/rand`, and the
range operations intentionally scan the base level. Stable readiness needs
clearer local implementation ownership plus checked-in performance evidence for
the bounded in-process contract.

## Scope

- Move skiplist random-level state behind the skiplist instance.
- Preserve existing sorted-set ordering behavior.
- Add a deterministic test seam for skiplist level generation if needed.
- Record the existing benchmark commands and scale boundary in docs/evidence.
- Keep range operations bounded by documented `MaxMembersPerSet`.

## Out of Scope

- Replacing skiplist range scans with indexed score-range counters.
- Changing exported leaderboard APIs.
- Claiming Redis-scale range performance.

## Files

- `x/cache/leaderboard/skiplist.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Skiplist randomness is instance-owned and leaderboard docs/evidence record the
bounded performance contract without implying Redis-scale behavior.
