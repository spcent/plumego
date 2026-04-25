# Card 2208: Frontend Error Message Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend_test.go`
Depends On: none

Goal:
Consolidate repeated frontend error message containment assertions.

Problem:
`frontend_test.go` repeats nil checks and `strings.Contains(err.Error(), ...)`
for invalid filesystem and directory inputs. The checks verify the same contract
but each test spells out the boilerplate independently.

Scope:
- Add a local helper for asserting an error mentions an expected fragment.
- Use it for straightforward frontend constructor and registration error tests.

Non-goals:
- Do not change frontend error construction, routing, or serving behavior.
- Do not rewrite tests that intentionally accept OS-specific alternative error
  strings.
- Do not add dependencies.

Files:
- `x/frontend/frontend_test.go`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- Straightforward frontend error message assertions use a named helper.
- The listed validation commands pass.

Outcome:
- Added `assertErrorContains` for frontend invalid input error checks.
- Validation passed for frontend race tests, normal tests, and vet.
