# Card 2294

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P1
State: active
Primary Module: x/ai
Owned Files:
- specs/extension-beta-evidence.yaml
- internal/checks/extension-beta-evidence/main.go
- docs/modules/x-ai/README.md
- tasks/cards/active/README.md
Depends On: 2287, 2290

Goal:
Advance `x/ai` stable-tier subpackage evidence without promoting the root
`x/ai` family.

Scope:
- Add evidence-ledger entries for `x/ai/provider`, `x/ai/session`,
  `x/ai/streaming`, and `x/ai/tool` as subpackage candidates.
- Teach `extension-beta-evidence` to validate stable-tier subpackage candidates
  against `x/ai/module.yaml`.
- Keep root `x/ai` status experimental.

Non-goals:
- Do not promote root `x/ai`.
- Do not evaluate experimental AI subpackages.
- Do not add live provider tests or credentials.

Files:
- `specs/extension-beta-evidence.yaml`
- `internal/checks/extension-beta-evidence/main.go`
- `docs/modules/x-ai/README.md`
- `tasks/cards/active/README.md`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/2294-x-ai-stable-tier-evidence-ledger.md`

Docs Sync:
- Required because AI subpackage evidence tracking changes.

Done Definition:
- AI stable-tier subpackages are tracked in the evidence ledger.
- The evidence check validates those subpackages without requiring root `x/ai`
  promotion.

Outcome:
