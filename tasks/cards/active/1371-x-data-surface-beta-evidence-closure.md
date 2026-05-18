# Card 1371

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: active
Primary Module: x/data
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-data.md`
- `docs/extension-evidence/release-artifacts.md`
Depends On: release-backed snapshots and owner sign-off

Goal:
- Complete beta evidence closure for selected `x/data` surfaces.
- This card is part of the post-v1 maturity roadmap recorded in `tasks/cards/active/1452-post-v1-maturity-roadmap.md`.

Problem:
The evidence ledger tracks `x/data/file` and `x/data/idempotency` with first
`v1.0.0` release refs and checked-in baseline snapshots, but both surfaces still
lack second release refs, complete release-backed snapshots, and owner sign-off.

Scope:
- Add the second real release ref only after the next qualifying tag or release
  commit exists.
- Generate release-to-release API snapshots with `extension-release-evidence`.
- Record owner sign-off from `persistence`.

Non-goals:
- Do not promote root `x/data`.
- Do not include topology-heavy `x/data` surfaces by default.
- Do not change data runtime behavior.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-data.md`
- `docs/extension-evidence/release-artifacts.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Required when evidence is added.

Done Definition:
- Selected `x/data` surfaces have checked-in snapshots, two release refs, release snapshots, and owner sign-off, or blockers remain explicit.

Outcome:
-
