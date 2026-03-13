# Tenant Module

> **Primary boundary**: `github.com/spcent/plumego/x/tenant`
>  
> **Shared primitives**: `github.com/spcent/plumego/x/tenant/core`

`x/tenant` is Plumego's experimental multi-tenancy extension boundary.

## Current Structure

- shared tenant types, hooks, quota, policy, and rate-limit primitives: `x/tenant/core`
- config manager adapters: `x/tenant/config`
- transport middleware: `x/tenant/resolve`, `x/tenant/quota`, `x/tenant/policy`, `x/tenant/ratelimit`
- tenant-aware store helpers: `x/tenant/store/*`

## Canonical Rule

Tenant functionality is opt-in and lives under `x/tenant/*`. Stable packages must not depend on tenant internals.

## Canonical Wiring Pattern

Use explicit router-group middleware composition and import tenant primitives from `x/tenant/core`.

## Compatibility Note

The historical public `tenant/` root has been removed. Remaining documents in this directory are migration notes and examples for the `x/tenant/*` layout.
