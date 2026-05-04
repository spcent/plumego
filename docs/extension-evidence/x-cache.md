# x/cache Maturity Evidence

Module: `x/cache`

Owner: `platform`

Current status: `experimental`

Candidate status: not selected

Evidence state: stability blocker inventory

## Current Coverage

- `x/cache/distributed` covers hash-ring routing, node management, replication,
  failover strategies, lifecycle idempotency, invalid construction inputs, and
  concurrent operation smoke tests. The second stabilization pass also covers
  hash-collision handling, weighted virtual nodes, replication failure metrics,
  replication lag metrics, and configurable health probes. The third pass adds
  nil-cache node rejection, insufficient-replica errors, fail-closed clear
  behavior, bounded async replication contexts, primary-only
  `ReplicationNone` selection, `Exists` failover coverage, and explicit
  partial-write coverage for synchronous `Incr`/`Append` secondary failures.
- `x/cache/leaderboard` covers skiplist ordering, sorted-set operations,
  expiration, metrics, context/key validation, invalid members, and duplicate
  update regressions. The second stabilization pass also covers idempotent
  close, concurrent `MaxLeaderboards` enforcement, invalid range errors, and
  `ZIncrBy` mutation metrics. The third pass adds failed-create cleanup,
  missing-key contract coverage, and actual-removal metrics.
- `x/cache/redis` covers minimal adapter operations, key validation, optional
  atomic interfaces, disabled flush behavior, unsupported atomic clients,
  option-based construction, and namespaced clear selection. The third pass adds
  constructor-owned option behavior for new call sites and stable key-error
  wrapping.

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
| Distributed cache | `x/cache/distributed` | Experimental | Replication and failover semantics are explicit, fail-closed paths are covered, and partial synchronous write outcomes are documented, but async secondary failures are still metrics-only and best-effort | Record exported API snapshots and decide whether metrics-only async failure visibility is stable enough |
| Leaderboard cache | `x/cache/leaderboard` | Possible beta candidate after inventory | In-process sorted-set behavior has focused correctness, lifecycle, limits, missing-key, and metrics coverage | Snapshot the exported sorted-set API and decide Redis-compatibility scope |
| Redis adapter | `x/cache/redis` | Experimental | New option-based call sites have constructor-owned behavior, but the adapter still depends on caller-provided clients and optional capabilities | Define concrete client compatibility expectations and production clear guidance |

Do not promote `x/cache` as a root module from this inventory. Promotion work
should select one surface, snapshot only that surface, and then prove release
stability with `internal/checks/extension-release-evidence`.

## Next Evidence Needed

- Generate exported API snapshots for selected candidate surfaces.
- Record release-history evidence after surface selection.
- Decide whether distributed async replication needs callback/repair hooks
  beyond metrics.
- Decide whether leaderboard should promise Redis sorted-set compatibility or
  only Plumego-local ranked-data behavior.
- Define at least one concrete Redis client compatibility matrix before
  treating the adapter as stable.
- Record scale and performance expectations for each selected surface.
- Document owner sign-off for the selected surface.
- Keep `x/cache/module.yaml` status as `experimental` until the promotion
  process in `docs/EXTENSION_STABILITY_POLICY.md` is complete.

## Current Decision

Keep `x/cache` experimental.
