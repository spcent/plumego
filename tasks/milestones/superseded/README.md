# Superseded Milestone Drafts

This directory holds milestone drafts that were found in
`tasks/milestones/active/` but are not the canonical executable milestone spec.

Canonical active milestones use the directory form:

```text
tasks/milestones/active/M-NNN-short-name/
  M-NNN.md
  plan-M-NNN.md
  verify-M-NNN.md
```

Files here are retained for history only. Do not execute them directly; use the
matching directory-form milestone under `tasks/milestones/active/` or create a
new milestone if the superseded draft describes follow-up work.

## Current Coverage Map

Use `tasks/milestones/ARCHIVE_INDEX.md` as the canonical truth source for how
these drafts relate to milestone history on disk.

Current relationships:

- `M-012-beta-validation-draft.md` is superseded by
  `tasks/milestones/done/M-012-input-validation-bridge/`
- `M-014-openapi-beta-draft.md` is superseded by
  `tasks/milestones/done/M-014-openapi-3-1-generation/`
- `M-011-internal-bench-draft.md` and `M-013-stdlib-migration-draft.md` remain
  historical draft artifacts only; no canonical `done/` directory currently
  exists for those milestone ids in this repository state
