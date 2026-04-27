# Card 0588

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P1
State: done
Primary Module: x/ai
Owned Files:
- docs/extension-evidence/x-ai-provider.md
- docs/extension-evidence/x-ai-session.md
- docs/extension-evidence/x-ai-streaming.md
- docs/extension-evidence/x-ai-tool.md
- docs/modules/x-ai/README.md
Depends On: 2294, 2296

Goal:
Advance AI stable-tier subpackage evidence with release snapshot guidance while
keeping root `x/ai` experimental.

Scope:
- Add release-aware snapshot commands for provider, session, streaming, and tool.
- Re-state experimental exclusions for root `x/ai`.
- Keep all blockers until real release refs and owner sign-off exist.

Non-goals:
- Do not evaluate experimental AI subpackages.
- Do not add live provider tests.
- Do not promote root `x/ai`.

Files:
- `docs/extension-evidence/x-ai-provider.md`
- `docs/extension-evidence/x-ai-session.md`
- `docs/extension-evidence/x-ai-streaming.md`
- `docs/extension-evidence/x-ai-tool.md`
- `docs/modules/x-ai/README.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `scripts/check-spec tasks/cards/done/0588-ai-stable-tier-snapshot-guidance.md`

Docs Sync:
- Required because AI evidence guidance changes.

Done Definition:
- AI stable-tier evidence docs reference the release-aware API snapshot workflow.
- Root `x/ai` remains explicitly excluded from promotion.

Outcome:
- Added release-aware comparison workflow commands to the `x/ai/provider`,
  `x/ai/session`, `x/ai/streaming`, and `x/ai/tool` evidence docs.
- Updated the `x/ai` primer to state that release comparison commands apply
  only to the named stable-tier subpackage, not root-family promotion.
- Kept all blockers in place until real release refs, snapshot files, and owner
  sign-off exist.

Validations:
- `go run ./internal/checks/extension-beta-evidence`
- `go test -timeout 20s ./x/ai/provider ./x/ai/session ./x/ai/streaming ./x/ai/tool`
- `scripts/check-spec tasks/cards/done/0588-ai-stable-tier-snapshot-guidance.md`
