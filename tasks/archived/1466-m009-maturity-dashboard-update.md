# Card 1513

Milestone: M-009
Recipe: specs/change-recipes/update-docs.yaml
Priority: P1
State: done
Primary Module: extension evidence
Owned Files:
- `docs/EXTENSION_MATURITY.md`
- `docs/DEPRECATION.md`

Goal:
- Update docs/EXTENSION_MATURITY.md to reflect all four surfaces promoted to
  beta in M-009 (x/ai stable-tier, x/tenant, x/gateway/discovery, x/data/file
  and x/data/idempotency).
- Update docs/DEPRECATION.md beta extension table with the newly promoted entries.

Scope:
- Change status column for each promoted surface in EXTENSION_MATURITY.md from
  experimental to beta.
- Add promotion date and first two release refs for each entry.
- Add each promoted surface to the beta extension table in DEPRECATION.md.
- Run go run ./internal/checks/extension-maturity to confirm dashboard aligns
  with module.yaml status fields.

Non-goals:
- Do not promote root x/ai module.
- Do not change any module.yaml status fields in this card (those are changed
  in cards 1367, 1370, 1371, 1372).
- Do not document planned future promotions as current.

Files:
- `docs/EXTENSION_MATURITY.md`
- `docs/DEPRECATION.md`

Tests:
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
- This card is itself a docs-sync card; no further sync required.

Done Definition:
- docs/EXTENSION_MATURITY.md shows beta for all four Round 1 promoted surfaces.
- docs/DEPRECATION.md beta extension table is current.
- `go run ./internal/checks/extension-maturity` exits 0.

Outcome:
- Done with one explicit exception: 1372 remained blocked because
  `x/gateway/discovery` did not exist at `v1.0.0`, so no beta promotion was
  recorded for that surface. Updated docs/EXTENSION_MATURITY.md and
  docs/DEPRECATION.md for the completed Round 1 promotions: `x/tenant`, the
  four `x/ai` stable-tier surfaces, and `x/data/file` plus
  `x/data/idempotency`.
