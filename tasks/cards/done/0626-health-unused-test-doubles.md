# Card 0626: Health Unused Test Doubles

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: health
Owned Files:
- `health/mock_test.go`
Depends On: 5401

Goal:
Remove unused health test doubles so the package tests only carry fixtures that
are connected to active assertions.

Scope:
- Delete `mockComponent` and `mockError`, which have no call sites.

Non-goals:
- Do not add new component-check orchestration tests in `health`; orchestration
  belongs to `x/ops/healthhttp`.
- Do not change public runtime code.

Files:
- `health/mock_test.go`

Tests:
- `rg -n "mockComponent|mockError" health`
- `go test -race -timeout 60s ./health/...`
- `go test -timeout 20s ./health/...`
- `go vet ./health/...`

Docs Sync:
No docs change required; this is test hygiene.

Done Definition:
- The unused test helper file is gone.
- No `mockComponent` or `mockError` references remain in `health`.
- The listed validation commands pass.

Outcome:
- Deleted `health/mock_test.go`; both helper types were unused.
- Confirmed `rg -n "mockComponent|mockError" health` returns no matches.
- Validation run:
  - `go test -race -timeout 60s ./health/...`
  - `go test -timeout 20s ./health/...`
  - `go vet ./health/...`
