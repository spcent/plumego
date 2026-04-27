# Card 0029

Priority: P1

Goal:
- Audit stable `store` for topology-heavy placement debt and define the next migration steps for distributed cache and provider-heavy adapters.

Scope:
- stable `store` placement review
- baseline and roadmap alignment

Non-goals:
- Do not move packages in this card.
- Do not change runtime store behavior in this card.

Files:
- `store/module.yaml`
- `docs/modules/store/README.md`
- `docs/ROADMAP.md`
- `specs/repo.yaml`
- `specs/check-baseline/*`

Tests:
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Keep store guidance aligned with the architecture blueprint and roadmap.

Done Definition:
- The repo explicitly states what remains acceptable in stable `store`.
- The next migration targets are documented without ambiguity.
- No new placement drift is normalized by silence.
