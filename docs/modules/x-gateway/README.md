# x/gateway

## Purpose

`x/gateway` is the only app-facing extension surface for gateway and edge transport work.

## v1 Status

- `Experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen
- Beta candidate once the extension stability policy's two-release API freeze
  evidence is available. Current blocker: no repository release history proves
  two consecutive minor releases without exported `x/gateway/*` API changes.

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
- keep `RegisterRoute` and `RegisterProxy` explicit about invalid args: nil routers, blank paths, and nil handlers must return errors instead of hiding app wiring mistakes

## Current test coverage

- `newBackendCircuitBreaker`: nil-config defaults (closed state), explicit config, Trip/Reset lifecycle
- `NewGateway`, `NewGatewayE`, `NewGatewayBackendPool` (valid URLs, invalid URL error), `NewGatewayProtocolRegistry`
- `RegisterRoute`: valid wiring (route reachable), nil router error, blank path error, nil handler error
- `RegisterProxy`: valid proxy wiring with live test server, nil router error, blank path error
- balancer, backend, health, proxy, rewrite, transform, cache, and protocolmw subpackages each have dedicated test files

## Runnable edge example

`x/gateway/example_test.go` shows the recommended app-facing path: create a
router in app-local wiring, call `RegisterProxy` with an explicit backend target,
and handle invalid dynamic configuration as an error. Discovery remains
caller-owned; the example does not install a discovery default.

Core app examples should use the current explicit lifecycle: build from
`core.DefaultConfig()`, call `core.New(cfg, core.AppDependencies{...})`, register
gateway handlers through `app.Any` or `app.AddRoute`, then prepare and serve via
`Prepare` and `Server`. Do not use removed raw router escape hatches in examples.

## Beta readiness

`x/gateway` satisfies the current coverage and boundary portions of
`docs/EXTENSION_STABILITY_POLICY.md`: gateway construction, backend pools,
proxy registration, route registration, circuit-breaker lifecycle, balancer,
backend, health, rewrite, transform, cache, and protocol middleware behavior
have focused tests.

The module remains `experimental` until the release-history criterion is
verifiable. Promotion to `beta` requires evidence that exported `x/gateway/*`
symbols have not changed for two consecutive minor releases, plus owner
sign-off recorded with the promotion card. Discovery backend selection remains
caller-owned and must not become a gateway default.

## Canonical change shape

- keep gateway behavior explicit and adapter-local
- use `x/gateway` as the only app-facing entrypoint for edge transport work
- use `NewGatewayE` or `RegisterProxy` for dynamic or user-provided gateway
  configuration so invalid targets return errors instead of panicking; reserve
  `NewGateway` for panic-compatible static wiring
- route reusable resource-interface work to `x/rest`
- protocol middleware default error responses use stable codes and safe stage details; raw adapter, transform, executor, encoder, and read errors remain available only to caller-provided error hooks
