# Card 1448

Milestone: M-007
Recipe: specs/change-recipes/docs-only.yaml
Priority: P1
State: active
Primary Module: tasks
Owned Files:
- `tasks/cards/active/README.md`
- `tasks/cards/active/`
- `tasks/cards/done/`
- `tasks/milestones/M-007.verify.md`
- `docs/release/POST_V1_EVIDENCE.md`
Depends On:
- 1447

Goal:
- Reconcile the active queue after post-v1 extension evidence intake.

Scope:
- Move completed M-007 cards to done.
- Add the M-007 verify artifact.
- Update the active queue so the next blocker is explicit and no completed work
  remains at the front.
- Update the post-v1 evidence index with the M-007 outcome.

Non-goals:
- Do not promote extensions.
- Do not change runtime behavior.
- Do not rewrite older verify artifacts.

Files:
- `tasks/cards/active/README.md`
- `tasks/cards/active/`
- `tasks/cards/done/`
- `tasks/milestones/M-007.verify.md`
- `docs/release/POST_V1_EVIDENCE.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/agent-workflow`
- `git status --short --branch`

Docs Sync:
- Required because queue and release evidence routing changes.

Done Definition:
- M-007 verify artifact exists.
- Active queue reflects only remaining actionable or explicitly blocked work.
- Extension evidence blockers remain accurate.

Outcome:
-
