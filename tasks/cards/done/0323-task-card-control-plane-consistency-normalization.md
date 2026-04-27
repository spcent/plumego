# Card 0323: Task Card Control-Plane Consistency Normalization

Priority: P1
State: done
Primary Module: tasks/cards

## Goal

Make the task-card control plane internally consistent so the physical queue state,
card metadata, and queue README guidance all describe the same execution truth.

## Problem

- `tasks/cards/done/` contains a large number of archived cards whose top-level
  metadata still says `State: active`.
- A representative scan shows `118` archived cards in `tasks/cards/done/` still
  carrying `State: active`, which makes the archive contradict the on-disk queue.
- Some archived cards also append a second `State: done` line under `Outcome`,
  creating two conflicting state declarations in the same file.
- `tasks/cards/active/README.md` has also drifted from the repo-wide read order:
  it starts at `docs/CANONICAL_STYLE_GUIDE.md`, while both `AGENTS.md` and
  `tasks/cards/README.md` start with `docs/CODEX_WORKFLOW.md`.

This is no longer just cosmetic drift. The execution surface used to decide
"what is active now" versus "what is already done" is partially encoded in
directory layout and partially encoded in stale file headers.

## Scope

- Normalize archived card headers in `tasks/cards/done/` so the top-level state
  matches archive reality.
- Remove duplicate or contradictory in-card `State:` lines when they were added
  later under `Outcome`.
- Align `tasks/cards/active/README.md` with the canonical read order used by
  `AGENTS.md` and `tasks/cards/README.md`.
- Tighten `tasks/cards/README.md` if needed so archiving rules explicitly say
  the card header must be updated when a card moves to `done/`.
- Align task-card scaffolding examples and templates with the live numeric card
  naming convention instead of the stale `C-001` / `C-XXXX` pattern.

## Non-Goals

- Do not rewrite historical outcomes beyond state normalization.
- Do not renumber, merge, or re-scope archived cards.
- Do not introduce a new physical `blocked/` or `superseded/` directory.

## Files

- `tasks/cards/README.md`
- `tasks/cards/active/README.md`
- `tasks/cards/TEMPLATE.md`
- `tasks/milestones/PLAN_TEMPLATE.md`
- `Makefile`
- affected files under `tasks/cards/done/*.md`

## Tests

- `rg -n '^State: active$' tasks/cards/done`
- verify no archived card carries more than one `State:` line
- `rg -n '^1\. `docs/CODEX_WORKFLOW\.md`' tasks/cards/README.md tasks/cards/active/README.md`
- `rg -n 'C-XXXX|C-001|active/C-' Makefile tasks/cards/README.md tasks/cards/TEMPLATE.md tasks/milestones/PLAN_TEMPLATE.md`

## Docs Sync

- None outside `tasks/cards/*` unless the cleanup reveals a repo-level workflow
  contradiction.

## Done Definition

- Archived cards in `tasks/cards/done/` no longer claim `State: active` in their
  top-level metadata.
- Archived cards that use the optional `State:` field do not carry a second
  contradictory `State:` line later in the file.
- `tasks/cards/active/README.md` uses the same leading read order as
  `AGENTS.md` and `tasks/cards/README.md`.
- Task-card scaffolding no longer advertises the stale `C-` filename/header
  convention.

## Outcome

- Normalized archived card headers so `tasks/cards/done/` no longer carries
  stale top-level `State: active` metadata from finished work.
- Removed duplicate `State: done` lines that had been appended under `Outcome`
  in a subset of archived cards, leaving one canonical top-level state field.
- Updated `tasks/cards/active/README.md` to restore the canonical read order
  starting with `docs/CODEX_WORKFLOW.md`.
- Tightened `tasks/cards/README.md` so the archive rule explicitly requires
  updating the header state when a card moves to `done/`.
- Updated the task-card template, milestone plan template, and `make new-card`
  / `make check-card` examples so scaffolding now matches the live numeric card
  naming scheme used in `tasks/cards/active/` and `tasks/cards/done/`.

## Validation Run

```bash
rg -n '^State: active$' tasks/cards/done
sh -c 'for f in tasks/cards/done/*.md; do c=$(grep -c "^State: " "$f"); if [ "$c" -gt 1 ]; then echo "$f:$c"; fi; done'
rg -n '^1\. `docs/CODEX_WORKFLOW\.md`' tasks/cards/README.md tasks/cards/active/README.md
rg -n 'C-XXXX|C-001|active/C-' Makefile tasks/cards/README.md tasks/cards/TEMPLATE.md tasks/milestones/PLAN_TEMPLATE.md
```
