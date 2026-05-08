# Card 0726

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: blocked
Primary Module: x/observability
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-observability.md`
- `docs/extension-evidence/release-artifacts.md`
Depends On: release refs and owner sign-off

Goal:
- Complete beta evidence closure for `x/observability` when real release refs and owner sign-off are available.

Problem:
The evidence ledger has an `x/observability` evidence doc and current-head snapshot, but beta promotion remains blocked by missing release history, matching release snapshots, and owner sign-off.

Scope:
- Add two real release refs only after tags or release commits exist.
- Generate release-to-release API snapshots with `extension-release-evidence`.
- Record owner sign-off from `observability`.
- Keep blockers until all evidence is present.

Non-goals:
- Do not promote `x/observability` without complete evidence.
- Do not use `HEAD` as release-history evidence.
- Do not change observability runtime behavior.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-observability.md`
- `docs/extension-evidence/release-artifacts.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Required when evidence is added.

Done Definition:
- `x/observability` has two release refs, matching release snapshots, and owner sign-off, or the blocker remains explicit.

Outcome:
-
