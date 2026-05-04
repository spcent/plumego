# 0744 - x/cache leaderboard contract and efficiency

Status: active
Priority: P1
Primary module: `x/cache`

## Problem

Leaderboard validation borrows `MemoryCache.Exists` for key checks, missing-key
behavior is split across commands, and some internal counters/scans are noisy or
inefficient. Stable readiness needs clearer local ranked-data semantics.

## Scope

- Replace validation side effects with direct context/key validation.
- Remove unused cleanup state.
- Track leaderboard count without scanning all leaderboards for every create.
- Document the deliberate missing-key behavior split or normalize where safe.
- Add regression tests for validation, cleanup, limits, and missing-key
  semantics.

## Out of Scope

- Promising Redis sorted-set compatibility.
- Replacing the skip list implementation.
- Public API shape changes.

## Files

- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard key validation is explicit, create-limit accounting avoids full
map scans, and the command contract is documented and covered by tests.

## Outcome

Pending.
