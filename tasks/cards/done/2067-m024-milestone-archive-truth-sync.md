# Card 2067

Milestone: M-024
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: control-plane
Priority: P2
State: done
Primary Module: tasks
Owned Files:
- `tasks/milestones/ROADMAP.md`
- `tasks/milestones/STATUS.md`
- `tasks/milestones/superseded/README.md`
- `tasks/milestones/ARCHIVE_INDEX.md`
Depends On: none

## Goal

Reconcile milestone archive truth with the files actually present on disk and
make superseded-draft history explicit instead of implicit.

## Scope

Explain which historical milestone ids have archived directories, which are
represented only in roadmap/history prose today, and how superseded draft files
relate to canonical milestone directories.

## Non-goals

- Do not fabricate historical milestone directories that do not exist.
- Do not rewrite historical milestone goals or outcomes.
- Do not widen this card into new roadmap prioritization work.

## Files

- `tasks/milestones/ROADMAP.md`
- `tasks/milestones/STATUS.md`
- `tasks/milestones/superseded/README.md`
- `tasks/milestones/ARCHIVE_INDEX.md`

## Acceptance Tests

<!-- none; task-history truthfulness card -->

## Tests

- `go run ./internal/checks/agent-workflow`

## Docs Sync

- `tasks/milestones/ROADMAP.md`
- `tasks/milestones/STATUS.md`

## Validation

- `go run ./internal/checks/agent-workflow`
- `gofmt -l .`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Added `tasks/milestones/ARCHIVE_INDEX.md` as the canonical archive-coverage
truth file, documented how superseded drafts relate to canonical milestones,
and updated `ROADMAP.md` plus `STATUS.md` so on-disk archive coverage and
current local milestone execution state are explicit.
