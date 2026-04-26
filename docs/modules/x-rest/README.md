# x/rest

## Purpose

`x/rest` is Plumego's app-facing reusable resource-interface layer for CRUD route standardization, query parsing, pagination rules, and repository-backed resource wiring.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen
- Beta candidate once the extension stability policy's two-release API freeze
  evidence is available. Current blocker: no repository release history proves
  two consecutive minor releases without exported `x/rest` API changes.

## Use this module when

- the task is standardizing reusable resource APIs
- the change is controller-level CRUD transport behavior
- the change is shared query, pagination, hook, or transformer behavior for resource endpoints

## Do not use this module for

- application bootstrap
- gateway or reverse-proxy topology
- business repository ownership
- domain validation that belongs in the owning package

## First files to read

- `x/rest/module.yaml`
- `x/rest/spec.go`
- `x/rest/entrypoints.go`
- `reference/standard-service` when checking canonical bootstrap shape

## Public entrypoints

- `ResourceSpec`
- `RegisterResourceRoutes`
- `RegisterContextResourceRoutes`
- `NewDBResource`
- `RegisterDBResource`

## Main risks when changing this module

- non-canonical resource route registration
- response or error drift away from `contract`
- hidden repository injection or implicit bootstrap behavior
- controller defaults diverging across services

## Canonical change shape

- keep route binding explicit in app-local wiring
- keep reusable CRUD transport behavior in `x/rest`
- keep response and error conventions aligned with `contract`
- keep domain validation and business rules outside `x/rest`

## Recommended layering

1. Define a `ResourceSpec`
2. Build a repository implementation
3. Create a `DBResourceController` or custom context-aware controller
4. Register routes explicitly in an app-local wiring package

Recommended ownership split:

- app-local wiring: owns route binding and transport composition
- `x/rest`: owns reusable CRUD transport behavior
- repository: owns persistence and query execution
- domain package: owns business validation and business rules

## Canonical example

```go
package users

import (
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/x/rest"
)

func RegisterRoutes(r *router.Router, repository rest.Repository[User]) {
	spec := rest.DefaultResourceSpec("users").WithPrefix("/api/users")

	controller := rest.NewDBResource(spec, repository)
	rest.RegisterContextResourceRoutes(r, spec.Prefix, controller)
}
```

## Response and error guidance

- keep using `contract.WriteError` for structured error responses
- keep using contract-based response helpers for transport output
- default resource controllers should write errors directly through `contract`, not through `x/rest`-local response wrappers
- `DBResourceController` default error responses use stable contract codes and safe public messages; hook, repository, and transformer internals are not exposed through default client-facing messages
- do not introduce a separate `x/rest`-specific response envelope family
- treat `x/rest` as reusable controller and route-binding infrastructure, not as a replacement transport contract layer

## Boundary rules

- `x/rest` is a transport-layer library; it does not own persistence, domain validation, or application bootstrap
- keep route binding explicit in app-local wiring; do not register routes automatically at import time
- keep response and error conventions aligned with `contract`; do not introduce `x/rest`-local error envelopes
- keep domain validation and business rules outside `x/rest`
- `RegisterResourceRoutes` accepts nil router or controller without panicking — callers are responsible for providing non-nil args when routes are needed

## Current test coverage

- `RegisterResourceRoutes`: canonical route surface (all 11 routes), `RouteOptions` selective enable/disable, `BaseContextResourceController` wiring, nil router and nil controller no-op, empty prefix normalization
- `DefaultRouteOptions`: all three flags enabled
- `BaseResourceController`: all 7 not-implemented methods return HTTP 501 with `contract.CodeNotImplemented`, `Options` sets CORS headers and returns 204, `Head` returns 200, empty `ResourceName` defaults to `"resource"`
- `ResourceSpec` / `ApplyResourceSpec`: controller defaults preservation, spec-driven query normalization, legacy sort field filtering
- `NewPaginationMeta`: first-page (HasPrev=false, HasNext=true), last-page (HasNext=false), zero-item (TotalPages=0, no navigation)
- `QueryBuilder`: page-size clamping to max, invalid page input uses default (page=1), page=0 treated as default, unknown sort field filtered out, unknown filter field filtered out, descending sort prefix (`-`) parsing

## Beta readiness

`x/rest` satisfies the current coverage and boundary portions of
`docs/EXTENSION_STABILITY_POLICY.md`: documented route registration,
controller defaults, query parsing, pagination, nil-argument behavior, and
not-implemented negative paths have focused tests.

The module remains `experimental` until the release-history criterion is
verifiable. Promotion to `beta` requires evidence that exported `x/rest`
symbols have not changed for two consecutive minor releases, plus owner
sign-off recorded with the promotion card.

## Agent guidance

- If the task is "standardize CRUD/resource API shape", start here
- If the task is "proxy/edge transport", start in `x/gateway`
- If the task is "bootstrap a service", start in `reference/standard-service`
- If the task is "domain validation or business rules", keep that logic outside `x/rest`
