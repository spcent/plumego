# Card 0005

Priority: P2

Goal:
- Make `tasks/cards/` part of the normal agent workflow instead of an undocumented side directory.

Scope:
- task-card discovery in repo metadata and top-level guidance
- minimal task-card conventions

Non-goals:
- Do not build a task runner.
- Do not add automation or scheduling logic.

Files:
- `specs/repo.yaml`
- `docs/ROADMAP.md`
- `AGENTS.md`
- `CLAUDE.md`
- `tasks/cards/README.md`

Tests:
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Ensure roadmap, repo spec, and agent guides all mention `tasks/cards/`.

Done Definition:
- Agents can discover task cards from canonical repo metadata.
- The repository documents what a task card is and how to use it.
- `tasks/cards/` becomes a durable workflow surface, not a temporary scratch area.
