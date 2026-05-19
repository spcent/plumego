# Card 1503

Milestone: M-008
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: release
Owned Files:
- git tag v1.1.0

Goal:
- Create the annotated git tag v1.1.0 on the current HEAD of the milestone branch after all
  pre-release checks pass.

Scope:
- Verify gate output from card 1500 shows exit 0.
- Verify snapshot comparison from card 1502 shows zero exported-symbol changes.
- Create an annotated tag: `git tag -a v1.1.0 -m "v1.1.0 release — gate output in docs/release/v1.1.0.md"`.
- Confirm the tag exists with `git tag --list v1.1.0`.

Non-goals:
- Do not push the tag to the remote in this card; that happens at the end of Phase 4.
- Do not create additional patch tags (v1.1.1, etc.) in this card.
- Do not modify any source files.

Files:
- git tag v1.1.0 (in-repository artifact)
- `docs/release/v1.1.0.md` (read-only; tag message references this file)

Tests:
- `git tag --list v1.1.0`

Docs Sync:
- none

Done Definition:
- `git tag --list v1.1.0` outputs `v1.1.0`.
- Tag is annotated (not lightweight): `git show v1.1.0` shows the tag message.
- Tag message references docs/release/v1.1.0.md by path.

Outcome:
- Done. Created annotated tag `v1.1.0` on HEAD `58aad7ee`; `git show
  v1.1.0` shows the tag message `v1.1.0 release - gate output in
  docs/release/v1.1.0.md`.
