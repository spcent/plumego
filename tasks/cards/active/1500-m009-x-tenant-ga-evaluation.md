# Card 1500

Milestone: M-009
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: active
Primary Module: x/tenant
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-tenant.md`
- `docs/EXTENSION_MATURITY.md`
- `x/tenant/module.yaml`

Goal:
- Evaluate whether x/tenant qualifies for GA promotion based on release-history
  evidence from v1.0.0 → v1.1.0.
- If promotion criteria are met, update manifests, evidence, and maturity dashboard.
- If criteria are not yet met, record explicit blockers and the unblock conditions.

Problem:
x/tenant was promoted to beta at v1.1.0. GA promotion requires:
1. Two consecutive minor releases without exported API changes (v1.1.x → v1.2.x).
2. Checked-in API snapshots for both release refs in `docs/extension-evidence/snapshots/`.
3. Owner sign-off recorded in `specs/extension-beta-evidence.yaml`.
4. No open integration test failures across resolve → policy → quota → ratelimit chain.
5. Updated `docs/EXTENSION_MATURITY.md` and `x/tenant/module.yaml` (`status: ga`).

Scope:
- Run `go run ./internal/checks/extension-release-evidence -module ./x/tenant -base v1.1.0 -head v1.2.0`
  once v1.2.0 exists; inspect output for API changes.
- Compare exported symbols against the checked-in snapshot.
- If no API changes: complete evidence ledger, update manifests, update roadmap.
- If API changes detected: record them as blockers with the specific symbols.

Non-goals:
- Do not change x/tenant runtime behavior from this card.
- Do not move tenant concerns into stable roots.
- Do not promote any x/tenant subordinate surface (session, store adapters) separately
  unless their own evidence records are complete.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-tenant.md`
- `docs/EXTENSION_MATURITY.md`
- `x/tenant/module.yaml`
- `docs/ROADMAP.md`

Tests:
- `go test -race -timeout 60s ./x/tenant/...`
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Required: update `docs/EXTENSION_MATURITY.md`, `x/tenant/module.yaml` status,
  `docs/ROADMAP.md` Phase 9, and `docs/modules/x/tenant/README.md` v-status table.

Done Definition:
- `x/tenant` has `status: ga` in its module.yaml, OR
- The card is moved to blocked with explicit API-change blockers and unblock conditions.

Notes:
- x/tenant integration test in `x/tenant/integration_test.go` already covers the
  full resolve → policy → quota → ratelimit chain; confirm it passes before evaluation.
- The promotion checklist is at `docs/EXTENSION_STABILITY_POLICY.md`.
- Owner sign-off must come from the `multitenancy` owner group.
