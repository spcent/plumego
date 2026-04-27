# Card 0436: x/gateway Safe Entrypoint Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/gateway
Owned Files:
- `x/gateway/entrypoints.go`
- `x/gateway/proxy.go`
- `x/gateway/entrypoints_test.go`
- `docs/modules/x-gateway/README.md`
Depends On: none

Goal:
Make app-facing gateway construction use the same explicit error-returning path
as the lower-level proxy constructor.

Problem:
`x/gateway/proxy.go` already exposes `NewE(config)` for callers that need
configuration errors instead of panics, but `entrypoints.go` only exposes
`NewGateway(cfg)` and `RegisterProxy(...)`. `RegisterProxy` returns
`(*GatewayProxy, error)` but constructs through `NewGateway`, so invalid gateway
configuration can still panic before the function has a chance to return an
error. This makes the app-facing gateway entrypoint less safe than the lower
level proxy surface it wraps.

Scope:
- Add a safe app-facing constructor such as `NewGatewayE` that delegates to
  `NewE`.
- Update `RegisterProxy` to construct through the safe path and return invalid
  config/backend errors instead of panicking.
- Preserve `NewGateway` as the panic-compatible wrapper if needed for existing
  callers.
- Add tests for invalid `RegisterProxy` config and the safe app-facing
  constructor.
- Keep `RegisterRoute` nil-safe/no-op behavior aligned with the current module
  primer unless this card explicitly updates docs and tests.

Non-goals:
- Do not redesign proxy lifecycle or load balancing.
- Do not change route paths, proxy retry semantics, or health-check behavior.
- Do not add dependencies.
- Do not introduce hidden gateway globals or auto-registration.

Files:
- `x/gateway/entrypoints.go`
- `x/gateway/proxy.go`
- `x/gateway/entrypoints_test.go`
- `docs/modules/x-gateway/README.md`

Tests:
- `go test -race -timeout 60s ./x/gateway/...`
- `go test -timeout 20s ./x/gateway/...`
- `go vet ./x/gateway/...`

Docs Sync:
Update `docs/modules/x-gateway/README.md` to point dynamic or user-provided
configuration paths at the error-returning app-facing constructor.

Done Definition:
- `RegisterProxy` returns invalid gateway config errors without panicking.
- The app-facing constructor surface documents the safe and panic-compatible
  paths clearly.
- Existing nil-safe `RegisterRoute` semantics remain either unchanged or
  explicitly documented with updated tests.
- The listed validation commands pass.

Outcome:
- Added `NewGatewayE` as the app-facing error-returning gateway constructor.
- Preserved `NewGateway` as the panic-compatible wrapper over the safe path.
- Updated `RegisterProxy` to keep nil router / empty path no-op behavior while
  returning invalid gateway configuration errors for real route registrations.
- Added tests for safe constructor success/error paths, invalid
  `RegisterProxy` configuration without panic, and nil-router/empty-path no-op
  registration inputs.
- Updated gateway module docs to route dynamic or user-provided configuration
  through `NewGatewayE` or `RegisterProxy`.

Validation:
- `go test -race -timeout 60s ./x/gateway/...`
- `go test -timeout 20s ./x/gateway/...`
- `go vet ./x/gateway/...`
