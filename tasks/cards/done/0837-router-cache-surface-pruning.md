# Card 0837

Priority: P1
State: active
Primary Module: router
Owned Files:
- `router/cache.go`
- `router/router.go`
- `router/module.yaml`
- `docs/modules/router/README.md`
- `core`

Goal:
- Remove route-cache implementation and tuning helpers from the public stable router surface so `router` exposes only structural routing APIs.
- Converge cache behavior behind `Router` internals instead of exporting cache-specific types and convenience constructors.

Problem:
- `router/module.yaml` and `docs/modules/router/README.md` describe a compact public surface centered on `Router`, `Group`, and `Param`.
- The code still exports `RouteCache`, `NewRouteCache`, `CacheStats`, `Router.CacheStats()`, and `NewRouterWithCacheCapacity(...)`, which exposes route-cache implementation details and performance-tuning entrypoints as part of the stable public API.
- Most repository usage is internal tests and router internals, not real external application wiring, which suggests these exports are implementation spill rather than intentional stable surface.
- Keeping cache types public weakens the “route structure only” boundary and expands the stable API with performance internals that are not part of canonical route registration.

Scope:
- Internalize route-cache implementation types and helpers that are not part of canonical route registration.
- Remove cache-specific convenience constructors that duplicate the canonical `NewRouter(...RouterOption)` path.
- Decide whether any cache tuning knob still belongs as a `RouterOption`; if not, remove it as part of the same convergence.
- Update router/core tests and any internal call sites in the same change.
- Sync router docs and manifest to the reduced public surface.

Non-goals:
- Do not change route matching semantics, reverse routing behavior, or path parameter extraction.
- Do not move cache logic into `core`.
- Do not add a new router diagnostics package as a replacement compatibility layer.
- Do not preserve deleted cache exports via aliases or wrappers.

Files:
- `router/cache.go`
- `router/router.go`
- `router/module.yaml`
- `docs/modules/router/README.md`
- `core`

Tests:
- `go test -timeout 20s ./router/... ./core ./x/devtools/...`
- `go test -race -timeout 60s ./router/... ./core ./x/devtools/...`
- `go vet ./router/... ./core ./x/devtools/...`

Docs Sync:
- Keep the router manifest and primer aligned on the rule that public router ownership is route structure and registration, not exported cache implementation details.

Done Definition:
- Stable `router` no longer exports route-cache implementation types or duplicate cache-specific constructors beyond the canonical router entrypoint.
- Any remaining cache tuning is either internal or represented by one canonical router option surface.
- Router/core tests compile with no residual references to deleted route-cache exports.
- Router docs and manifest describe the same reduced public surface the code implements.

Outcome:
- Completed.
- Internalized route-cache implementation types and removed duplicate cache-specific router constructors from the stable public surface.
- Stable router ownership now exposes the canonical router entrypoint without exported cache internals or tuning helpers.
