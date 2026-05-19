# x/data Maturity Evidence

Module: `x/data`

Owner: `persistence`

Current status: `experimental`

Candidate status: selected surfaces `beta`

Evidence state: selected-surface complete

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

## Primer And Boundary State

- Primer: `docs/modules/x-data/README.md`
- Manifest: `x/data/module.yaml`
- Boundary state: topology-heavy behavior is documented as extension-owned and
  kept outside stable `store`.

## Why No `beta` Candidate Is Selected Yet

`x/data` is a broad umbrella over several topology-heavy capabilities. A single
root-level beta promise would mix file metadata, idempotency providers, embedded
KV, read/write routing, and sharding contracts before their compatibility
surfaces are segmented.

## Candidate Surface Inventory

| Surface | Package | Current decision | Why | Next blocker |
| --- | --- | --- | --- | --- |
| File storage | `x/data/file` | Beta surface at v1.1.0 | Clear storage and metadata interfaces with local and S3 helper coverage | Root `x/data` remains experimental |
| Idempotency providers | `x/data/idempotency` | Beta surface at v1.1.0 | Adapts stable `store/idempotency` contracts through KV and SQL providers | Root `x/data` remains experimental |
| Embedded KV engine | `x/data/kvengine` | Experimental | Owns persistence format, WAL, snapshots, metrics hooks, eviction, and serializer behavior | Separate operational compatibility promise from in-process API promise |
| Read/write routing | `x/data/rw` | Experimental | Topology, health, load balancing, and routing policy are still broad | Define the minimal app-facing router and health API before snapshotting |
| Sharding | `x/data/sharding` | Experimental | Parser, resolver, rewriter, config, metrics, and tracing surface is large | Split config, routing strategy, and SQL rewrite inventories before promotion |

Do not promote the root `x/data` package from this inventory. Promotion work
should select one surface, snapshot only that surface, and then prove release
stability with `internal/checks/extension-release-evidence`.

## Required Release Evidence

Recorded for the selected `x/data/file` and `x/data/idempotency` surfaces. Both
surfaces have two consecutive minor release refs with unchanged exported API.

Release refs:

- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target) for
  `x/data/file`
- `v1.1.0` for `x/data/file`
- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target) for
  `x/data/idempotency`
- `v1.1.0` for `x/data/idempotency`

## API Snapshot Evidence

Release-backed API snapshots are recorded for the selected surfaces below. The
v1 baseline intake artifacts remain useful history, but the v1.0.0 to v1.1.0
snapshots are the promotion evidence.

Snapshot refs:

- `docs/extension-evidence/snapshots/v1-baseline/x-data-file/base.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-data-file/head.snapshot`
- `docs/extension-evidence/snapshots/x-data-file/base.snapshot`
- `docs/extension-evidence/snapshots/x-data-file/head.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-data-idempotency/base.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-data-idempotency/head.snapshot`
- `docs/extension-evidence/snapshots/x-data-idempotency/base.snapshot`
- `docs/extension-evidence/snapshots/x-data-idempotency/head.snapshot`

## Release Evidence

`specs/extension-beta-evidence.yaml` tracks `surface_candidates` for:

- `x/data:file` covering package `x/data/file`
- `x/data:idempotency` covering package `x/data/idempotency`

Both selected surfaces have complete release history, API snapshot evidence, and
owner sign-off. They do not imply root `x/data` promotion.

Current state:

- Selected release candidates: `x/data/file`, `x/data/idempotency`
- API snapshot comparison: `v1.0.0` to `v1.1.0`, unchanged for both selected surfaces
- Release-history comparison: two release refs recorded for both selected surfaces

## Owner Sign-Off

Signed off by `persistence` for v1.1.0:

> I confirm that the selected `x/data/file` and `x/data/idempotency` surfaces
> meet the beta criteria in docs/EXTENSION_STABILITY_POLICY.md and accept the
> beta compatibility obligations for the documented public surfaces.

## Next Evidence Needed

- Keep `kvengine`, `rw`, and `sharding` experimental until their operational
  compatibility promises are narrower.

## Blockers

None for `x/data/file` and `x/data/idempotency`. Root `x/data`, `kvengine`,
`rw`, and `sharding` remain experimental.

## Promotion Posture

Promote only `x/data/file` and `x/data/idempotency` as beta surfaces at v1.1.0.
Keep root `x/data`, `kvengine`, `rw`, and `sharding` experimental.
