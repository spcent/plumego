# Card 1441

Milestone: M-006
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
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

Outcome:
- `gh release list` showed v6 release lines for `actions/checkout`,
  `actions/setup-node`, and `actions/setup-go`.
- Updated `.github/workflows/quality-gates.yml` from checkout v4, setup-node
  v4, and setup-go v5 to their v6 major tags.
- Gate semantics were unchanged.
- Remote verification passed:
  - run: `25954419567`
  - URL: `https://github.com/spcent/plumego/actions/runs/25954419567`
  - result: PASS
- The previous Node 20 deprecation annotations did not reappear in the v6
  action run output.
