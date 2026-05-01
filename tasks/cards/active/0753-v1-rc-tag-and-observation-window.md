# Card 0753

Milestone: v1
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: active
Primary Module: release
Owned Files:
- `docs/release/v1.0.0-rc.1.md`
- `docs/release/`
- `tasks/cards/active/README.md`
- `tasks/cards/done/`
- `specs/extension-beta-evidence.yaml`
Depends On: 0752

Goal:
- Tag `v1.0.0-rc.1`, observe the release candidate, and define the exact path from rc.1 to final `v1.0.0`.

Problem:
Final v1 should not be tagged until the release candidate has passed GitHub gates and any observation-window regressions are classified. The tag also becomes the first real release ref for future extension beta evidence.

Scope:
- Create the annotated `v1.0.0-rc.1` tag only after the evidence package is complete.
- Verify GitHub quality gates on the tagged commit.
- Record observation-window status in release notes.
- If rc.1 is accepted, prepare the final `v1.0.0` checklist.
- If regressions are found, create bounded blocker cards and keep final v1 untagged.

Non-goals:
- Do not tag final `v1.0.0` in the same card.
- Do not promote extensions using only rc.1.
- Do not widen scope beyond release blockers found during observation.

Files:
- `docs/release/v1.0.0-rc.1.md`
- `docs/release/`
- `tasks/cards/active/README.md`
- `tasks/cards/done/`
- `specs/extension-beta-evidence.yaml`

Tests:
- `make gates`
- `git status --short --branch`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
- Required for tag status, observation notes, and future extension evidence refs.

Done Definition:
- `v1.0.0-rc.1` is tagged only from the reviewed release commit.
- GitHub gates are green or blockers are documented with active cards.
- rc.1 is recorded as the first real release ref candidate for later extension evidence.
- The final `v1.0.0` decision is either ready or explicitly blocked.
