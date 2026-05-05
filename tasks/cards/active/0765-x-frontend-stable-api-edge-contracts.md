# Card 0765: x/frontend Stable API Edge Contracts

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P1
State: active
Primary Module: x/frontend
Owned Files:
- `x/frontend/mount.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`
Depends On: 0764

Goal:
Freeze small public API edge semantics before stable promotion.

Scope:
- Document `Mount.Prefix` nil receiver behavior.
- Document `Mount.Handler` nil receiver behavior.
- Confirm AddRoute-only registrar partial registration remains an accepted
  public contract unless changed before stable promotion.

Non-goals:
- Do not change `Mount` method behavior.
- Do not add rollback to AddRoute-only registrars.
- Do not promote module status.

Files:
- `x/frontend/mount.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
- `docs/extension-evidence/x-frontend.md`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Go documentation and module docs must agree.

Done Definition:
- Nil receiver behavior is documented as intentional.
- Registrar partial-registration behavior remains explicit.
- The listed validation commands pass.
