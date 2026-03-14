# Task Cards

`tasks/cards/` is Plumego's repo-native execution queue for reversible work items.

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

## Authoring Rules

- prefer one primary module
- keep the file set small
- keep validation short and relevant
- avoid mixing unrelated runtime, docs, and architecture changes in one card

## Read Order

When a task card applies, read it after:

1. `docs/CANONICAL_STYLE_GUIDE.md`
2. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
3. `specs/repo.yaml`
4. `specs/agent-entrypoints.yaml`

Then use the card to drive the concrete change sequence.

## Relationship to Other Repo Surfaces

- `docs/` explains the architecture and roadmap
- `specs/` defines machine-readable rules and recipes
- `tasks/cards/` turns roadmap work into executable cards
