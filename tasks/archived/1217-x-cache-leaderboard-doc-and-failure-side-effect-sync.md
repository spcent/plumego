# 1217 - x/cache leaderboard doc and failure side-effect sync

Status: done
Priority: P2
State: done
Primary module: `x/cache`

## Problem

Leaderboard package documentation is split between `leaderboard.go` and
`doc.go`, and failed `ZIncrBy` operations may rebuild skiplist placement while
preserving logical member state. Stable readiness needs package docs and failure
side-effect semantics to be aligned.

## Scope

- Sync `x/cache/leaderboard/doc.go` with the Plumego-local non-Redis contract.
- Document `ZIncrBy` invalid-result rollback as logical state preservation,
  not a no-side-effect structural guarantee.
- Add regression coverage that invalid `ZIncrBy` preserves member score and
  cardinality.
- Update module docs and evidence.

## Out of Scope

- Replacing random skiplist levels.
- Adding Redis compatibility.
- Changing exported leaderboard API.

## Files

- `x/cache/leaderboard/doc.go`
- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard package docs are consistent and invalid-score rollback semantics are
documented and covered by tests.

## Outcome

- Synced `x/cache/leaderboard/doc.go` with the Plumego-local non-Redis
  compatibility contract.
- Documented invalid-result `ZIncrBy` rollback as logical state preservation,
  not a structural no-op guarantee.
- Extended invalid increment regression coverage to score, cardinality, rank,
  and range output.
- Updated x/cache module docs and evidence.

## Validation Run

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
