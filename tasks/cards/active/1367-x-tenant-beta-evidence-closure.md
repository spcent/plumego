# Card 1367

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: active
Primary Module: x/tenant
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-tenant.md`
- `docs/extension-evidence/release-artifacts.md`
Depends On: release-backed snapshots and owner sign-off

Goal:
- Complete beta evidence closure for `x/tenant` when real release refs and owner sign-off are available.
- This card is part of the post-v1 maturity roadmap recorded in `tasks/cards/active/1452-post-v1-maturity-roadmap.md`.

Problem:
The evidence ledger has an `x/tenant` evidence doc, current-head snapshot, first
`v1.0.0` release ref, and v1 baseline snapshot artifacts, but beta promotion
remains blocked by the missing second release ref, complete release-backed
snapshots, and owner sign-off.

Scope:
- Add the second real release ref only after the next qualifying tag or release
  commit exists.
- Generate release-to-release API snapshots with `extension-release-evidence`.
- Record owner sign-off from `multitenancy`.
- Keep blockers until all evidence is present.

Non-goals:
- Do not promote `x/tenant` without complete evidence.
- Do not use `HEAD` as release-history evidence.
- Do not change tenant runtime behavior.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-tenant.md`
- `docs/extension-evidence/release-artifacts.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Required when evidence is added.

Done Definition:
- `x/tenant` has two release refs, matching release snapshots, and owner sign-off, or the blocker remains explicit.

Outcome:
-
