# Card 0885

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/fuzz_test.go, docs/modules/router/README.md, router/module.yaml, tasks/cards/active/README.md
Depends On: 0730-router-cache-internal-api-cleanup

Goal:
Add lightweight release-readiness regression coverage for router path and
reverse-routing invariants.

Scope:
- Add seed-based fuzz tests that run under normal `go test`.
- Cover path normalization no-panic behavior and URL reverse-routing parity.
- Document these checks as part of router stable readiness.
- Mark the active queue empty when complete.

Non-goals:
- Long-running fuzz campaigns.
- Benchmark threshold enforcement.
- Repo-wide release tagging.

Files:
- router/fuzz_test.go
- docs/modules/router/README.md
- router/module.yaml
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- New fuzz seeds run in normal router tests.
- Router stable docs reference the release-readiness coverage.
- Router tests, race tests, and vet pass.

Outcome:
- Added seed-based fuzz coverage for route path helper normalization and named
  reverse routing URL generation.
- Documented the new release-readiness fuzz coverage in router module docs.
- Updated router manifest risk wording to include path normalization drift.
- Marked the active queue empty.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
- Extra: `go run ./internal/checks/module-manifests`
