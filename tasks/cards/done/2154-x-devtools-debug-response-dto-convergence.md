# Card 2154: x/devtools Debug Response DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: x/devtools
Owned Files:
- `x/devtools/devtools.go`
- `x/devtools/devtools_test.go`
- `docs/modules/x-devtools/README.md`
Depends On: none

Goal:
Converge devtools debug JSON endpoints on local typed DTO structs while keeping
debug behavior opt-in and response fields stable.

Problem:
`x/devtools/devtools.go` still assembles several success payloads with ad hoc
`map[string]any` literals: routes JSON, middleware, config/info, metrics, clear,
and reload success responses. It also manually converts the typed
`ConfigSnapshot` into a map, duplicating the field layout already captured by
JSON tags. These debug endpoints are experimental but user-facing when enabled,
so their response contracts should be explicit and testable.

Scope:
- Replace devtools JSON success maps with local DTO structs.
- Return `ConfigSnapshot` directly instead of rebuilding it through a map.
- Keep route paths, status codes, response field names, and pprof behavior
  unchanged.
- Add focused tests that decode the contract envelope into typed DTOs for
  routes, config, metrics, clear, reload, and info responses.
- Update the x/devtools primer with the debug DTO response policy.

Non-goals:
- Do not change whether devtools is enabled.
- Do not move debug routes into `core`.
- Do not change pprof handlers or env reload semantics.
- Do not add dependencies.

Files:
- `x/devtools/devtools.go`
- `x/devtools/devtools_test.go`
- `docs/modules/x-devtools/README.md`

Tests:
- `go test -race -timeout 60s ./x/devtools/...`
- `go test -timeout 20s ./x/devtools/...`
- `go vet ./x/devtools/...`

Docs Sync:
Update `docs/modules/x-devtools/README.md` to state that debug JSON endpoints
use local typed DTO payloads.

Done Definition:
- Devtools JSON success responses no longer use one-off maps.
- Config/info responses use the typed snapshot shape directly.
- Focused endpoint tests cover the DTO response shapes.
- The listed validation commands pass.

Outcome:
- Added local typed DTOs for routes, middleware, info, metrics, and action
  responses.
- Returned `ConfigSnapshot` directly for config and info payloads, removing the
  duplicate snapshot-to-map conversion.
- Updated endpoint tests to decode the contract envelope into typed devtools
  payloads and cover the disabled metrics shape.
- Documented the x/devtools debug JSON DTO policy.

Validation:
- `go test -race -timeout 60s ./x/devtools/...`
- `go test -timeout 20s ./x/devtools/...`
- `go vet ./x/devtools/...`
