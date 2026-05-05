# Card 0755: x/frontend Registrar Atomicity Contract

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P2
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml`
Depends On: 0754

Goal:
Make frontend route registration atomicity guarantees explicit for stable use.

Scope:
- Document that snapshot-capable registrars get duplicate-route preflight before
  mutation.
- Document that AddRoute-only custom registrars are best-effort sequential and
  may partially register if a later route fails.
- Keep the current `Registrar` interface unchanged.

Non-goals:
- Do not require all custom registrars to expose route snapshots.
- Do not add rollback behavior.
- Do not change router behavior.

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
This is a contract documentation card.

Done Definition:
- Atomicity and partial-registration behavior are explicit.
- No public API change occurs.
- The listed validation commands pass.
