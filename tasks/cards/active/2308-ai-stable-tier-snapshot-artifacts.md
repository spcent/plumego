# Card 2308

Milestone:
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: active
Primary Module: x/ai
Owned Files:
- specs/extension-beta-evidence.yaml
- docs/extension-evidence/x-ai-provider.md
- docs/extension-evidence/x-ai-session.md
- docs/extension-evidence/x-ai-streaming.md
- docs/extension-evidence/x-ai-tool.md
Depends On: 2302

Goal:
Add current-head API snapshot artifacts for the stable-tier AI subpackages
without promoting the `x/ai` root family.

Scope:
- Generate snapshot artifacts for `x/ai/provider`, `x/ai/session`,
  `x/ai/streaming`, and `x/ai/tool`.
- Record artifact paths under subpackage candidates in the evidence ledger.
- Keep root `x/ai` experimental and keep blockers until release evidence exists.

Non-goals:
- Do not promote `x/ai`.
- Do not snapshot experimental AI subpackages.
- Do not require live provider credentials.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-ai-provider.md`
- `docs/extension-evidence/x-ai-session.md`
- `docs/extension-evidence/x-ai-streaming.md`
- `docs/extension-evidence/x-ai-tool.md`

Tests:
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/2308-ai-stable-tier-snapshot-artifacts.md`

Docs Sync:
- Required because AI evidence docs change.

Done Definition:
- Stable-tier AI subpackage evidence has current-head snapshot artifacts while
  root promotion remains explicitly out of scope.

Outcome:
