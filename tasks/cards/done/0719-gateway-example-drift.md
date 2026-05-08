# Card 0719

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/gateway
Owned Files:
- `x/gateway/proxy.go`
- `x/gateway/cache/http_cache.go`
- `x/gateway/transform/transform.go`
- `docs/modules/x-gateway/README.md`
Depends On: —

Goal:
- Sync gateway package examples and comments with the current canonical core API.

Problem:
Gateway comments still show stale code such as `core.New(core.DefaultConfig())` and `app.Router().Group(...)`. Current `core.New` requires `core.AppDependencies`, and the raw `App.Router()` escape hatch has been removed.

Scope:
- Replace stale examples with current canonical wiring:
  - construct `core.New(cfg, core.AppDependencies{...})`
  - register gateway handlers through explicit `app.Any` or `app.AddRoute`
  - avoid `App.Router()` examples
- Scan gateway package docs/comments for similar stale core API references.
- Keep example changes documentation-only unless tests currently compile package examples.

Non-goals:
- Do not change gateway runtime behavior.
- Do not introduce new route helper aliases.
- Do not update unrelated extension examples.

Files:
- `x/gateway/proxy.go`
- `x/gateway/cache/http_cache.go`
- `x/gateway/transform/transform.go`
- `docs/modules/x-gateway/README.md`

Tests:
- `go test -timeout 20s ./x/gateway/...`
- `go vet ./x/gateway/...`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- Required. This card is documentation/example sync.

Done Definition:
- Gateway examples no longer reference removed or outdated core APIs.
- Gateway package tests and vet still pass.
- Reference app remains the canonical bootstrap path.

Outcome:
- Replaced stale gateway comments that used `core.New(core.DefaultConfig())` with current `core.New(cfg, core.AppDependencies{...})` wiring.
- Removed `App.Router()` examples from gateway comments and replaced them with explicit `app.Any` / `app.AddRoute` registration.
- Updated gateway cache and transform package examples to check `app.Use` errors.
- Synced stale `x/gateway/DESIGN.md` snippets away from removed `core.New()`, path-scoped `app.Use`, and `app.Start` APIs.
- Documented the current core app example shape in `docs/modules/x-gateway/README.md`.
- Validation:
  - `go test -timeout 20s ./x/gateway/...`
  - `go vet ./x/gateway/...`
  - `go run ./internal/checks/reference-layout`
