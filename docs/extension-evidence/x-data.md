# x/data Maturity Evidence

Module: `x/data`

Owner: `persistence`

Current status: `experimental`

Candidate status: not selected

Evidence state: initial triage

## Current Coverage

- `x/data/file` has tests for local storage, S3 behavior, signed URLs, image
  helpers, metadata helpers, and metadata persistence.
- `x/data/idempotency` has KV and SQL provider tests for the stable
  `store/idempotency` contract.
- `x/data/kvengine` covers embedded KV behavior.
- `x/data/rw` covers cluster routing, health, load balancing, and routing
  policy.
- `x/data/sharding` covers routing, strategies, parser, resolver, rewriter,
  config, logging, metrics, and tracing behavior.

## Boundary State

- Primer: `docs/modules/x-data/README.md`
- Manifest: `x/data/module.yaml`
- Boundary state: topology-heavy behavior is documented as extension-owned and
  kept outside stable `store`.

## Why This Is Not A Beta Candidate Yet

`x/data` is a broad umbrella over several topology-heavy capabilities. A single
root-level beta promise would mix file metadata, idempotency providers, embedded
KV, read/write routing, and sharding contracts before their compatibility
surfaces are segmented.

## Next Evidence Needed

- Split feature-level API inventories for `file`, `idempotency`, `kvengine`,
  `rw`, and `sharding`.
- Identify which subpackages are beta candidates and which remain
  experimental.
- Generate API snapshots for the selected candidate surfaces.
- Record release-history evidence after candidate surfaces are narrowed.

## Current Decision

Keep `x/data` experimental.
