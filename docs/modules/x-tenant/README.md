# X/Tenant Module

> **Package**: `github.com/spcent/plumego/x/tenant`  
> **Stability**: Experimental

`x/tenant` is the target extension boundary for Plumego multi-tenancy.

## Responsibilities

- tenant resolution
- tenant quota
- tenant policy
- tenant rate limiting
- tenant-aware store adapters

## Non-Goals

- core bootstrap behavior
- generic middleware registry behavior
- generic store ownership

## Target Subpackages

- `resolve`
- `policy`
- `quota`
- `ratelimit`
- `config`
- `transport`
- `store/db`
- `store/cache`

## Architectural Rule

Tenant functionality is opt-in. Stable packages must not depend on `x/tenant`.
