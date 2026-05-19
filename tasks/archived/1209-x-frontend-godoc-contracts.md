# Card 1209: x/frontend Godoc Contracts

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/config.go`
- `x/frontend/mount.go`
Depends On: 0757

Goal:
Make exported x/frontend API contracts understandable from Go documentation.

Scope:
- Document that `Option` is intentionally sealed by the package-private config
  target.
- Document `Registrar` snapshot versus AddRoute-only registration behavior.
- Document `RegisterFS`, `NewMountFS`, and `NewHandlerFS` lazy custom
  filesystem semantics and `http.Dir` normalization.

Non-goals:
- Do not change exported API shape.
- Do not change runtime behavior.
- Do not promote module status.

Files:
- `x/frontend/config.go`
- `x/frontend/mount.go`

Tests:
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Go documentation is the docs target for this card.

Done Definition:
- Exported Go comments carry the stable-relevant API contracts.
- No runtime behavior changes occur.
- The listed validation commands pass.

Outcome:
- Documented sealed `Option` semantics in exported Go comments.
- Documented snapshot-capable registrar preflight and AddRoute-only sequential
  partial-registration semantics on `Registrar` and `Mount.Register`.
- Documented `RegisterFS`, `NewMountFS`, and `NewHandlerFS` filesystem
  semantics for `http.Dir` versus other lazy custom filesystems.
- Validation passed:
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
