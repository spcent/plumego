# Card 1447

Milestone: M-007
Recipe: specs/change-recipes/docs-only.yaml
Priority: P1
State: active
Primary Module: extension evidence
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-data.md`
- `docs/extension-evidence/x-discovery.md`
- `docs/extension-evidence/x-messaging.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/extension-evidence/snapshots/`
- `tasks/cards/active/README.md`
Depends On:
- 1446

Goal:
- Add missing v1 baseline snapshot evidence for selected `x/data`,
  `x/discovery`, and `x/messaging` surfaces while preserving promotion
  blockers.

Scope:
- Generate checked-in baseline API snapshots for `x/data/file`,
  `x/data/idempotency`, `x/discovery`, and `x/messaging`.
- Record `v1.0.0` as the first release ref when the surface has a usable
  snapshot.
- Keep owner sign-off and second-release blockers explicit.

Non-goals:
- Do not promote any selected surface.
- Do not widen selected data surfaces.
- Do not change runtime behavior.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-data.md`
- `docs/extension-evidence/x-discovery.md`
- `docs/extension-evidence/x-messaging.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/extension-evidence/snapshots/`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`
- snapshot parse or compare checks for generated artifacts

Docs Sync:
- Required because evidence artifacts are added.

Done Definition:
- Selected surfaces have checked-in baseline snapshots or an explicit blocker
  explaining why not.
- Remaining blockers are still validated by checks.
- Card is moved to done with validation output.

Outcome:
-
