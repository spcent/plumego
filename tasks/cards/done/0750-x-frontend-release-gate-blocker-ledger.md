# Card 0750: x/frontend Release Gate Blocker Ledger

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: done
Primary Module: x/frontend
Owned Files:
- `docs/extension-evidence/x-frontend.md`
- `docs/EXTENSION_MATURITY.md`
- `docs/modules/x-frontend/README.md`
- `x/frontend/module.yaml`
Depends On: 0749

Goal:
Record the verified remaining stable blockers and release-gate expectations for
`x/frontend`.

Scope:
- Keep release history, release-backed API snapshots, owner sign-off, and full
  release gate evidence as remaining stable blockers.
- Record that current-head snapshots and targeted module tests are not release
  evidence.
- Verify the previously suspected `x/mq` single-test blocker before mentioning
  it as a current blocker.
- Keep `x/frontend` status `experimental`.

Non-goals:
- Do not fix unrelated non-frontend gate failures.
- Do not invent release refs or owner sign-off.
- Do not promote `x/frontend`.

Files:
- `docs/extension-evidence/x-frontend.md`
- `docs/EXTENSION_MATURITY.md`
- `docs/modules/x-frontend/README.md`
- `x/frontend/module.yaml`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go test -timeout 20s ./x/mq -run TestKVDeduperLifecycle -count=1`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
This is a release readiness evidence card.

Done Definition:
- Remaining stable blockers are listed as unresolved and actionable.
- `x/mq TestKVDeduperLifecycle` is only recorded if it currently fails.
- No status promotion occurs.
- The listed validation commands pass.

Outcome:
- Documented that `x/frontend` stable promotion still needs release history,
  release-backed API snapshots, owner sign-off, and a passing repository release
  gate from the candidate release state.
- Rechecked the previously suspected `x/mq TestKVDeduperLifecycle` blocker; it
  currently passes and is not recorded as an active `x/frontend` blocker.
- Preserved `experimental` status and did not invent release refs or sign-off.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go test -timeout 20s ./x/mq -run TestKVDeduperLifecycle -count=1`
  - `go run ./internal/checks/extension-beta-evidence`
