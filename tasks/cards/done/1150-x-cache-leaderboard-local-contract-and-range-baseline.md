# 1150 - x/cache leaderboard local contract and range baseline

Status: done
Priority: P1
State: done
Primary module: `x/cache`

## Problem

Leaderboard missing-key behavior is intentionally Plumego-local rather than
Redis sorted-set compatible, range/count operations remain linear over level 0,
and metrics snapshots are approximate. Stable readiness needs these choices to
be explicit and backed by tests/docs rather than inferred from implementation.

## Scope

- Add package-level documentation for Plumego-local missing-key semantics.
- Add tests that lock down missing-key behavior across sorted-set methods.
- Add explicit documentation for approximate leaderboard metrics snapshots.
- Record range/count linear-scan behavior and size guidance in module docs.
- Add a focused benchmark or regression note that preserves current complexity
  expectations without promising Redis-scale behavior.

## Out of Scope

- Replacing the skiplist implementation.
- Full Redis sorted-set compatibility.
- Distributed leaderboard storage.

## Files

- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `x/cache/leaderboard/leaderboard_bench_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard local contracts, missing-key behavior, range complexity, and
approximate metrics semantics are explicit, tested, and documented.

## Outcome

- Added package-level documentation that leaderboard is Plumego-local ranked
  data, not Redis sorted-set compatibility.
- Locked down `ZRangeByScore` missing-key behavior alongside the existing
  missing-key contract tests.
- Documented approximate metrics snapshot semantics in code and module docs.
- Documented score range operations as skiplist base-level scans and added a
  full-range benchmark baseline.

## Validation Run

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
