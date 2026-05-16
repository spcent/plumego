# Card 1446

Milestone: M-007
Recipe: specs/change-recipes/docs-only.yaml
Priority: P1
State: active
Primary Module: x/ai
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-ai-provider.md`
- `docs/extension-evidence/x-ai-session.md`
- `docs/extension-evidence/x-ai-streaming.md`
- `docs/extension-evidence/x-ai-tool.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/extension-evidence/snapshots/`
- `tasks/cards/active/README.md`
Depends On:
- 1445

Goal:
- Record `v1.0.0` as the first release-ref intake point for `x/ai` stable-tier
  subpackages without changing their promotion state.

Scope:
- Add the `v1.0.0` tag target commit as the first release ref for
  `provider`, `session`, `streaming`, and `tool`.
- Generate or register v1 baseline snapshot artifacts in evidence docs.
- Keep beta blockers explicit until the two-release rule and owner sign-off are
  complete.

Non-goals:
- Do not promote `x/ai` or any subpackage.
- Do not change AI runtime behavior.
- Do not add owner sign-off.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-ai-provider.md`
- `docs/extension-evidence/x-ai-session.md`
- `docs/extension-evidence/x-ai-streaming.md`
- `docs/extension-evidence/x-ai-tool.md`
- `docs/extension-evidence/release-artifacts.md`
- `docs/extension-evidence/snapshots/`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`
- snapshot parse or compare checks for generated artifacts

Docs Sync:
- Required because the evidence state changes.

Done Definition:
- Each stable-tier subpackage has a recorded first release-ref intake point.
- Remaining blockers are still validated by checks.
- Card is moved to done with validation output.

Outcome:
-
