# x/discovery

## Purpose

`x/discovery` provides service discovery primitives and adapters for gateway and client integration.

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
