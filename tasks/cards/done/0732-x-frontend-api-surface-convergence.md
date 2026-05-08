# Card 0732: x/frontend API Surface Convergence

Milestone: none
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/embedded_fs.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0731

Goal:
Clarify the public API surface before stable promotion by making router
registration contracts explicit and removing misleading package-owned embedded
helpers.

Scope:
- Export the router registration interface used by frontend mounts.
- Update public constructors and registration helpers to use the exported
  interface name.
- Remove `RegisterEmbedded` and `NewMountEmbedded` after enumerating all
  in-repo call sites.
- Update tests and docs to use caller-owned `embed.FS` through `RegisterFS`.

Non-goals:
- Do not add application bootstrap behavior.
- Do not change handler-only construction.
- Do not introduce deprecation wrappers.

Files:
- `x/frontend/frontend.go`
- `x/frontend/embedded_fs.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `rg -n --glob '*.go' 'RegisterEmbedded|NewMountEmbedded' .`
- `rg -n --glob '*.go' 'routeRegistrar' x/frontend`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`
- `go build ./...`

Docs Sync:
Update frontend README and module primer to reflect the supported embedded
asset path.

Done Definition:
- No in-repo references to removed embedded helper symbols remain.
- Exported registration contracts are named and documented.
- Docs no longer suggest package-owned embedded helpers for applications.
- The listed validation commands pass.

Outcome:
- Exported the frontend router registration contract as `Registrar`.
- Updated frontend registration and mount APIs to use the exported interface
  name.
- Removed package-owned `RegisterEmbedded` and `NewMountEmbedded` helpers after
  enumerating in-repo call sites.
- Updated tests and docs to use caller-owned embedded filesystems through
  `RegisterFS`.
- Validation passed:
  - `rg -n --glob '*.go' 'RegisterEmbedded|NewMountEmbedded' .` returned no
    Go call sites.
  - `rg -n --glob '*.go' 'routeRegistrar' x/frontend` returned no results.
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
  - `go build ./...`
