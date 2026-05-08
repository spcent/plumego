# Card 0967

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- docs/stable-api/snapshots/middleware-head.snapshot
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md
Depends On:
- 0737-middleware-cors-strict-config-validation

Goal:
Close middleware release-readiness gaps by reconciling the stable API snapshot
and tightening shared conformance coverage for changed wrapper behavior.

Scope:
- Regenerate the middleware stable API snapshot after all code-bearing cards.
- Keep snapshot changes intentional and reviewable.
- Add any remaining shared conformance coverage needed by the preceding fixes.

Non-goals:
- Do not broaden middleware public API beyond changes required by earlier cards.
- Do not snapshot unrelated extension packages.
- Do not run repository-wide refactors.

Files:
- docs/stable-api/snapshots/middleware-head.snapshot
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md

Tests:
- go run ./internal/checks/extension-api-snapshot -module ./middleware/... -out /private/tmp/current-middleware.snapshot
- diff -u docs/stable-api/snapshots/middleware-head.snapshot /private/tmp/current-middleware.snapshot
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Checked-in middleware snapshot matches generated snapshot.
- Shared conformance coverage reflects the final wrapper contract.
- Middleware-wide tests pass.

Outcome:
- Regenerated `docs/stable-api/snapshots/middleware-head.snapshot` from the
  current middleware API surface.
- Verified the checked-in snapshot matches a fresh generated snapshot.
- Confirmed middleware-wide tests and vet pass after the final snapshot update.

Validation:
- `go run ./internal/checks/extension-api-snapshot -module ./middleware/... -out docs/stable-api/snapshots/middleware-head.snapshot`
- `go run ./internal/checks/extension-api-snapshot -module ./middleware/... -out /private/tmp/current-middleware.snapshot`
- `diff -u docs/stable-api/snapshots/middleware-head.snapshot /private/tmp/current-middleware.snapshot`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`
