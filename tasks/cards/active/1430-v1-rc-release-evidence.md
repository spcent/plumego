# Card 1430

Milestone: M-005
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: active
Primary Module: release
Owned Files:
- `docs/release/v1.0.0-rc.1.md`
- `tasks/cards/active/1375-v1-rc-tag-and-observation-window.md`
- `tasks/cards/active/README.md`
Depends On:
- 1429

Goal:
- Create and verify `v1.0.0-rc.1` as the release candidate evidence anchor.

Scope:
- Create the annotated rc tag from the reviewed release commit.
- Run local release gates and record the result.
- Push tag/branch and record GitHub gate status.

Non-goals:
- Do not tag final `v1.0.0`.
- Do not promote extension status.
- Do not change runtime behavior unless a release blocker card is opened.

Files:
- `docs/release/v1.0.0-rc.1.md`
- `tasks/cards/active/1375-v1-rc-tag-and-observation-window.md`
- `tasks/cards/active/README.md`

Tests:
- `make gates`
- `go run ./internal/checks/extension-beta-evidence`
- `git tag -l 'v1.0.0-rc.1'`

Docs Sync:
- Required for release notes and observation-window status.

Done Definition:
- `v1.0.0-rc.1` exists locally and remotely.
- Local and GitHub gates are recorded.
- Final v1 is either ready for observation or explicitly blocked.

Outcome:
- Prepared `docs/release/v1.0.0-rc.1.md` to resolve the final tag target from
  the annotated `v1.0.0-rc.1` tag created by this card.
- Linked existing card 1375 to the M-005 release-execution path.
- First local `make gates` run failed in `x/data/sharding` because
  `TestLoggingRouter_LogShardResolution` scanned the full JSON log line for
  `123`; blocker card 1435 fixed the false positive with structured JSON field
  assertions.
- Re-ran `GOCACHE=/private/tmp/plumego-gocache GOMODCACHE=/private/tmp/plumego-gomodcache make gates`; all gates passed, including race tests, normal tests, stable-root coverage, `cmd/plumego` submodule checks, and website sync/check/build.
- Stable-module coverage reported by `make gates`: 87.2%.
- Committed local gate evidence at `a596805e`.
- Local tag creation is blocked: normal `git tag -a v1.0.0-rc.1 -m "Plumego v1.0.0-rc.1"` failed with `.git` temporary-file permission errors, and two escalated approval requests timed out.
- After explicit user approval to continue, two additional escalated tag
  attempts timed out before the approval review completed. The local
  `v1.0.0-rc.1` tag still does not exist.
- Remote GitHub gate evidence remains pending until the release branch and rc
  tag are pushed.
