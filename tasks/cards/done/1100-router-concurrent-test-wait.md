# Card 1100

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: done
Primary Module: router
Owned Files: router/optimization_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0748-router-cache-lookup-api-cleanup

Goal:
Make concurrent serving test coverage deterministic and aligned with the
documented lifecycle contract.

Scope:
- Wait for all concurrent request goroutines in `TestConcurrentSafety`.
- Avoid implying runtime registration during serving is part of the stable
  contract.
- Add status/body assertions that cannot be missed by early test return.
- Mark the active queue empty after this final card completes.

Non-goals:
- Changing production router concurrency behavior.
- Adding dynamic runtime route registration support.
- Repo-wide release tagging.

Files:
- router/optimization_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Only if lifecycle wording needs clarification.

Done Definition:
- Concurrent request goroutines are waited on before the test exits.
- Test intent matches the documented build-before-serve lifecycle.
- Router targeted tests, race tests, and vet pass.
- Active queue is empty.

Outcome:
- Updated `TestConcurrentSafety` to wait for all concurrent request goroutines
  before returning.
- Froze the router before the concurrent serving phase so the test aligns with
  the documented build-before-serve lifecycle.
- Added response body assertions for concurrent serving.
- Marked the active queue empty.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
