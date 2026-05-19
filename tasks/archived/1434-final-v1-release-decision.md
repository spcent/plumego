# Card 1434

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: release
Owned Files:
- `docs/release/v1.0.0.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/M-005.verify.md`
- `tasks/cards/active/README.md`
Depends On:
- 1433

Goal:
- Decide whether to tag final `v1.0.0` or open bounded blocker cards for the
  next release candidate.

Scope:
- Apply the pre-v1 release checklist to the rc evidence package.
- Tag final `v1.0.0` only if no P0/P1 blockers remain.
- If blocked, create or reference concrete rc follow-up cards.

Non-goals:
- Do not skip the rc observation window.
- Do not tag final v1 with missing gate evidence.
- Do not hide extension blockers by changing labels.

Files:
- `docs/release/v1.0.0.md`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/milestones/M-005.verify.md`
- `tasks/cards/active/README.md`

Tests:
- `make gates`
- `go run ./internal/checks/extension-beta-evidence`
- `git status --short --branch`

Docs Sync:
- Required for final release notes or no-go blocker summary.

Done Definition:
- Final decision is GO or NO-GO with evidence.
- `v1.0.0` is tagged only on GO.
- NO-GO creates bounded blocker cards before the milestone stops.

Outcome:
- Applied the pre-v1 release checklist to current rc evidence.
- Validation passed:
  - `make gates`
  - `go run ./internal/checks/extension-beta-evidence`
  - `git status --short --branch`
- Remote GitHub Actions run `25920615874` for `v1.0.0-rc.1` passed.
- Final decision: NO-GO for final `v1.0.0` on May 15, 2026 because the rc
  observation window has not completed.
- Created card 1436 to own the observation-window result and final v1 handoff.
- Final `v1.0.0` was not tagged.
