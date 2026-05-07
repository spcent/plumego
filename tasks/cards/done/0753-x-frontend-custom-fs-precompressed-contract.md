# Card 0753: x/frontend Custom FS Precompressed Contract

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`
Depends On: 0752

Goal:
Make the custom filesystem precompressed probing cost an explicit stable
contract decision.

Scope:
- Document that non-`http.Dir` custom filesystems keep lazy probing.
- Document that lazy probing may open `.br` and `.gz` candidates on original
  responses to preserve `Vary: Accept-Encoding` correctness.
- Keep directory-backed immutable variant planning unchanged.

Non-goals:
- Do not add a new public API.
- Do not cache custom filesystem variant misses.
- Do not change runtime behavior.

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
- Custom filesystem lazy precompressed probing cost is explicitly documented.
- Directory-backed behavior remains the recommended production path.
- The listed validation commands pass.

Outcome:
- Documented that non-`http.Dir` custom filesystems intentionally keep lazy
  `.br`/`.gz` probing to preserve dynamic filesystem behavior and
  `Vary: Accept-Encoding` correctness.
- Documented that original responses may probe compressed candidates and that
  directory-backed mounts avoid that per-request cost through immutable variant
  metadata.
- Added the probing cost to the module risk list.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go run ./internal/checks/extension-maturity`
  - `go run ./internal/checks/extension-beta-evidence`
