# Card 0736: x/frontend Navigation Fallback Contract

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`
Depends On: 0735

Goal:
Make SPA fallback fail closed for missing asset requests while preserving
navigation fallback for browser page loads.

Scope:
- Only fall back to the configured index for requests that look like browser
  navigations.
- Return 404 for missing asset-like paths such as `.js`, `.css`, `.wasm`,
  images, and source maps.
- Add coverage for asset miss, extensionless navigation, Accept-based
  navigation, and fallback-disabled behavior.
- Update docs with the exact fallback contract.

Non-goals:
- Do not change router matching primitives.
- Do not add app-specific frontend route discovery.
- Do not change public option names.

Files:
- `x/frontend/frontend.go`
- `x/frontend/frontend_test.go`
- `x/frontend/README.md`
- `docs/modules/x-frontend/README.md`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
Document navigation-only fallback and missing asset 404 behavior.

Done Definition:
- Missing asset-like requests never return the SPA index.
- Navigation-like missing routes still return the index when fallback is
  enabled.
- Fallback-disabled behavior is unchanged.
- The listed validation commands pass.

Outcome:
- SPA fallback now only applies to extensionless navigation-like requests that
  have no `Accept` header or accept HTML.
- Missing asset-like paths such as `.js` and `.js.map` return 404 instead of
  the index response.
- Documentation now states the exact fallback contract.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
