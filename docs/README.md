# Documentation Navigation (Canonical Priority)

Last updated: 2026-03-10

Use this page as the entrypoint for `docs/` to avoid referencing historical plans by mistake.

## Read Order (Authoritative)

1. `docs/other/V1_CORE_API_FREEZE.md`  
   v1 API freeze list, stability tiers, and compatibility boundaries.
2. `docs/CANONICAL_STYLE_GUIDE.md`  
   Canonical coding/documentation style and handler/middleware conventions.
3. `docs/getting-started.md`  
   Quick canonical wiring for app/router/middleware usage.
4. `docs/modules/README.md`  
   Module-by-module documentation index.

If guidance conflicts, this priority order wins.

## Canonical vs Historical

- Canonical v1 guidance:
  - `docs/other/V1_CORE_API_FREEZE.md`
  - `docs/CANONICAL_STYLE_GUIDE.md`
  - `docs/getting-started.md`
  - `docs/modules/*` (unless explicitly marked experimental)
- Historical / archived (do not treat as current policy):
  - `docs/CORE_API_CONTRACTION_PROPOSAL.md`
  - `docs/CORE_REFACTOR_PLAN.md`
  - `docs/other/V1.0_READINESS_REPORT.md` (superseded by newer execution/scope docs)
  - `docs/tenant/*` planning/checklist docs (see `docs/tenant/README.md`)
  - `docs/lowcode/*` phase-design drafts (not part of v1 canonical docs)

## Experimental Modules

Per v1 scope decision (`docs/other/V1_CORE_API_FREEZE.md`), `tenant/*` and `net/mq/*` are experimental in v1.0 and are not covered by GA stability guarantees.
