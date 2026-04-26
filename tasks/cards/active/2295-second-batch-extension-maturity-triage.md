# Card 2295

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P2
State: active
Primary Module: docs
Owned Files:
- docs/EXTENSION_MATURITY.md
- docs/extension-evidence/x-messaging.md
- docs/extension-evidence/x-data.md
- docs/extension-evidence/x-discovery.md
- tasks/cards/active/README.md
Depends On: 2291, 2294

Goal:
Start second-batch extension family maturity evaluation after the first beta
candidate pipeline is in place.

Scope:
- Add initial maturity evidence notes for `x/messaging`, `x/data`, and
  `x/discovery`.
- Update the dashboard with concrete blockers and recommended next evidence.
- Keep all three families experimental.

Non-goals:
- Do not promote modules.
- Do not add broad tests in this card.
- Do not evaluate subordinate primitives independently from their parent family.

Files:
- `docs/EXTENSION_MATURITY.md`
- `docs/extension-evidence/x-messaging.md`
- `docs/extension-evidence/x-data.md`
- `docs/extension-evidence/x-discovery.md`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/2295-second-batch-extension-maturity-triage.md`

Docs Sync:
- Required because maturity guidance changes.

Done Definition:
- Second-batch families have explicit evidence notes and dashboard blockers.
- No module status changes are made.

Outcome:
