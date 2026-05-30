# Plan for M-009: Beta Promotions Round 1

Milestone: `M-009`
Objective: Promote four extension surfaces to beta status using the two-release
evidence now available after v1.1.0: x/ai stable-tier subpackages (provider,
session, tool, streaming), x/tenant, x/gateway/discovery core-static, and
x/data/file plus x/data/idempotency. Update module.yaml status fields, generate
release-to-release API snapshots, record owner sign-off, and reflect all changes
in docs/concepts/extension-maturity.md.
Constraints: two release refs required (v1.0.0 + v1.1.0), no runtime behavior
changes, no stable-root public API additions, owner sign-off must be explicit.
Affected Modules: extension evidence, x/ai, x/tenant, x/gateway/discovery, x/data.

## Phase Map

- Phase 1: Orient — confirm v1.1.0 tag and second_release_ref populated before any edits.
- Phase 2: Parallel Promotions — run all four promotion cards concurrently since they
  touch different modules.
- Phase 3: Dashboard and Validate — update EXTENSION_MATURITY.md and DEPRECATION.md,
  run acceptance criteria, commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1367 | Promote x/tenant to beta | x/tenant | `specs/extension-beta-evidence.yaml`, `docs/evidence/extension/x-tenant.md`, `x/tenant/module.yaml`, `docs/concepts/extension-maturity.md` | M-008 | `extension-beta-evidence`, `extension-maturity` |
| 1370 | Promote x/ai provider, session, tool, streaming to beta | x/ai | `specs/extension-beta-evidence.yaml`, `docs/evidence/extension/x-ai-provider.md`, `docs/evidence/extension/x-ai-session.md`, `docs/evidence/extension/x-ai-tool.md`, `docs/evidence/extension/x-ai-streaming.md`, `x/ai/provider/module.yaml`, `x/ai/session/module.yaml`, `x/ai/tool/module.yaml`, `x/ai/streaming/module.yaml` | M-008 | `extension-beta-evidence`, `extension-maturity` |
| 1371 | Promote x/data/file and x/data/idempotency to beta | x/data | `specs/extension-beta-evidence.yaml`, `docs/evidence/extension/x-data.md`, `x/data/file/module.yaml`, `x/data/idempotency/module.yaml` | M-008 | `extension-beta-evidence`, `extension-maturity` |
| 1372 | Promote x/gateway/discovery core-static to beta | x/gateway | `specs/extension-beta-evidence.yaml`, `docs/evidence/extension/x-discovery.md`, `x/gateway/discovery/module.yaml` | M-008 | `extension-beta-evidence`, `extension-maturity` |

## Dependency Edges

- All four cards depend on M-008 (second_release_ref must exist).
- Cards 1367, 1370, 1371, 1372 are independent of each other.

## Parallel Groups

- Group A (parallel): cards 1367, 1370, 1371, 1372 — different modules, no file overlap.
- Group B (sequential after A): dashboard update (EXTENSION_MATURITY.md, DEPRECATION.md).

## Risk Register

- Risk: API snapshot comparison reveals an unintended symbol change in one surface.
  Mitigation: block that surface's promotion; record explicit blocker in evidence ledger
  and continue promoting other surfaces.
- Risk: owner sign-off record is ambiguous or attributed to the wrong team.
  Mitigation: each card requires an explicit owner_signoff field in the evidence doc
  naming the team; do not infer from prior history.

## Verification Strategy

- Card-level checks: run `go run ./internal/checks/extension-beta-evidence` and
  `go run ./internal/checks/extension-maturity` after each card completes.
- Milestone-level checks: run the full Acceptance Criteria suite after all four cards
  complete and before committing.
- Maturity check: confirm promoted surfaces appear at beta tier in extension-maturity output.

## Exit Condition

- all four promotion cards completed
- x/ai provider, session, tool, streaming; x/tenant; x/gateway/discovery; x/data/file
  and x/data/idempotency all show status = beta in their module.yaml files
- docs/concepts/extension-maturity.md reflects all promotions
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
