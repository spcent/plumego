# 0853 - x/cache leaderboard contract hardening

Status: done
Priority: P1
Primary module: `x/cache`

## Problem

`x/cache/leaderboard` has useful ranked-data behavior, but stable readiness is
blocked by ignored contexts, inconsistent key validation, nil member panics, and
skiplist/map divergence for duplicate or rollback paths.

## Scope

- Honor canceled contexts before leaderboard operations mutate or read state.
- Reuse cache key validation semantics for leaderboard keys.
- Reject nil `ZMember` values without panics.
- Fix duplicate member insertion for same-score updates.
- Preserve skiplist/map consistency when invalid score updates fail.
- Add focused negative and regression tests.

## Out of Scope

- Redis-compatible sorted-set API expansion.
- Distributed leaderboard behavior.
- Performance rewrites beyond correctness fixes.

## Files

- `x/cache/leaderboard/leaderboard.go`
- `x/cache/leaderboard/skiplist.go`
- `x/cache/leaderboard/leaderboard_test.go`
- `x/cache/leaderboard/test_helpers_test.go`

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`

## Done Definition

Leaderboard operations fail predictably on invalid inputs and canceled contexts,
and duplicate/update paths leave cardinality, range, and score state consistent.

## Outcome

- Reused stable `store/cache` validation for leaderboard keys and canceled
  contexts.
- Added explicit invalid member handling for nil and empty sorted-set members.
- Fixed same-score member updates so skiplist state does not duplicate entries.
- Preserved skiplist/map consistency when score increments overflow.
- Added focused regression tests for each hardened behavior.

## Validation

- `go test -race -timeout 60s ./x/cache/leaderboard`
