# Card 0749: x/frontend Public API Freeze Ledger

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
Depends On: 0748

Goal:
Make the intended stable public API surface explicit without promoting
`x/frontend`.

Scope:
- Enumerate the public API symbols that must remain stable before promotion.
- Clarify that `Option` is intentionally sealed and extension should happen
  through exported `With*` helpers.
- Record that `RegisterFromDir`, `http.Dir` inputs, and custom filesystem inputs
  have distinct startup/safety contracts.
- Keep current status `experimental`.

Non-goals:
- Do not export `config` or add a public builder.
- Do not change runtime behavior.
- Do not clear release-backed evidence blockers.

Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
This is a public API contract documentation card.

Done Definition:
- Public API freeze candidates are listed in package/module evidence.
- Sealed `Option` intent is unambiguous.
- No status promotion occurs.
- The listed validation commands pass.

Outcome:
- Added explicit stable candidate API ledgers to `x/frontend` package docs,
  module primer, evidence docs, and module manifest comments.
- Documented `Mount` methods, `Registrar`, constructors, registration helpers,
  sealed `Option`, and all current `With*` helpers as release-snapshot
  comparison candidates.
- Preserved `experimental` status and did not export config or add a builder.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go run ./internal/checks/extension-api-snapshot -compare docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot`
  - `go run ./internal/checks/extension-maturity`
