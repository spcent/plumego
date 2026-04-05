# x/rest

## Purpose

`x/rest` is Plumego's app-facing reusable resource-interface layer for CRUD route standardization, query parsing, pagination rules, and repository-backed resource wiring.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

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
- do not introduce a separate `x/rest`-specific response envelope family
- treat `x/rest` as reusable controller and route-binding infrastructure, not as a replacement transport contract layer

## Agent guidance

- If the task is "standardize CRUD/resource API shape", start here
- If the task is "proxy/edge transport", start in `x/gateway`
- If the task is "bootstrap a service", start in `reference/standard-service`
- If the task is "domain validation or business rules", keep that logic outside `x/rest`
