# Card 0827

Priority: P1
State: active
Primary Module: router
Owned Files:
- `router/router.go`
- `router/params.go`
- `router/module.yaml`
- `docs/modules/router/README.md`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Prune `router` down to its pure route-structure API so the stable public surface matches the documented boundary exactly.

Problem:
- `router` still exports app-level convenience surface that does not belong to a route-structure module: `WithLogger`, `SetLogger`, `Logger`, and logger-carrying group propagation.
- `router` also exports stdlib-shadow aliases such as `Handler` and `HandlerFunc`, plus an extra `ParamFromRequest` helper alongside `Param`, which expands the stable surface without adding a distinct ownership boundary.
- `docs/modules/router/README.md` and `docs/modules/core/README.md` already describe a narrower kernel/router split than the code actually enforces.

Scope:
- Remove app-level logger carriage from `router` completely.
- Remove redundant stdlib-shadow and duplicate param-access surface so `router` exposes one clear public API.
- Sync router and core module docs/manifests to the reduced, canonical route-structure surface.

Non-goals:
- Do not change route matching behavior or reverse-routing semantics.
- Do not add replacement convenience aliases in another stable root.
- Do not preserve removed router surface for backward compatibility.

Files:
- `router/router.go`
- `router/params.go`
- `router/module.yaml`
- `docs/modules/router/README.md`
- `docs/modules/core/README.md`

Tests:
- `go test -timeout 20s ./router ./core`
- `go test -race -timeout 60s ./router ./core`
- `go vet ./router ./core`

Docs Sync:
- Keep `router` and `core` docs aligned on the rule that logger ownership stays on `core.App`, not in `router`, and that router APIs stay stdlib-shaped instead of alias-heavy.

Done Definition:
- `router` no longer carries logger state or exposes logger-related public API.
- Redundant public aliases/helpers that duplicate stdlib or other canonical router entrypoints are removed.
- Router and core docs describe the same narrowed public surface that the code implements.

Outcome:
- Completed.
- Removed router logger carriage, duplicate stdlib-shadow helper aliases, and the extra param helper surface outside the canonical router API.
- Kept stable router ownership centered on route registration, matching, params, and reverse lookup only.
