# Card 1372

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: blocked
Primary Module: x/gateway/discovery
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-discovery.md`
- `docs/extension-evidence/release-artifacts.md`
Depends On: path-migration release evidence and owner sign-off

Goal:
- Complete beta evidence closure for the `x/gateway/discovery` core/static surface.
- This card is part of the post-v1 maturity roadmap recorded in `tasks/cards/active/1452-post-v1-maturity-roadmap.md`.

Problem:
The evidence ledger tracks `x/gateway/discovery:core-static` with a first `v1.0.0`
release ref and checked-in baseline snapshot, but it still lacks a second
release ref, complete release-backed snapshots, and owner sign-off.

Scope:
- Add the second real release ref only after the next qualifying tag or release
  commit exists.
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
- `x/gateway/discovery:core-static` has checked-in snapshots, two release refs, release snapshots, and owner sign-off, or blockers remain explicit.

Outcome:
- Blocked. `go run ./internal/checks/extension-release-evidence -module
  ./x/gateway/discovery -base v1.0.0 -head v1.1.0` cannot generate a base
  snapshot because `x/gateway/discovery` does not exist at `v1.0.0`; the
  checked-in v1 baseline uses the old `./x/discovery` path. Promotion needs an
  explicit path-migration evidence card or another release interval where the
  `x/gateway/discovery` path exists at both refs.
