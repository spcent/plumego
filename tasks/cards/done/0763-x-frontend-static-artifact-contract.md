# Card 0763: x/frontend Static Artifact Contract

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`
Depends On: 0762

Goal:
Make directory-backed frontend assets an explicit static deployment artifact
contract.

Scope:
- Document that directory-backed precompressed plans are construction-time
  metadata, not runtime atomic snapshots.
- Document that deploys should replace frontend bundles atomically outside this
  package when assets can change during service runtime.
- Record mutable directory updates as a stable risk.

Non-goals:
- Do not add file watching.
- Do not add runtime locking.
- Do not change request behavior.

Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
This is a behavior contract documentation card.

Done Definition:
- Static artifact semantics are explicit.
- Runtime mutation is documented as caller/deployment responsibility.
- The listed validation commands pass.

Outcome:
- Documented directory-backed bundles as static deployment artifacts.
- Clarified that construction-time precompressed metadata is not a runtime
  atomic snapshot for in-place file mutations.
- Recorded mutable directory asset updates as a module risk.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go run ./internal/checks/extension-maturity`
  - `go run ./internal/checks/extension-beta-evidence`
