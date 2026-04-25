# Card 2143: Task Card Archive State Normalization

Milestone: none
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: done
Primary Module: tasks/cards
Owned Files:
- `tasks/cards/done`
- `tasks/cards/active/README.md`
- `tasks/cards/README.md`
Depends On: none

Goal:
Normalize task-card metadata so archived cards do not continue to advertise
`State: active` after they have moved to `tasks/cards/done`.

Problem:
Several archived cards in `tasks/cards/done/` still contain `State: active`
while also having completed `Outcome` and `Validation` sections. This weakens
the task queue as an execution control surface: agents that search state fields
can misclassify completed work as active, and humans have to rely on directory
location instead of metadata.

Scope:
- Search `tasks/cards/done` for `State: active`.
- Change only archived cards that clearly have completed outcome/validation
  sections to `State: done`.
- Leave intentionally blocked, superseded, or incomplete cards untouched and
  list them in the outcome.
- Add or update a lightweight check only if an existing task-card check already
  owns this rule; otherwise keep this as a metadata cleanup card.

Non-goals:
- Do not rewrite card content beyond the `State` field.
- Do not move cards between active and done.
- Do not change milestone specs.
- Do not introduce a broad new task workflow.

Files:
- `tasks/cards/done/*`
- `tasks/cards/active/README.md`
- `tasks/cards/README.md`

Tests:
- `rg -n "^State: active$" tasks/cards/done`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`

Docs Sync:
Update the task-card README only if the active/done state rule is not already
explicit.

Done Definition:
- Completed cards under `tasks/cards/done` no longer say `State: active`.
- Any exceptions are documented in the card outcome.
- The listed validation commands pass.

Outcome:
- Normalized all archived cards under `tasks/cards/done` that still declared
  `State: active`; no archived card now advertises itself as active.
- No blocked or superseded exceptions were found in `tasks/cards/done`.
- No README update was needed because `tasks/cards/README.md` already states
  that cards moved to `done/` must update the top-level state to `done`.

Validation:
- `rg -n "^State: active$" tasks/cards/done` returned no matches.
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
