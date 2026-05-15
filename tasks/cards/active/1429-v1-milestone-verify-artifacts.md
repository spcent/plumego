# Card 1429

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: release
Owned Files:
- `tasks/milestones/M-001.verify.md`
- `tasks/milestones/M-002.verify.md`
- `tasks/milestones/M-003.verify.md`
- `tasks/milestones/STATUS.md`
- `tasks/milestones/ROADMAP.md`
Depends On:
- none

Goal:
- Add the missing milestone verify artifacts and reconcile milestone status so
  release execution starts from a trustworthy control plane.

Scope:
- Record current evidence for M-001, M-002, and M-003.
- Update milestone status and roadmap only with repository facts.
- Keep M-004 completion visible without changing runtime code.

Non-goals:
- Do not create or push release tags.
- Do not change Go runtime behavior.
- Do not promote extension status.

Files:
- `tasks/milestones/M-001.verify.md`
- `tasks/milestones/M-002.verify.md`
- `tasks/milestones/M-003.verify.md`
- `tasks/milestones/STATUS.md`
- `tasks/milestones/ROADMAP.md`

Tests:
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `git status --short --branch`

Docs Sync:
- Required for milestone status and release control-plane truth.

Done Definition:
- Verify artifacts exist for M-001, M-002, and M-003.
- Status and roadmap describe the current milestone sequence accurately.
- Control-plane checks pass.

Outcome:
- Created `tasks/milestones/M-001.verify.md`,
  `tasks/milestones/M-002.verify.md`, and
  `tasks/milestones/M-003.verify.md`.
- Updated `tasks/milestones/STATUS.md` and `tasks/milestones/ROADMAP.md` to
  show that verify artifacts now exist and remaining release work moves to
  M-005 cards.
- Validation passed:
  - `go run ./internal/checks/agent-workflow`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/extension-maturity`
  - `go run ./internal/checks/extension-beta-evidence`
  - `go run ./internal/checks/dependency-rules`
  - `go run ./internal/checks/deprecation-inventory -strict`
  - `go test -timeout 20s ./contract ./core ./router ./middleware/... ./security/... ./store/... ./health ./log ./metrics`
  - `cd cmd/plumego && go test -timeout 20s ./...`
  - `cd cmd/plumego && go run . new --template canonical --dry-run trust-check`
  - `cd cmd/plumego && go run . new --template rest-api --dry-run trust-check`
  - `cd cmd/plumego && go run . new --template invalid-template-name --dry-run trust-check` failed with exit status 3 as expected.
