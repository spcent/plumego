# Card 1078

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: done
Primary Module: stable roots
Owned Files:
- `docs/stable-api/`
- `specs/deprecation-inventory.yaml`
- `specs/dependency-rules.yaml`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/cards/active/README.md`
Depends On: 0746

Goal:
- Complete the final v1 API freeze audit for stable roots.

Problem:
The v1 contract is the stable-root surface, not the entire repository. Before tagging, the repository needs explicit evidence that `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, and `metrics` have a frozen, narrow, dependency-safe public surface.

Scope:
- Review each stable root `module.yaml` for GA status, owner, risk, and non-goals.
- Refresh or verify stable API snapshots for all stable roots.
- Confirm stable roots do not import `x/*`.
- Search stable roots for stale `deprecated`, `legacy`, `not implemented`, `TODO`, and compatibility aliases.
- Register every intentionally retained stable compatibility symbol in `specs/deprecation-inventory.yaml`.

Non-goals:
- Do not change API behavior unless the audit finds a P0 release blocker.
- Do not refactor stable package internals opportunistically.
- Do not add new stable roots.

Files:
- `docs/stable-api/`
- `specs/deprecation-inventory.yaml`
- `specs/dependency-rules.yaml`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/dependency-rules`
- `go test ./core ./router ./contract ./middleware ./security/... ./store/... ./health ./log ./metrics`

Docs Sync:
- Required if snapshots, compatibility decisions, or release notes change.

Done Definition:
- All stable roots have current freeze evidence.
- No stable root imports `x/*`.
- Every retained stable compatibility item has an explicit inventory decision.
- Any required symbol removal is split into a dedicated symbol-change card before implementation.

Outcome:
- Verified stable root manifests for `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, and `metrics` are `status: ga`.
- Verified checked-in stable API snapshots match the current exported API surface.
- Verified stable root production code has no `x/*` imports.
- Confirmed retained stable compatibility aliases are registered in `specs/deprecation-inventory.yaml`.
- Updated `docs/stable-api/README.md` and `docs/release/v1.0.0-rc.1.md` with the audit result.
- Validation passed:
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/dependency-rules`
  - `go test ./core ./router ./contract ./middleware ./security/... ./store/... ./health ./log ./metrics`
