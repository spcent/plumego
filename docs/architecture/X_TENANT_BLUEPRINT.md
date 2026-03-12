# X/Tenant Blueprint

`x/tenant` is the target home for all tenant-specific capability code.

## Scope

`x/tenant` owns:

- tenant resolution
- tenant quota
- tenant policy
- tenant rate limiting
- tenant-aware store adapters
- tenant-specific transport mapping

`x/tenant` does not own:

- core app bootstrap
- generic middleware registry
- generic store abstractions
- business onboarding workflows

## Target Package Layout

```text
x/tenant/
  resolve/
  policy/
  quota/
  ratelimit/
  config/
  transport/
  store/db/
  store/cache/
```

## Migration Rules

- New tenant middleware should be added under `x/tenant/*`.
- Stable `middleware` must not gain more tenant-specific behavior.
- Stable `store` must not gain more tenant-aware adapters.
- `core` must not gain tenant-specific component wiring.
