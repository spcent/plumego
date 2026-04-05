# x/gateway

## Purpose

`x/gateway` is the only app-facing extension surface for gateway and edge transport work.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is reverse proxy behavior
- the task is gateway routing, rewriting, or balancing
- edge transport adapters or gateway-local health behavior are involved

## Do not use this module for

- application bootstrap
- reusable resource-interface conventions
- business-specific gateway policy hidden in shared helpers

## First files to read

- `x/gateway/module.yaml`
- `x/gateway/entrypoints.go`
- `specs/extension-taxonomy.yaml`

## Main risks when changing this module

- proxy behavior regression
- transport determinism regression
- hidden global state in adapters

## Canonical change shape

- keep gateway behavior explicit and adapter-local
- use `x/gateway` as the only app-facing entrypoint for edge transport work
- route reusable resource-interface work to `x/rest`
