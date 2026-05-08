# Card 1138

Milestone: v1
Recipe: specs/change-recipes/doc-sync.yaml
Priority: P0
State: done
Primary Module: release
Owned Files:
- `docs/release/v1.0.0-rc.1.md`
- `docs/release/`
- `docs/stable-api/`
- `docs/extension-evidence/`
- `tasks/cards/active/README.md`
Depends On: 0746, 0747, 0748, 0749, 0750, 0751

Goal:
- Package the complete v1.0.0-rc.1 release evidence for review and tagging.

Problem:
The release candidate needs a single reviewable evidence bundle that captures the exact commit, gate results, stable API snapshots, extension blockers, compatibility decisions, and generated-doc status.

Scope:
- Update `docs/release/v1.0.0-rc.1.md` with final commit, branch, and gate status.
- Link or list stable API snapshots and extension evidence snapshots.
- Record final compatibility decisions and known blockers.
- Record generated website/docs sync status.
- Confirm no generated file drift remains after release checks.

Non-goals:
- Do not change runtime behavior.
- Do not add new release scope.
- Do not resolve post-v1 extension evidence blockers in this card.

Files:
- `docs/release/v1.0.0-rc.1.md`
- `docs/release/`
- `docs/stable-api/`
- `docs/extension-evidence/`
- `tasks/cards/active/README.md`

Tests:
- `make gates`
- `git diff --exit-code`
- `go run ./internal/checks/deprecation-inventory -strict`

Docs Sync:
- Required. This card is the release documentation sync point.

Done Definition:
- The release evidence package can be reviewed without rerunning discovery.
- Gate results are tied to a concrete commit.
- Known blockers are explicit and limited to post-v1 extension maturity or documented non-blockers.
- The worktree has no generated drift after validation.

Outcome:
- Added local gate evidence to `docs/release/v1.0.0-rc.1.md`, including gate command, deprecation inventory command, generated website drift status, and final tag target handoff to card 0753.
- Confirmed stable API snapshots and extension evidence snapshots are linked from the release notes.
- Kept known blockers limited to post-v1 extension beta evidence: release history, API snapshots, and owner sign-off.
- Did not change runtime behavior or extension maturity.
- Validation passed:
  - `GOCACHE=/private/tmp/plumego-gocache make gates`
  - `git diff --exit-code`
  - `GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/deprecation-inventory -strict`
