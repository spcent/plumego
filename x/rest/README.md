# x/rest

`x/rest` is Plumego's reusable resource-interface layer and the only app-facing entrypoint for reusable resource API standardization.

It is the place for:

- shared CRUD controller patterns
- query parsing and pagination rules
- repository-backed resource wiring
- reusable resource hooks and transformers
- canonical route registration for resource APIs

It is not the place for:

- application bootstrap
- gateway or reverse-proxy topology
- edge transport wiring
- business repository ownership

## Role in the repo

Use these entrypoints by intent:

- app bootstrap: `reference/standard-service`
- gateway and edge transport: `x/gateway`
- reusable resource API conventions: `x/rest`

## Canonical building blocks

- `ResourceSpec`: reusable declaration of resource name, prefix, routes, query defaults, hooks, and transformer behavior
- `RegisterResourceRoutes`: bind a plain `ResourceController`
- `RegisterContextResourceRoutes`: bind a `ContextResourceController`
- `NewDBResource`: build a repository-backed controller from one resource spec
- `RegisterDBResource`: build and register a repository-backed controller in one step

## Recommended layering

1. Define a `ResourceSpec`
2. Build a repository implementation
3. Create a `DBResourceController` or custom context-aware controller
4. Register routes explicitly in an app-local wiring package

Recommended ownership split:

- handler or app-local wiring: owns route binding and transport composition
- `x/rest` controller: owns reusable CRUD transport behavior
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

## Reuse rules

- Reuse `ResourceSpec` across services when the resource interface should stay consistent
- Keep transport rules in `x/rest`; keep business validation in the owning domain package
- Keep route registration explicit; do not hide resource binding behind bootstrap side effects
- Prefer one `ResourceSpec` per resource family instead of ad hoc controller-local defaults

## Response and error guidance

- keep using `contract.WriteError` for structured error responses
- keep using contract-based response helpers for transport output
- do not introduce a separate `x/rest`-specific response envelope family
- treat `x/rest` as reusable controller and route-binding infrastructure, not as a replacement transport contract layer

## Extension points

- `ResourceHooks`: before/after list, create, update, delete hooks
- `ResourceTransformer`: response shaping
- `QueryBuilder`: shared filtering, sorting, search, and pagination parsing
- `Repository[T]`: persistence boundary for repository-backed controllers

## Agent guidance

- If the task is "standardize CRUD/resource API shape", start here
- If the task is "proxy/edge transport", start in `x/gateway`
- If the task is "bootstrap a service", start in `reference/standard-service`
- If the task is "domain validation or business rules", keep that logic outside `x/rest`
