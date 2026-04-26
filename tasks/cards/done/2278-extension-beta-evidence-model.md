# Card 2278

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: specs
Owned Files:
- specs/extension-beta-evidence.yaml
- specs/checks.yaml
- docs/EXTENSION_STABILITY_POLICY.md
- docs/ROADMAP.md
Depends On:

Goal:
Define the machine-readable evidence model used to decide whether an `x/*` module can move from `experimental` to `beta`.

Scope:
- Add a repo-native evidence file that lists beta candidates, required release refs, exported API snapshot refs, owner sign-off state, and current blockers.
- Update the stability policy to point promotion work at the evidence file instead of relying on prose-only claims.
- Register the evidence file in the checks/control-plane documentation where appropriate.
- Keep the model narrow enough to support `x/rest`, `x/websocket`, `x/tenant`, `x/observability`, and `x/gateway` first.

Non-goals:
- Do not promote any module to `beta`.
- Do not implement exported API diffing in this card.
- Do not add release automation or network-dependent checks.

Files:
- `specs/extension-beta-evidence.yaml`
- `specs/checks.yaml`
- `docs/EXTENSION_STABILITY_POLICY.md`
- `docs/ROADMAP.md`

Tests:
- `go run ./internal/checks/agent-workflow`
- `scripts/check-spec tasks/cards/done/2278-extension-beta-evidence-model.md`

Docs Sync:
- Required because beta promotion workflow changes from prose-only to evidence-backed.

Done Definition:
- The repo has a canonical evidence model for extension beta promotion.
- Candidate modules and blocker fields can be read without opening individual module primers.
- No module status changes are made.

Outcome:
- Added `specs/extension-beta-evidence.yaml` as the canonical evidence ledger
  for extension beta promotion.
- Seeded the five current beta candidates with manifest-aligned owners,
  missing release refs, missing API snapshots, missing owner sign-off, and
  explicit blockers.
- Updated the extension stability policy and roadmap so promotion work points
  at the evidence ledger before any `module.yaml` status changes.
- Registered the evidence file as a control-plane file in `specs/checks.yaml`.

Validations:
- `go run ./internal/checks/agent-workflow`
