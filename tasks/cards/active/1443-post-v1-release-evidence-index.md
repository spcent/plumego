# Card 1443

Milestone: M-006
Recipe: specs/change-recipes/docs-only.yaml
Priority: P2
State: active
Primary Module: release
Owned Files:
- `docs/release/`
- `tasks/milestones/M-006.verify.md`
Depends On:
- 1442

Goal:
- Add a concise post-v1 evidence index that points agents to the exact tag,
  run IDs, release docs, and remaining extension blockers.

Scope:
- Create or update release evidence index docs under `docs/release/`.
- Create `tasks/milestones/M-006.verify.md` with M-006 card evidence.
- Keep extension blocker state synchronized with
  `specs/extension-beta-evidence.yaml`.

Non-goals:
- Do not change release tags.
- Do not promote extensions.
- Do not alter runtime behavior.

Files:
- `docs/release/`
- `tasks/milestones/M-006.verify.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/agent-workflow`
- `git status --short --branch`

Docs Sync:
- Required for post-v1 evidence discoverability.

Done Definition:
- Agents have one evidence index for post-v1 release facts.
- M-006 verify artifact exists.
- Remaining extension blockers are still explicit.
