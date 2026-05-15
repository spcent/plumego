# Card 1430

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: blocked
Primary Module: release
Owned Files:
- `docs/release/v1.0.0-rc.1.md`
- `tasks/cards/active/1375-v1-rc-tag-and-observation-window.md`
- `tasks/cards/active/README.md`
Depends On:
- 1429

Goal:
- Create and verify `v1.0.0-rc.1` as the release candidate evidence anchor.

Scope:
- Create the annotated rc tag from the reviewed release commit.
- Run local release gates and record the result.
- Push tag/branch and record GitHub gate status.

Non-goals:
- Do not tag final `v1.0.0`.
- Do not promote extension status.
- Do not change runtime behavior unless a release blocker card is opened.

Files:
- `docs/release/v1.0.0-rc.1.md`
- `tasks/cards/active/1375-v1-rc-tag-and-observation-window.md`
- `tasks/cards/active/README.md`

Tests:
- `make gates`
- `go run ./internal/checks/extension-beta-evidence`
- `git tag -l 'v1.0.0-rc.1'`

Docs Sync:
- Required for release notes and observation-window status.

Done Definition:
- `v1.0.0-rc.1` exists locally and remotely.
- Local and GitHub gates are recorded.
- Final v1 is either ready for observation or explicitly blocked.

Outcome:
-
