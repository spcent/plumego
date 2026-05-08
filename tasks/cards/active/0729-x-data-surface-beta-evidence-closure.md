# Card 0729

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: blocked
Primary Module: x/data
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-data.md`
- `docs/extension-evidence/release-artifacts.md`
Depends On: release refs and owner sign-off

Goal:
- Complete beta evidence closure for selected `x/data` surfaces.

Problem:
The evidence ledger tracks `x/data/file` and `x/data/idempotency`, but both surfaces currently lack checked-in API snapshots, release history, release snapshots, and owner sign-off.

Scope:
- Generate current-head snapshots for `x/data/file` and `x/data/idempotency` when the surfaces are ready for candidate review.
- Add two real release refs only after tags or release commits exist.
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
