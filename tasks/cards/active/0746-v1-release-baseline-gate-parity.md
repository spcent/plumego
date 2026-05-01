# Card 0746

Milestone: v1
Recipe: specs/change-recipes/doc-sync.yaml
Priority: P0
State: active
Primary Module: release
Owned Files:
- `docs/release/v1.0.0-rc.1.md`
- `.github/workflows/quality-gates.yml`
- `Makefile`
- `specs/checks.yaml`
- `tasks/cards/active/README.md`
Depends On:

Goal:
- Freeze the v1 release baseline and prove that local gates, CI gates, and the release checklist use the same live control-plane paths.

Problem:
v1 readiness depends on a trustworthy baseline. If the release note, workflow, and local gate entrypoints drift, agents can produce passing local evidence while GitHub blocks the tag, or document a release target that does not match the actual branch and commit.

Scope:
- Confirm the active v1 preparation branch and release target commit.
- Update release notes so branch, commit, and gate expectations match the current repository state.
- Verify CI does not depend on unavailable runner tools such as `rg`.
- Verify `make gates` includes the same boundary, maturity, beta-evidence, deprecation, CLI, and website checks expected by CI.
- Record any gate mismatch as a follow-up card instead of silently widening scope.

Non-goals:
- Do not change stable public APIs.
- Do not promote any `x/*` module.
- Do not add new release checks unless an existing documented gate is missing.

Files:
- `docs/release/v1.0.0-rc.1.md`
- `.github/workflows/quality-gates.yml`
- `Makefile`
- `specs/checks.yaml`
- `tasks/cards/active/README.md`

Tests:
- `make gates`
- `go run ./internal/checks/agent-workflow`
- `git diff --exit-code`

Docs Sync:
- Required for release notes and active queue ordering.

Done Definition:
- `docs/release/v1.0.0-rc.1.md` names the actual release branch and target commit or explicitly marks them as pending tag-target values.
- CI and `make gates` use the same required live paths for v1 release readiness.
- No known runner-only dependency remains in the GitHub gate workflow.
- Any remaining mismatch is captured as a new bounded card.
