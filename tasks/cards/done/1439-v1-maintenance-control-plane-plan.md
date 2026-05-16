# Card 1439

Milestone: M-006
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: release
Owned Files:
- `tasks/milestones/active/M-006.md`
- `tasks/milestones/M-006.plan.md`
- `tasks/cards/active/README.md`
- `tasks/milestones/ROADMAP.md`
- `tasks/milestones/STATUS.md`
Depends On:
- M-005

Goal:
- Write the v1.0.1 maintenance control plane into `tasks/`.

Scope:
- Create the M-006 milestone and plan.
- Add bounded follow-up cards for generated data, CI warning cleanup, CLI
  onboarding truth, and post-v1 evidence indexing.
- Front-load the active queue with the next executable maintenance card.

Non-goals:
- Do not change runtime behavior.
- Do not rewrite the published `v1.0.0` tag.
- Do not promote experimental extensions.

Files:
- `tasks/milestones/active/M-006.md`
- `tasks/milestones/M-006.plan.md`
- `tasks/cards/active/README.md`
- `tasks/milestones/ROADMAP.md`
- `tasks/milestones/STATUS.md`

Tests:
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `git status --short --branch`

Docs Sync:
- Required for task control-plane state only.

Done Definition:
- M-006 exists.
- M-006 has a card inventory and dependency order.
- Active queue starts with card 1440.

Outcome:
- Added M-006 as the post-v1 maintenance lane.
- Added cards 1440 through 1443.
- Left extension evidence cards blocked until complete evidence exists.
