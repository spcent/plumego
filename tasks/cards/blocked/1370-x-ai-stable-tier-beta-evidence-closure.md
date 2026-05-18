# Card 1370

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: blocked
Primary Module: x/ai
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-ai-provider.md`
- `docs/extension-evidence/x-ai-session.md`
- `docs/extension-evidence/x-ai-streaming.md`
- `docs/extension-evidence/x-ai-tool.md`
Depends On: second release refs and owner sign-off

Goal:
- Complete beta evidence closure for `x/ai` stable-tier subpackages.
- This card is part of the post-v1 maturity roadmap recorded in `tasks/cards/active/1452-post-v1-maturity-roadmap.md`.

Problem:
The evidence ledger tracks `x/ai/provider`, `x/ai/session`, `x/ai/streaming`,
and `x/ai/tool` with evidence docs, current-head snapshots, first `v1.0.0`
release refs, and v1 baseline snapshot artifacts, but each remains blocked by
missing second release refs, complete release-backed snapshots, and owner
sign-off.

Scope:
- Add the second real release ref only after the next qualifying tag or release
  commit exists.
- Generate release-to-release API snapshots for each stable-tier subpackage.
- Record owner sign-off from `ai-gateway`.
- Keep blockers until all evidence is present.

Non-goals:
- Do not promote root `x/ai`.
- Do not use `HEAD` as release-history evidence.
- Do not change AI runtime behavior.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-ai-provider.md`
- `docs/extension-evidence/x-ai-session.md`
- `docs/extension-evidence/x-ai-streaming.md`
- `docs/extension-evidence/x-ai-tool.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Required when evidence is added.

Done Definition:
- All tracked `x/ai` stable-tier subpackages have two release refs, matching release snapshots, and owner sign-off, or blockers remain explicit.

Outcome:
-
