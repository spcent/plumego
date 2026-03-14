# X Tenant Blueprint

`x/tenant` is Plumego's extension boundary for tenant-aware behavior.

## Purpose

Use `x/tenant` for:

- tenant resolution
- tenant policy evaluation
- tenant quota enforcement
- tenant rate limiting
- tenant-aware store adapters

Do not use `x/tenant` for:

- core bootstrap ownership
- stable middleware defaults
- stable store defaults
- generic business onboarding flows

## Boundary Rules

- tenant-aware logic must stay out of stable `middleware` and stable `store`
- reference applications must not depend on `x/tenant` by default
- policy errors must fail closed
- tenant middleware remains transport-only

## Canonical Read Path

1. `x/tenant/module.yaml`
2. `docs/modules/x-tenant/README.md`
3. the owning package under `x/tenant/*`

## Typical Package Entry Points

- `x/tenant/resolve`
- `x/tenant/policy`
- `x/tenant/quota`
- `x/tenant/ratelimit`
- `x/tenant/store`

## Design Direction

- keep tenant resolution explicit
- keep policy and quota checks reviewable
- keep tenant-aware adapters subordinate to the owning tenant package
- do not push tenant semantics back into stable roots
