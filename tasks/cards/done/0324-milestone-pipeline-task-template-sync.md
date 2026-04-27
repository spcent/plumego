# Card 0324: Milestone Pipeline Task Template Sync

Priority: P2
State: done
Primary Module: docs

## Goal

Bring the milestone pipeline documentation back in sync with the current
task-card template and scaffolding so milestone planning guidance describes the
same card shape the repository actually generates and archives.

## Problem

- `docs/MILESTONE_PIPELINE.md` still shows the card header as
  `# Card C-XXX`, even though the live task-card template and queue now use the
  numeric form `# Card XXXX`.
- The same document's card field list omits `Recipe`, while
  `tasks/cards/TEMPLATE.md` and the task-card control plane treat `Recipe` as a
  first-class field in the scaffolded shape.
- This leaves milestone-oriented documentation one cleanup step behind the
  actual task-card control plane:
  - milestone authors read one card format in `docs/MILESTONE_PIPELINE.md`
  - scaffolding generates another format from `tasks/cards/TEMPLATE.md`
  - queue docs and `Makefile` now already follow the live numeric format

## Scope

- Update `docs/MILESTONE_PIPELINE.md` so its task-card examples and field lists
  match the current `tasks/cards/TEMPLATE.md`.
- Keep milestone-oriented examples aligned with the current numeric task-card
  naming convention.
- Mention `Recipe` wherever the milestone pipeline describes the canonical card
  metadata shape.

## Non-Goals

- Do not redesign the milestone workflow itself.
- Do not change milestone file naming (`M-NNN`) or branch conventions.
- Do not reopen task-card queue rules that were normalized in `0954`.

## Files

- `docs/MILESTONE_PIPELINE.md`
- `tasks/cards/TEMPLATE.md`
- optionally `tasks/milestones/README.md` if wording there needs a matching note

## Tests

- `rg -n 'C-XXX|C-001|# Card C-' docs/MILESTONE_PIPELINE.md`
- verify the card field list in `docs/MILESTONE_PIPELINE.md` includes `Recipe`
- verify the example card header matches `tasks/cards/TEMPLATE.md`

## Docs Sync

- Only sync milestone/task-control-plane docs touched by this mismatch.

## Done Definition

- Milestone pipeline docs show the same task-card header format as the live
  template.
- Milestone pipeline docs list `Recipe` as part of the canonical task-card
  metadata shape.
- No milestone-facing example still teaches the stale `C-` task-card format.

## Outcome

- Updated `docs/MILESTONE_PIPELINE.md` so the milestone pipeline now teaches the
  live numeric task-card header form `# Card XXXX` instead of the stale
  `# Card C-XXX` example.
- Added `Recipe` to the milestone pipeline's required task-card field list and
  example card block, bringing it back in line with
  `tasks/cards/TEMPLATE.md` and the current scaffolding shape.
- No further milestone README changes were needed because the mismatch was local
  to the pipeline document.

## Validation Run

```bash
rg -n 'C-XXX|C-001|# Card C-' docs/MILESTONE_PIPELINE.md
rg -n '`Recipe`|Recipe: specs/change-recipes/<recipe>.yaml|# Card XXXX' docs/MILESTONE_PIPELINE.md
```
