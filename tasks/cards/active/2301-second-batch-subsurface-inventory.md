# Card 2301

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P2
State: active
Primary Module: docs
Owned Files:
- docs/extension-evidence/x-data.md
- docs/extension-evidence/x-messaging.md
- docs/extension-evidence/x-discovery.md
- docs/EXTENSION_MATURITY.md
- tasks/cards/active/README.md
Depends On: 2295

Goal:
Split second-batch maturity into sub-surface inventories so future promotion
work is not root-package sized.

Scope:
- Add candidate surface tables for `x/data`, `x/messaging`, and `x/discovery`.
- Identify likely beta candidates and explicitly experimental surfaces.
- Update dashboard blockers to point at sub-surface inventory work.

Non-goals:
- Do not promote modules.
- Do not add tests or code behavior.
- Do not evaluate subordinate primitives as independent beta candidates yet.

Files:
- `docs/extension-evidence/x-data.md`
- `docs/extension-evidence/x-messaging.md`
- `docs/extension-evidence/x-discovery.md`
- `docs/EXTENSION_MATURITY.md`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/2301-second-batch-subsurface-inventory.md`

Docs Sync:
- Required because maturity guidance changes.

Done Definition:
- Second-batch evidence notes identify smaller candidate surfaces for future
  beta evaluation.

Outcome:
