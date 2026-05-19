# Card 1445

Milestone: M-007
Recipe: specs/change-recipes/docs-only.yaml
Priority: P1
State: done
Primary Module: x/tenant
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-tenant.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/extension-evidence/snapshots/`
- `tasks/cards/active/README.md`
Depends On:
- 1444

Goal:
- Record `v1.0.0` as the first release-ref intake point for `x/tenant` beta
  evidence without promoting `x/tenant`.

Scope:
- Add the `v1.0.0` tag target commit as the first `x/tenant` release ref.
- Generate or register a v1 baseline snapshot artifact in evidence docs.
- Keep beta blockers explicit until there are two release refs, enough
  release-backed snapshots, and owner sign-off.

Non-goals:
- Do not promote `x/tenant`.
- Do not change tenant runtime behavior.
- Do not add owner sign-off.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-tenant.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/extension-evidence/snapshots/`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`
- snapshot parse or compare check for the generated artifact

Docs Sync:
- Required because the evidence state changes.

Done Definition:
- `x/tenant` has a recorded first release-ref intake point.
- Remaining blockers are still validated by checks.
- Card is moved to done with validation output.

Outcome:
- Added `v1.0.0` tag target
  `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` as the first `x/tenant`
  release-ref intake point.
- Generated unchanged `v1.0.0` to `v1.0.0` snapshot artifacts under
  `docs/extension-evidence/snapshots/v1-baseline/x-tenant/`.
- Kept `release_history_missing`, `api_snapshot_missing`, and
  `owner_signoff_missing`; `x/tenant` remains experimental.
