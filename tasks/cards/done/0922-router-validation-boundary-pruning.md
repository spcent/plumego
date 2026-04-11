# Card 0922: Router Validation Boundary Pruning

Priority: P1
State: done
Primary Module: router

## Goal

Keep stable `router` focused on route structure: method/path matching, params, groups, reverse routing, and static mount primitives. Remove route parameter validation ownership from `router` so validation policy lives with the handler, application, or owning extension.

## Problem

`router/validator.go` still exposes a route validation subsystem:

- `ParamValidator`
- `RouteValidation`
- `NewRouteValidation`
- `(*RouteValidation).AddParam`
- `(*RouteValidation).Validate`
- `WithValidation`
- `WithValidationRule`
- `(*Router).AddValidation`

This is inconsistent with the current `router/module.yaml` boundary:

- `router` non-goals include business validation.
- Route registration should stay one method + one path + one handler.
- Validation should be visible at the handler or owning transport layer, not hidden in route lookup.

It also creates two ways to validate route params:

- Canonical handler-level validation using `router.Param(r, "id")` or `contract.RequestContextFromContext`.
- Router-owned validation callbacks hidden behind registration-time side tables.

## Scope

- Enumerate all references to the exported validation symbols before editing.
- Remove router-owned validation public API and internal state.
- Remove validation execution from the matcher/dispatch path.
- Update router tests that assert validation registration or dispatch behavior.
- Keep route param extraction behavior unchanged.
- Update router docs and manifest to state that route param validation belongs in handlers or extension/resource layers.

## Non-Goals

- Do not change route matching semantics.
- Do not change route params stored in `contract.RequestContext`.
- Do not add a replacement validation package in stable roots.
- Do not introduce middleware-driven DTO binding or hidden validation.

## Expected Files

- `router/validator.go`
- `router/router.go`
- `router/dispatch.go`
- `router/*_test.go`
- `docs/modules/router/README.md`
- `router/module.yaml`

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./router ./core
go test -race -timeout 60s ./router
go vet ./router ./core
```

Then run the required repo-wide gates before committing.

## Done Definition

- `router` no longer exports validation-specific symbols.
- The old validation symbols have zero residual references.
- Route matching, params, groups, reverse routing, and method-not-allowed tests still pass.
- Router docs/manifests describe validation as outside router ownership.
- Focused gates and repo-wide gates pass.

## Outcome

- Removed router validation types and registration paths (`RouteValidation`, `ParamValidator`, `AddValidation`, and related options).
- Dropped validation checks from dispatch/matcher paths and deleted validation tests/benchmarks.
- Updated router docs and module manifest to keep validation policy outside router ownership.
