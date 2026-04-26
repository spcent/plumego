# Card 2304

Milestone:
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: active
Primary Module: docs
Owned Files:
- specs/extension-beta-evidence.yaml
- docs/extension-evidence/x-tenant.md
- docs/extension-evidence/x-observability.md
- docs/extension-evidence/x-gateway.md
- docs/extension-evidence/snapshots/first-batch/
Depends On: 2303

Goal:
Add checked-in current-head API snapshot artifacts for the remaining first-batch
beta candidates without claiming release-history completion.

Scope:
- Generate current-head exported API snapshots for `x/tenant`,
  `x/observability`, and `x/gateway`.
- Record artifact paths in the evidence ledger.
- Preserve blockers until two real release refs and owner sign-off exist.

Non-goals:
- Do not promote modules.
- Do not add release refs unless they are real release refs.
- Do not include AI subpackage or second-batch surfaces in this card.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-tenant.md`
- `docs/extension-evidence/x-observability.md`
- `docs/extension-evidence/x-gateway.md`
- `docs/extension-evidence/snapshots/first-batch/`

Tests:
- `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-tenant-head.snapshot docs/extension-evidence/snapshots/first-batch/x-tenant-head.snapshot`
- `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-observability-head.snapshot docs/extension-evidence/snapshots/first-batch/x-observability-head.snapshot`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
- Required because beta evidence docs and artifacts change.

Done Definition:
- All first-batch beta candidates have checked-in current-head snapshot artifacts
  referenced by the ledger, while promotion blockers remain accurate.

Outcome:
