# Card 0730

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: blocked
Primary Module: x/discovery
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-discovery.md`
- `docs/extension-evidence/release-artifacts.md`
Depends On: release refs and owner sign-off

Goal:
- Complete beta evidence closure for the `x/discovery` core/static surface.

Problem:
The evidence ledger tracks `x/discovery:core-static`, but it currently lacks checked-in API snapshots, release history, release snapshots, and owner sign-off.

Scope:
- Generate a current-head snapshot for the core/static discovery surface when ready for candidate review.
- Add two real release refs only after tags or release commits exist.
- Generate release-to-release API snapshots with `extension-release-evidence`.
- Record owner sign-off from `edge`.

Non-goals:
- Do not promote provider-specific Consul, Kubernetes, or etcd adapters from this card.
- Do not use `HEAD` as release-history evidence.
- Do not change discovery runtime behavior.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-discovery.md`
- `docs/extension-evidence/release-artifacts.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- Required when evidence is added.

Done Definition:
- `x/discovery:core-static` has checked-in snapshots, two release refs, release snapshots, and owner sign-off, or blockers remain explicit.

Outcome:
-
