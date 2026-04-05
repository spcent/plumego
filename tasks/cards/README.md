# Task Cards

`tasks/cards/` is Plumego's repo-native execution queue for reversible work items.

Directory layout:

- `tasks/cards/active/` — current working queue
- `tasks/cards/done/` — completed card archive

Lifecycle states:

- `active` — ready to execute now and expected to be actionable without more decomposition
- `blocked` — still relevant, but waiting on a prerequisite, decision, or earlier card
- `done` — completed and validated; move to `tasks/cards/done/`
- `superseded` — no longer the right execution unit because another card replaced or absorbed it

Current operating rule:

- `active` and `done` are the implemented physical directories
- until dedicated `blocked/` or `superseded/` directories are introduced, non-actionable cards must not quietly remain mixed into the live queue
- if a card becomes blocked or superseded, update its title or body to say so immediately and either:
  - keep it in `active/` only while it is still part of the immediate queue
  - or replace it with a clearer active card and archive the obsolete one once the queue has been cleaned up

## Purpose

Use task cards when work should be:

- small enough for one focused implementation pass
- scoped to one primary module when possible
- easy to validate and easy to revert in one commit

Task cards are workflow assets, not archival prose. Keep long-form planning in `docs/ROADMAP.md` and machine-readable rules in `specs/`.

## Card Format

Each card should define:

- Goal
- Scope
- Non-goals
- Files
- Tests
- Docs Sync
- Done Definition

Optional operational fields when useful:

- Priority
- State
- Blocked By
- Supersedes
- Outcome
- Validation Run

## Authoring Rules

- prefer one primary module
- keep the file set small
- keep validation short and relevant
- avoid mixing unrelated runtime, docs, and architecture changes in one card
- keep the live queue short enough that the next card is obvious
- archive completed cards promptly so `active/` remains a working queue rather than a history dump

## Queue Management

- `active/` is ordered by execution intent, not by historical id alone
- the first card in the queue should be the best next action, not merely the oldest open card
- when a card finishes, move it to `done/` and record the actual outcome and validations that ran
- when a card is no longer the best execution unit, replace or retire it explicitly instead of leaving stale work in the queue
- do not keep roadmap-scale planning items in `active/`; split them before queueing them

## Read Order

When a task card applies, read it after:

1. `docs/CANONICAL_STYLE_GUIDE.md`
2. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
3. `specs/repo.yaml`
4. `specs/task-routing.yaml`

Then use the card to drive the concrete change sequence.

## Relationship to Other Repo Surfaces

- `docs/` explains the architecture and roadmap
- `specs/` defines machine-readable rules and recipes
- `tasks/cards/` turns roadmap work into executable cards

In Plumego, the roadmap names the work, but the active queue decides what gets executed next.
