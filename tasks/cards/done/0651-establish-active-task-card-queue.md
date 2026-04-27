# Card 0651

Priority: P1

Goal:
- Establish the active task-card queue and lifecycle rules.

Scope:
- active queue layout
- state model
- archival rules

Non-goals:
- Do not migrate the full roadmap in one card.
- Do not introduce a complex workflow engine for cards.

Files:
- `tasks/cards/README.md`
- `tasks/cards/active/README.md`
- `docs/ROADMAP.md`

Tests:
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Keep active queue rules aligned with roadmap execution expectations.

Done Definition:
- Plumego has a documented live execution surface under `tasks/cards/active/`.
- Card lifecycle expectations are clear enough for agents and maintainers to use consistently.
- New near-term roadmap work can be added to the active queue without ad hoc conventions.

Outcome:
- `tasks/cards/README.md` now defines the lifecycle vocabulary for `active`, `blocked`, `done`, and `superseded`.
- `tasks/cards/active/README.md` now defines how the live queue should be maintained while only `active/` and `done/` exist physically.
- The roadmap documentation now records the active queue as a real execution surface rather than an aspirational one.

Validation Run:
- `go run ./internal/checks/agent-workflow`
