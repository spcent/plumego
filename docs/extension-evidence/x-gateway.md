# x/gateway Beta Evidence

Module: `x/gateway`

Owner: `edge`

Current status: `beta`

Evidence state: complete

## Current Coverage

- Gateway construction coverage includes `NewGateway`, `NewGatewayE`,
  `NewGatewayBackendPool`, and protocol registry setup.
- Route and proxy registration coverage includes valid wiring plus nil router,
  empty path, nil handler, invalid target, and no-op behavior.
- Circuit breaker coverage includes nil-config defaults and trip/reset
  lifecycle.
- Balancer, backend, health, proxy, rewrite, transform, cache, and protocol
  middleware packages have dedicated tests.
- Runnable edge proxy behavior is covered by `x/gateway/example_test.go`.

## Primer And Boundary State

- Primer: `docs/modules/x-gateway/README.md`
- Manifest: `x/gateway/module.yaml`
- Boundary state: documented and aligned with keeping gateway/edge transport in
  `x/gateway` and caller-owned discovery backend selection outside defaults.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/gateway/*` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

One current-head baseline snapshot is recorded. It is useful for comparing the
candidate surface during development, but it is not release evidence and does
not clear `api_snapshot_missing` by itself.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/gateway/... -out /tmp/plumego-x-gateway-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/first-batch/x-gateway-head.snapshot`

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/gateway/... \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-gateway-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until the
recorded refs and snapshot files come from real releases.

## Release Evidence

Release refs: `d2c25c3`, `ec70358`

API snapshot comparison:

- Base: `docs/extension-evidence/snapshots/x-gateway/base.snapshot`
- Head: `docs/extension-evidence/snapshots/x-gateway/head.snapshot`
- Result: **API unchanged** across both refs

## Owner Sign-Off

Signed off by `edge` at v0.2.0:

> I confirm that x/gateway meets the beta criteria in
> docs/EXTENSION_STABILITY_POLICY.md and accept the beta compatibility
> obligations for the documented x/gateway public surface.

## Blockers

None. All promotion blockers cleared.

## Promotion Decision

Promoted to `beta` at v0.2.0. API stable across d2c25c3–ec70358.
