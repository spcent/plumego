# Card 1429

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: active
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
-
