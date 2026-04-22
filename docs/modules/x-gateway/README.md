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

## Boundary rules

- `x/gateway` is the only app-facing surface for edge transport work; do not duplicate gateway routing logic in stable `router` or stable `middleware`
- keep circuit breaker, retry, and balancer state instance-scoped; do not introduce package-level globals or implicit registration at import time
- keep proxy rewrite, transform, and cache adapters contained within `x/gateway/*` subpackages; do not push edge-transport policy into stable roots
- do not couple discovery (`x/discovery`) selection to gateway-only defaults; discovery backend choice belongs to the caller's wiring
- keep `RegisterRoute` and `RegisterProxy` nil-safe and no-op for invalid args; callers are responsible for providing valid values when routing is needed

## Current test coverage

- `newBackendCircuitBreaker`: nil-config defaults (closed state), explicit config, Trip/Reset lifecycle
- `NewGateway`, `NewGatewayBackendPool` (valid URLs, invalid URL error), `NewGatewayProtocolRegistry`
- `RegisterRoute`: valid wiring (route reachable), nil router no-op, empty path no-op, nil handler no-op
- `RegisterProxy`: valid proxy wiring with live test server
- balancer, backend, health, proxy, rewrite, transform, cache, and protocolmw subpackages each have dedicated test files

## Canonical change shape

- keep gateway behavior explicit and adapter-local
- use `x/gateway` as the only app-facing entrypoint for edge transport work
- route reusable resource-interface work to `x/rest`
