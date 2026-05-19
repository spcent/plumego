# Milestone Status Matrix

This file reconciles active milestone specs with the current roadmap. It is a
truth matrix for agents; it does not mark a milestone complete by itself.

## Current State

| Milestone | Milestone file | Roadmap relationship | Current action |
| --- | --- | --- | --- |
| M-001 v1 trust baseline | `tasks/milestones/done/M-001-v1-trust-baseline/M-001.md` | Reflected in current docs and CLI template smoke tests | Archived; verify artifact exists |
| M-002 stable roots freeze | `tasks/milestones/done/M-002-stable-roots-freeze-and-reliability/M-002.md` | Stable API inventory and focused stable-root checks are recorded | Archived; final race/vet release evidence moved to M-005 |
| M-003 extension evidence pipeline | `tasks/milestones/done/M-003-extension-evidence-pipeline/M-003.md` | Implemented through extension evidence docs, snapshots, checks, and beta evidence template | Archived; remaining experimental blockers are explicit |
| M-004 stable root cleanup freeze | `tasks/milestones/done/M-004-stable-root-cleanup-freeze/M-004.md` | Outcome section records cards 1394-1401 and stable cleanup gates | Archived |
| M-005 v1 release execution | `tasks/milestones/done/M-005-v1-release-execution/M-005.md` | rc.1 to final v1/no-go execution completed | Archived; `v1.0.0` evidence recorded |
| M-006 v1.0.1 maintenance lane | `tasks/milestones/done/M-006-v1-0-1-maintenance-lane/M-006.md` | Post-v1 maintenance cards completed | Archived |
| M-007 extension v1 baseline evidence intake | `tasks/milestones/done/M-007-extension-v1-baseline-evidence-intake/M-007.md` | First `v1.0.0` release-ref intake completed | Archived; remaining beta closures are blocked |

## Reconciliation Rules

- Do not move an active milestone to `done/` without an `## Outcome` section.
- Do not mark roadmap phases complete from prose alone; use command evidence.
- If a roadmap phase says `substantially complete`, keep the corresponding
  milestone active until the remaining blocker is explicit.
- Prefer adding a verify artifact over editing historical task wording.

## Verify Artifacts

These artifacts back the archived milestones:

- `tasks/milestones/done/M-001-v1-trust-baseline/verify-M-001.md`: created by card 1429; contains template
  truth matrix, release tag check, and CLI smoke output.
- `tasks/milestones/done/M-002-stable-roots-freeze-and-reliability/verify-M-002.md`: created by card 1429; contains stable API
  inventory, stable-root test output, boundary checks, and freeze evidence.
- `tasks/milestones/done/M-003-extension-evidence-pipeline/verify-M-003.md`: created by card 1429; contains extension
  maturity and extension beta evidence check results.
- `tasks/milestones/done/M-005-v1-release-execution/verify-M-005.md`: rc tag evidence, stable-root final
  freeze evidence, CLI onboarding smoke evidence, extension maturity evidence,
  and final GO/NO-GO decision.
