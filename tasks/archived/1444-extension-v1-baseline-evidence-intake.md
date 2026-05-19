# Card 1444

Milestone: M-007
Recipe: specs/change-recipes/docs-only.yaml
Priority: P1
State: done
Primary Module: extension evidence
Owned Files:
- `tasks/milestones/done/M-007-extension-v1-baseline-evidence-intake/M-007.md`
- `tasks/milestones/done/M-007-extension-v1-baseline-evidence-intake/plan-M-007.md`
- `tasks/cards/active/README.md`
- `tasks/cards/active/1445-x-tenant-v1-baseline-evidence.md`
- `tasks/cards/active/1446-x-ai-stable-tier-v1-baseline-evidence.md`
- `tasks/cards/active/1447-data-discovery-messaging-v1-baseline-gap-index.md`
- `tasks/cards/active/1448-post-v1-extension-evidence-queue-finalize.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/release/POST_V1_EVIDENCE.md`
Depends On:
- M-006

Goal:
- Create the M-007 execution lane for post-v1 extension evidence intake.

Scope:
- Add the M-007 milestone and plan.
- Add executable cards 1445 through 1448.
- Refresh release evidence guidance so agents start from the published
  `v1.0.0` tag state instead of the stale no-release-ref state.

Non-goals:
- Do not change runtime behavior.
- Do not promote extensions.
- Do not add or modify owner sign-off.

Files:
- `tasks/milestones/done/M-007-extension-v1-baseline-evidence-intake/M-007.md`
- `tasks/milestones/done/M-007-extension-v1-baseline-evidence-intake/plan-M-007.md`
- `tasks/cards/`
- `docs/extension-evidence/release-artifacts.md`
- `docs/release/POST_V1_EVIDENCE.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/agent-workflow`
- `git diff --check`

Docs Sync:
- Required because release evidence routing changed after `v1.0.0`.

Done Definition:
- M-007 exists with sequential cards.
- The active queue starts with card 1445.
- Release evidence docs list `v1.0.0` as the first post-v1 baseline input.

Outcome:
- Added the M-007 milestone and plan.
- Added cards 1445 through 1448 to the active queue.
- Updated release evidence guidance without changing extension maturity.
