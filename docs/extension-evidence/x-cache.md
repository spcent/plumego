# x/cache Maturity Evidence

Module: `x/cache`

Owner: `platform`

Current status: `experimental`

Candidate status: not selected

Evidence state: stability blocker inventory

## Current Coverage

- `x/cache/distributed` covers hash-ring routing, node management, replication,
  failover strategies, lifecycle idempotency, invalid construction inputs, and
  concurrent operation smoke tests.
- `x/cache/leaderboard` covers skiplist ordering, sorted-set operations,
  expiration, metrics, context/key validation, invalid members, and duplicate
  update regressions.
- `x/cache/redis` covers minimal adapter operations, key validation, optional
  atomic interfaces, disabled flush behavior, and unsupported atomic clients.

## Boundary State

- Primer: `docs/modules/x-cache/README.md`
- Manifest: `x/cache/module.yaml`
- Boundary state: topology-heavy and provider-specific cache behavior remains
  extension-owned and outside stable `store/cache`.

## Why This Is Not A Beta Candidate Yet

`x/cache` still mixes distributed topology, in-process ranked-data behavior, and
provider adapter contracts. Those surfaces need separate release-history
evidence before a single module-level compatibility promise is credible.

## Candidate Surface Inventory

| Surface | Package | Current decision | Why | Next blocker |
| --- | --- | --- | --- | --- |
| Distributed cache | `x/cache/distributed` | Experimental | Replication and failover semantics are now explicit, but async secondary failures are best-effort | Record exported API snapshots and decide whether async replication error visibility is required |
| Leaderboard cache | `x/cache/leaderboard` | Possible beta candidate after inventory | In-process sorted-set behavior has focused correctness coverage | Snapshot the exported sorted-set API and decide Redis-compatibility scope |
| Redis adapter | `x/cache/redis` | Experimental | Adapter depends on caller-provided clients and optional atomic interfaces | Define concrete client compatibility expectations and namespaced clear guidance |

Do not promote `x/cache` as a root module from this inventory. Promotion work
should select one surface, snapshot only that surface, and then prove release
stability with `internal/checks/extension-release-evidence`.

## Next Evidence Needed

- Generate exported API snapshots for selected candidate surfaces.
- Record release-history evidence after surface selection.
- Decide whether distributed async replication needs observable error hooks.
- Document owner sign-off for the selected surface.
- Keep `x/cache/module.yaml` status as `experimental` until the promotion
  process in `docs/EXTENSION_STABILITY_POLICY.md` is complete.

## Current Decision

Keep `x/cache` experimental.
