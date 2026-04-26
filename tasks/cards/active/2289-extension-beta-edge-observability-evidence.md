# Card 2289

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: docs
Owned Files:
- specs/extension-beta-evidence.yaml
- docs/extension-evidence/x-observability.md
- docs/extension-evidence/x-gateway.md
- docs/EXTENSION_STABILITY_POLICY.md
Depends On: 2280

Goal:
Seed beta promotion evidence records for the observability and gateway candidate extension families.

Scope:
- Create concise evidence notes for `x/observability` and `x/gateway`.
- Record current test coverage, primer status, release-history blocker, owner sign-off blocker, and API snapshot status.
- Link both records from the machine-readable evidence file.
- Keep both modules `experimental`.

Non-goals:
- Do not promote `x/observability` or `x/gateway`.
- Do not add new observability or gateway behavior.
- Do not create speculative release history.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-observability.md`
- `docs/extension-evidence/x-gateway.md`
- `docs/EXTENSION_STABILITY_POLICY.md`

Tests:
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `scripts/check-spec tasks/cards/done/2289-extension-beta-edge-observability-evidence.md`

Docs Sync:
- Required because beta blocker evidence becomes user-visible.

Done Definition:
- `x/observability` and `x/gateway` have evidence records matching the first candidate set.
- The records make clear that release-history proof and owner sign-off are still missing.
- No module manifest status is changed.

Outcome:
