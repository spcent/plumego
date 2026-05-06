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
  behavior, and bounded async replication contexts. The fourth pass adds
  primary-only `ReplicationNone` selection, `Exists` failover coverage, and
  explicit partial-write coverage for synchronous `Incr`/`Append` secondary
  failures. The fifth pass adds async replication scheduling concurrency bounds,
  exhausted-scheduler failure metrics, and bounded hash-ring placement failure
  with rollback. The sixth pass clarifies synchronous replication as
  best-effort rather than strong consistency, documents partial `Set`/`Delete`/
  `Clear` side effects, exposes failover retry attempts/backoff as validated
  configuration, and replaces per-write async goroutine scheduling with a
  bounded worker queue plus optional drop callback. The seventh pass adds an
  explicit post-close `ErrClosed` contract, close-time queued async drop
  callbacks, drop-callback panic recovery, and documented zero-value default
  config semantics. The eighth pass starts by making distributed `Set`/`Append`
  own caller byte slices at node-cache boundaries and by bounding synchronous
  `Set` fanout through the existing replication concurrency setting. It also
  makes hash ring and health checker construction normalize configs without
  mutating caller-owned config objects, caps per-node virtual-node expansion,
  and documents the default health probe as a lightweight fallback. The ninth
  pass starts by copying bytes returned from distributed `Get` primary,
  failover, and retry paths, and by recovering panics from health-check failure
  callbacks.
- `x/cache/leaderboard` covers skiplist ordering, sorted-set operations,
  expiration, metrics, context/key validation, invalid members, and duplicate
  update regressions. The second stabilization pass also covers idempotent
  close, concurrent `MaxLeaderboards` enforcement, invalid range errors, and
  `ZIncrBy` mutation metrics. The third pass adds failed-create cleanup,
  missing-key contract coverage, and actual-removal metrics. The fourth pass
  adds direct key validation and tracked leaderboard-count accounting across
  cleanup and clear paths. The fifth pass adds constructor-local config
  normalization and explicit post-close leaderboard errors. The sixth pass
  records Plumego-local missing-key behavior, approximate metrics snapshots,
  and the score-range linear scan baseline. The seventh pass syncs package
  documentation and records invalid `ZIncrBy` rollback as logical state
  preservation rather than a structural no-op guarantee. The eighth pass adds
  an explicit `NoExpirationTTL` config contract while preserving zero-value
  default TTL behavior, documents expiration as lazy plus cleanup based, moves
  skiplist random-level generation behind each skiplist instance, and records
  the checked-in benchmark entry points for score-range baselines.
- `x/cache/redis` covers minimal adapter operations, key validation, optional
  atomic interfaces, disabled flush behavior, unsupported atomic clients,
  option-based construction, and namespaced clear selection. The third pass adds
  constructor-owned option behavior for new call sites and stable key-error
  wrapping. The fourth pass adds a validation-capable constructor and
  adapter-boundary byte-slice copies. The fifth pass adds append byte-slice
  ownership and adapter capability reporting. The sixth pass records the
  dependency-free production compatibility matrix, cache-miss mapping contract,
  optional capability failure behavior, and destructive clear guidance. The
  seventh pass records compatibility-field concurrency boundaries. The eighth
  pass reinforces option-owned behavior for validated Redis adapters and keeps
  the real-driver integration matrix as a promotion blocker.

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
| Distributed cache | `x/cache/distributed` | Experimental | Replication, failover, post-close lifecycle, fail-closed paths, caller byte ownership on reads and writes, bounded synchronous `Set` fanout, non-mutating constructor config normalization, bounded weighted hash-ring placement, and health-probe fallback/callback isolation guidance are covered; partial synchronous write outcomes are documented; failover retry tuning is configurable; async scheduling has bounded workers, a bounded queue, metrics, close-time drops, and optional drop callbacks, but async replication remains best-effort without durable repair | Record exported API snapshots and decide whether best-effort async failure visibility is stable enough |
| Leaderboard cache | `x/cache/leaderboard` | Possible beta candidate after inventory | In-process sorted-set behavior has focused correctness, lifecycle, limits, explicit no-expiration TTL, instance-owned skiplist randomness, Plumego-local missing-key, validation, approximate metrics, eventually-expired cleanup semantics, and score-range baseline coverage | Snapshot the exported sorted-set API and decide whether the bounded in-process range baseline is acceptable |
| Redis adapter | `x/cache/redis` | Experimental | New option-based call sites have constructor-owned behavior, a validation-capable constructor, adapter byte ownership, capability reporting, cache-miss mapping tests, compatibility-field boundaries, frozen clear policy tests, and a dependency-free compatibility matrix, but no concrete driver integration evidence is recorded | Validate at least one real Redis driver binding outside the dependency-free adapter package |

Do not promote `x/cache` as a root module from this inventory. Promotion work
should select one surface, snapshot only that surface, and then prove release
stability with `internal/checks/extension-release-evidence`.

## Next Evidence Needed

- Compare checked-in current-head API snapshots against real release refs after
  surface selection.
- Record release-history evidence after surface selection.
- Decide whether distributed async replication needs durable repair hooks beyond
  bounded queueing, metrics, and drop callbacks.
- Decide whether leaderboard should remain Plumego-local ranked-data behavior
  or grow a separate Redis-compatible surface.
- Validate at least one concrete Redis driver binding against the documented
  compatibility matrix before treating the adapter as stable.
- Record scale and performance expectations for each selected surface.
- Document owner sign-off for the selected surface.
- Keep `x/cache/module.yaml` status as `experimental` until the promotion
  process in `docs/EXTENSION_STABILITY_POLICY.md` is complete.

## Current Head API Snapshots

These snapshots record the current exported API shape only. They are not
release-history evidence and do not clear the two-release or owner sign-off
promotion blockers.

- `docs/extension-evidence/snapshots/x-cache/x-cache-distributed-head.snapshot`
- `docs/extension-evidence/snapshots/x-cache/x-cache-leaderboard-head.snapshot`
- `docs/extension-evidence/snapshots/x-cache/x-cache-redis-head.snapshot`

## Eighth Stabilization Pass Validation

- `go test -race -timeout 60s ./x/cache/...`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
- `go test -run ^$ -bench 'LeaderboardCache(ZRangeByScore|ZCount|ScoreRangeFullScanBaseline)' -benchtime=100ms ./x/cache/leaderboard`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/distributed -out docs/extension-evidence/snapshots/x-cache/x-cache-distributed-head.snapshot`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/leaderboard -out docs/extension-evidence/snapshots/x-cache/x-cache-leaderboard-head.snapshot`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/redis -out docs/extension-evidence/snapshots/x-cache/x-cache-redis-head.snapshot`

## Remaining Stable Blockers By Surface

- Distributed: async secondary failures have bounded queue/drop visibility but
  no durable repair contract has been selected.
- Leaderboard: current behavior is Plumego-local ranked-data behavior with a
  documented in-process score-range scan baseline; current-head API snapshots
  are checked in, but release refs and matching historical snapshots are still
  missing.
- Redis adapter: dependency-free client interface expectations are documented,
  but no concrete client integration evidence is recorded.
- Release governance: no selected surface has release refs, historical API
  snapshot comparisons, or owner sign-off.

## Current Decision

Keep `x/cache` experimental.
