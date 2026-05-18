# Card 1515

Milestone: M-010
Recipe: specs/change-recipes/update-docs.yaml
Priority: P1
State: active
Primary Module: extension evidence
Owned Files:
- `docs/EXTENSION_MATURITY.md`
- `docs/DEPRECATION.md`

Goal:
- Update docs/EXTENSION_MATURITY.md to reflect the x/messaging promotion to
  beta and the x/frontend evaluation result from M-010.
- Update docs/DEPRECATION.md beta extension table with the newly promoted entry.

Scope:
- Change x/messaging status in EXTENSION_MATURITY.md from experimental to beta.
- Record x/frontend result (beta or explicit blocker) in EXTENSION_MATURITY.md.
- Add x/messaging to the beta extension table in DEPRECATION.md.
- Run go run ./internal/checks/extension-maturity to confirm dashboard aligns
  with module.yaml status fields.

Non-goals:
- Do not change module.yaml status fields in this card (changed in cards 1373
  and 1514).
- Do not document x/resilience or other surfaces not evaluated in M-010.

Files:
- `docs/EXTENSION_MATURITY.md`
- `docs/DEPRECATION.md`

Tests:
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
- This card is itself a docs-sync card; no further sync required.

Done Definition:
- docs/EXTENSION_MATURITY.md shows beta for x/messaging and a resolved entry
  for x/frontend.
- docs/DEPRECATION.md beta extension table is current.
- `go run ./internal/checks/extension-maturity` exits 0.

Outcome:
-
