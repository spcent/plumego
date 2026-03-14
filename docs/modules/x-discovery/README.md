# x/discovery

## Purpose

`x/discovery` provides service discovery primitives and adapters for gateway and client integration.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is service lookup or resolver behavior
- the task is endpoint discovery for clients or gateways
- dynamic instance discovery is required

## Do not use this module for

- application bootstrap
- transport-specific proxy policy
- stable-root durable primitives before the design is proven

## First files to read

- `x/discovery/module.yaml`
- `specs/agent-entrypoints.yaml`
- `specs/extension-entrypoints.yaml`

## Main risks when changing this module

- hidden background state
- nondeterministic resolution behavior
- transport concerns leaking into discovery internals

## Canonical change shape

- keep discovery behavior explicit
- keep resolvers small and composable
- do not spread discovery concerns across stable roots

## Boundary with bootstrap

- `x/discovery` is a secondary capability root for resolver and discovery work, not an application bootstrap surface
- keep discovery concerns out of stable roots unless they become durable primitives
- keep service discovery wiring explicit in the owning application or extension
