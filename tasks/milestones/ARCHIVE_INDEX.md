# Milestone Archive Index

This index is the canonical explanation of milestone-history coverage on disk.
Use it when a roadmap row, status row, or superseded draft exists but no
matching `done/` directory is present.

## Archive Truth Rules

- A milestone row marked `[✓]` in `ROADMAP.md` means the work is recorded as
  merged in roadmap history. It does not guarantee that a canonical
  `tasks/milestones/done/M-NNN-*/` directory exists in this repository state.
- A canonical archived milestone directory is always the strongest source when
  it exists.
- `tasks/milestones/superseded/` contains historical draft artifacts only. It
  never overrides an active or done milestone directory.
- Do not invent missing historical milestone directories retroactively. Record
  the absence here instead.

## Canonical Archived Milestone Directories Present

| Milestone | Canonical on-disk archive |
| --- | --- |
| M-012 | `tasks/milestones/done/M-012-input-validation-bridge/` |
| M-014 | `tasks/milestones/done/M-014-openapi-3-1-generation/` |
| M-015 | `tasks/milestones/done/M-015-database-adapters/` |
| M-017 | `tasks/milestones/done/M-017-grpc-support/` |
| M-021 | `tasks/milestones/done/M-021-agent-first-ecosystem-visibility/` |
| M-023 | `tasks/milestones/done/M-023-ai-resilience-convergence/` |

## Merged Milestones Recorded Only In History Files

These milestones are represented in `ROADMAP.md`, `STATUS.md`, release notes,
or other historical control-plane prose, but no canonical `done/` directory is
currently present on disk:

- M-001
- M-002
- M-003
- M-004
- M-005
- M-006
- M-007
- M-008
- M-010
- M-011
- M-013
- M-018
- M-019
- M-020

## Active Milestones Present On Disk

These milestones currently have canonical `active/` directories:

- M-009
- M-016
- M-022
- M-024

## Superseded Draft Coverage

| Draft file | Canonical milestone relationship |
| --- | --- |
| `tasks/milestones/superseded/M-011-internal-bench-draft.md` | Historical draft only; no canonical `done/` directory currently exists |
| `tasks/milestones/superseded/M-012-beta-validation-draft.md` | Superseded by `tasks/milestones/done/M-012-input-validation-bridge/` |
| `tasks/milestones/superseded/M-013-stdlib-migration-draft.md` | Historical draft only; no canonical `done/` directory currently exists |
| `tasks/milestones/superseded/M-014-openapi-beta-draft.md` | Superseded by `tasks/milestones/done/M-014-openapi-3-1-generation/` |

## Numbering Note

Milestone numbering is continuous in the roadmap ledger through M-024. Apparent
"gaps" usually mean one of two things:

- the milestone is tracked only in historical prose today, without a preserved
  `done/` directory
- the only surviving on-disk artifact is a superseded draft rather than a
  canonical milestone directory
