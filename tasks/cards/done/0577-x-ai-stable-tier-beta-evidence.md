# Card 0577

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P2
State: active
Primary Module: x/ai
Owned Files:
- docs/modules/x-ai/README.md
- docs/extension-evidence/x-ai-provider.md
- docs/extension-evidence/x-ai-session.md
- docs/extension-evidence/x-ai-streaming.md
- docs/extension-evidence/x-ai-tool.md
Depends On: 2280

Goal:
Prepare subpackage-level beta evidence records for the `x/ai` stable-tier path without promoting the root family.

Scope:
- Treat `provider`, `session`, `streaming`, and `tool` as separate beta-evaluation surfaces.
- Record test coverage, examples, API snapshot status, and remaining release-history blockers for each subpackage.
- Update the `x/ai` primer to link to the subpackage evidence.
- Keep experimental AI subpackages out of the stable-tier evidence path.

Non-goals:
- Do not promote root `x/ai` to `beta`.
- Do not evaluate `orchestration`, `semanticcache`, `marketplace`, `distributed`, or `resilience` in this card.
- Do not add provider globals or live-provider tests.

Files:
- `docs/modules/x-ai/README.md`
- `docs/extension-evidence/x-ai-provider.md`
- `docs/extension-evidence/x-ai-session.md`
- `docs/extension-evidence/x-ai-streaming.md`
- `docs/extension-evidence/x-ai-tool.md`

Tests:
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go vet ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `scripts/check-spec tasks/cards/done/0577-x-ai-stable-tier-beta-evidence.md`

Docs Sync:
- Required because subpackage stability guidance changes.

Done Definition:
- Each stable-tier `x/ai` subpackage has its own beta evidence record.
- The root `x/ai` family remains experimental in docs and manifest.
- Experimental AI subpackages are explicitly excluded from this evidence path.

Outcome:
- Added subpackage-level beta evidence records for `x/ai/provider`,
  `x/ai/session`, `x/ai/streaming`, and `x/ai/tool`.
- Updated the `x/ai` primer to link those evidence records and state that the
  root `x/ai` family remains experimental.
- Explicitly excluded experimental AI subpackages from the stable-tier evidence
  path.

Validations:
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go vet ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `scripts/check-spec tasks/cards/done/0577-x-ai-stable-tier-beta-evidence.md`
