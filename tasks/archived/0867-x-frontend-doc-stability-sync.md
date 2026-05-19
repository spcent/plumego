# Card 0867: x/frontend Docs and Stability Evidence Sync

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `x/frontend/module.yaml`
- `docs/EXTENSION_MATURITY.md`
Depends On: 0728

Goal:
Synchronize frontend documentation and module evidence with the hardened
implementation while keeping status experimental unless promotion criteria are
fully met.

Scope:
- Remove stale or nonexistent API examples.
- Align README behavior descriptions with actual error, embedded, cache, and
  test behavior.
- Record remaining stable promotion blockers in docs or module metadata.
- Keep compatibility claims limited to implemented and tested behavior.

Non-goals:
- Do not promote `x/frontend` to stable without owner sign-off and release
  evidence.
- Do not document planned behavior as implemented.
- Do not change code behavior in this card.

Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `x/frontend/module.yaml`
- `docs/EXTENSION_MATURITY.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go test -timeout 20s ./x/frontend/...`

Docs Sync:
This is the docs sync card.

Done Definition:
- Public examples compile conceptually against exported APIs.
- Docs use the correct module test paths and response semantics.
- Remaining stable-readiness blockers are explicit.
- The listed validation commands pass.

Outcome:
- Synced frontend README behavior for precompressed negotiation, Vary, custom
  page validation, custom page cache policy, 405 Allow headers, and symlink
  safety.
- Removed the stale nonexistent `frontend.HasEmbedded()` example and corrected
  the module test path.
- Recorded remaining stable-readiness blockers without promoting module status.
- Updated the module manifest risk and review checklist entries within manifest
  size limits.
- Validation passed:
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/agent-workflow`
  - `go test -timeout 20s ./x/frontend/...`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-maturity`
