# 0971 - x/cache leaderboard contract consistency

Status: done
Priority: P1
Primary module: `x/cache`

## Problem

Leaderboard behavior still has unclear stable semantics: failed first writes can
leave empty leaderboards, missing-key behavior differs by operation, and metrics
do not distinguish attempted removals from actual removals.

## Scope

- Prevent failed first writes from leaving empty leaderboards.
- Document and test the chosen missing leaderboard behavior.
- Count actual removed members rather than requested removals.
- Add regression tests for failed create cleanup, missing-key semantics, and
  actual removal metrics.

## Out of Scope

- Redis sorted-set compatibility expansion.
- Skiplist algorithm rewrite.
- Distributed leaderboard support.

## Files

- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `docs/modules/x-cache/README.md`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard failure cleanup, missing-key behavior, and mutation metrics are
explicit, tested, and documented.

## Outcome

- Prevented failed first writes from leaving empty leaderboards.
- Documented and tested missing-key behavior across sorted-set operations.
- Changed `ZRems` to count actual removed members.
- Added regression coverage for failed create cleanup and actual removal
  metrics.

## Validation Run

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
