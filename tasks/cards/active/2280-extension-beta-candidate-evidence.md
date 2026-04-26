# Card 2280

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: docs
Owned Files:
- specs/extension-beta-evidence.yaml
- docs/extension-evidence/x-rest.md
- docs/extension-evidence/x-websocket.md
- docs/extension-evidence/x-tenant.md
- docs/extension-evidence/x-observability.md
- docs/extension-evidence/x-gateway.md
Depends On: 2279

Goal:
Seed beta promotion evidence records for the first three current candidate extension families.

Scope:
- Create one concise evidence note for each candidate: `x/rest`, `x/websocket`, and `x/tenant`.
- Record current test coverage, primer status, release-history blocker, owner sign-off blocker, and API snapshot status.
- Link each evidence note from the machine-readable evidence file.
- Keep the outcome as evidence gathering only; all three modules remain `experimental`.

Non-goals:
- Do not promote any candidate to `beta`.
- Do not create speculative release history.
- Do not cover `x/observability` or `x/gateway`; those follow in card 2289.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-rest.md`
- `docs/extension-evidence/x-websocket.md`
- `docs/extension-evidence/x-tenant.md`

Tests:
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `scripts/check-spec tasks/cards/done/2280-extension-beta-candidate-evidence.md`

Docs Sync:
- Required because beta blocker evidence becomes user-visible.

Done Definition:
- `x/rest`, `x/websocket`, and `x/tenant` have evidence records with the same fields and blocker shape.
- The evidence records make clear that release-history proof is still missing.
- No module manifest status is changed.

Outcome:
