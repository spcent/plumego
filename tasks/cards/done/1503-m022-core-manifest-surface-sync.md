# Card 1503

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/module.yaml`
Depends On:

Goal:
- Bring `core/module.yaml` back into sync with the exported lifecycle and
  route-group surface.
- Remove redundant or ghost `forbidden_imports` entries that no longer map to
  real package ownership rules.

Scope:
- Update only the `core` manifest entries that describe exported API and
  disallowed imports.

Non-goals:
- Do not change `core` runtime behavior.
- Do not rename or remove exported `core` symbols.
- Do not widen the card into `docs/` unless the manifest wording forces a doc
  correction.

Files:
- `core/module.yaml`

<!-- none; manifest-only card -->

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/public-entrypoints-sync`

Docs Sync:
- None expected unless the manifest wording diverges from `docs/modules/core/README.md`.

Validation:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/public-entrypoints-sync`
- `go run ./internal/checks/agent-workflow`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Added the missing `core` public entrypoints for `Run`, `Group`, `RouteGroup`,
  and the full `RouteGroup` method surface.
- Removed redundant `forbidden_imports` entries that duplicated `x/**` or
  pointed at discouraged top-level roots rather than real import paths.
- Validation:
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/public-entrypoints-sync`
  - `go run ./internal/checks/agent-workflow`
