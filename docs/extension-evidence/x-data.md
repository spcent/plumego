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

## Candidate Surface Inventory

| Surface | Package | Current decision | Why | Next blocker |
| --- | --- | --- | --- | --- |
| File storage | `x/data/file` | Likely beta candidate after inventory | Clear storage and metadata interfaces with local and S3 helper coverage | Freeze exported storage, metadata, signer, and image helper API snapshot |
| Idempotency providers | `x/data/idempotency` | Likely beta candidate after inventory | Adapts stable `store/idempotency` contracts through KV and SQL providers | Confirm provider constructors and config structs across release refs |
| Embedded KV engine | `x/data/kvengine` | Experimental | Owns persistence format, WAL, snapshots, metrics hooks, eviction, and serializer behavior | Separate operational compatibility promise from in-process API promise |
| Read/write routing | `x/data/rw` | Experimental | Topology, health, load balancing, and routing policy are still broad | Define the minimal app-facing router and health API before snapshotting |
| Sharding | `x/data/sharding` | Experimental | Parser, resolver, rewriter, config, metrics, and tracing surface is large | Split config, routing strategy, and SQL rewrite inventories before promotion |

Do not promote the root `x/data` package from this inventory. Promotion work
should select one surface, snapshot only that surface, and then prove release
stability with `internal/checks/extension-release-evidence`.

## Next Evidence Needed

- Complete exported API inventories for the likely beta candidates:
  `x/data/file` and `x/data/idempotency`.
- Generate API snapshots for the selected candidate surfaces.
- Record release-history evidence after candidate surfaces are narrowed.
- Keep `kvengine`, `rw`, and `sharding` experimental until their operational
  compatibility promises are narrower.

## Current Decision

Keep `x/data` experimental.
