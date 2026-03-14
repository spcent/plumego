# Card 0303

Priority: P1

Goal:
- Tighten the rule that tenant-aware behavior must not leak back into stable `middleware` or `store`.

Scope:
- boundary guidance
- dependency and architecture wording
- baseline review for tenant-related drift

Non-goals:
- Do not redesign `x/tenant`.
- Do not move working tenant code in this card.

Files:
- `specs/dependency-rules.yaml`
- `AGENTS.md`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `docs/modules/store/README.md`
- `docs/modules/middleware/README.md`

Tests:
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Keep tenant boundary rules aligned across specs and primers.

Done Definition:
- Stable-root docs and boundary rules make tenant leakage unmistakably invalid.
- Future tenant drift is easier to detect in review and checks.
