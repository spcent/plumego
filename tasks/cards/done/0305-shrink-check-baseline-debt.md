# Card 0305

Priority: P2

Goal:
- Review and reduce the current migration-debt baselines under `specs/check-baseline/*` instead of letting them silently persist.

Scope:
- baseline inventory
- debt documentation
- removal candidates

Non-goals:
- Do not suppress new failures by expanding baselines in this card.
- Do not rewrite checks.

Files:
- `specs/check-baseline/dependency-rules.txt`
- `specs/check-baseline/missing-module-manifests.txt`
- `specs/check-baseline/reference-layout-legacy-roots.txt`
- `docs/ROADMAP.md`
- `AGENTS.md`

Tests:
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- Keep baseline policy aligned with roadmap and agent guidance.

Done Definition:
- Each baseline file has a current reason to exist.
- At least one removal or reduction candidate is identified or executed.
- The repository keeps treating baselines as temporary migration debt.
