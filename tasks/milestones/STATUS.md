# Milestone Status Matrix

This file reconciles active milestone specs with the current roadmap. It is a
truth matrix for agents; it does not mark a milestone complete by itself.

## Current State

| Milestone | Active file | Roadmap relationship | Current action |
| --- | --- | --- | --- |
| M-001 v1 trust baseline | `tasks/milestones/active/M-001.md` | Mostly reflected in current docs and CLI template smoke tests | Keep active until a verify artifact records command evidence and release-claim audit output |
| M-002 stable roots freeze | `tasks/milestones/active/M-002.md` | Stable API inventory exists; `contract`, `router`, and `core` now have explicit freeze behavior matrices and focused regression tests | Continue with `middleware`, `security`, `store`, `health`, `log`, and `metrics` freeze coverage |
| M-003 extension evidence pipeline | `tasks/milestones/active/M-003.md` | Partially implemented through extension evidence docs, snapshots, checks, and the beta evidence template | Keep active until `x/rest` has release refs and owner sign-off, then produce the verify artifact |
| M-004 stable root cleanup freeze | `tasks/milestones/active/M-004.md` | Outcome section says cards 1394-1401 and stable cleanup gates are complete | Reconcile in card 1429 before archiving or leaving active |
| M-005 v1 release execution | `tasks/milestones/active/M-005.md` | New sequential release execution plan from rc.1 to final v1/no-go | Start with card 1429, then card 1430 for `v1.0.0-rc.1` evidence |

## Reconciliation Rules

- Do not move an active milestone to `done/` without an `## Outcome` section.
- Do not mark roadmap phases complete from prose alone; use command evidence.
- If a roadmap phase says `substantially complete`, keep the corresponding
  milestone active until the remaining blocker is explicit.
- Prefer adding a verify artifact over editing historical task wording.

## Next Verify Artifacts

Create these before archiving any active milestone:

- `tasks/milestones/M-001.verify.md`: template truth matrix, release tag check,
  CLI smoke output.
- `tasks/milestones/M-002.verify.md`: stable API inventory, stable-root test
  output, boundary checks, and the freeze matrices for `contract`, `router`,
  and `core`.
- `tasks/milestones/M-003.verify.md`: `x/rest` beta evidence sample,
  extension maturity check, extension beta evidence check.
- `tasks/milestones/M-005.verify.md`: rc tag evidence, stable-root final
  freeze evidence, CLI onboarding smoke evidence, extension maturity evidence,
  and final GO/NO-GO decision.
