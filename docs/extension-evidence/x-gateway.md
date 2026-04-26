# x/gateway Beta Evidence

Module: `x/gateway`

Owner: `edge`

Current status: `experimental`

Candidate status: `beta`

Evidence state: incomplete

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

Missing. Generate snapshots with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/gateway/... -out /tmp/plumego-x-gateway-api.snapshot
```

Snapshot refs:

- none recorded

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

## Owner Sign-Off

Missing. The `edge` owner must confirm the beta criteria before any
`module.yaml` status change.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/gateway` remains `experimental`.
