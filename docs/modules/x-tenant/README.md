# x/tenant

## Purpose

`x/tenant` is the extension boundary for multi-tenant policy, quota, rate limit, resolution, and tenant-aware stores.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is tenant policy or isolation work
- the change is tenant-aware by design

## Do not use this module for

- stable middleware defaults
- stable store defaults
- generic request middleware unrelated to tenant semantics

## First files to read

- `x/tenant/module.yaml`
- the owning subpackage under `x/tenant/*`
- `AGENTS.md` tenant boundary rules
