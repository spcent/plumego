# Card 1441

Milestone: M-006
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: active
Primary Module: CI
Owned Files:
- `.github/workflows/quality-gates.yml`
Depends On:
- 1440

Goal:
- Address or explicitly classify GitHub Actions Node runtime deprecation
  warnings seen in v1 tag and main gate runs.

Scope:
- Inspect the workflow action versions.
- Update action versions only if official upstream releases provide a clear
  Node 24-compatible path.
- Otherwise record the warning as a tracked CI maintenance blocker without
  changing runtime behavior.

Non-goals:
- Do not change gate semantics.
- Do not remove existing jobs.
- Do not pin unverified third-party actions.

Files:
- `.github/workflows/quality-gates.yml`

Tests:
- `git diff --check`
- `gh workflow view quality-gates.yml --repo spcent/plumego`
- Remote GitHub Actions run evidence after push, if the workflow changes.

Docs Sync:
- Required only if the warning is recorded instead of fixed.

Done Definition:
- Node runtime warning has either a committed workflow fix or a documented
  blocker with remote evidence.
