# x/data/cache Maturity Evidence

Module: `x/data/cache`

Owner: `platform`

Current status: `experimental`

Candidate status: `not selected`

Evidence state: surface inventory

## Current Coverage

- `x/data/cache/distributed` covers hash-ring routing, node management, replication,
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
  callbacks. It also gives synchronous `Set` the same primary-first ordering as
  `Incr`/`Append` and adds a separately named sync fanout concurrency setting
  with an async-field compatibility fallback.
- `x/data/cache/leaderboard` covers skiplist ordering, sorted-set operations,
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
  the checked-in benchmark entry points for score-range baselines. The ninth
  pass removes the reverse lock ordering between metrics snapshots and
  sorted-set mutation while preserving approximate operational metrics, then
  splits internal metrics counters from the exported metrics snapshot DTO and
  documents custom-config `EnableMetrics` zero-value behavior.
- `x/data/cache/redis` covers minimal adapter operations, key validation, optional
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
  the real-driver integration matrix as a promotion blocker. The ninth pass
  validates effective fields written by custom compatibility options while
  documenting package-provided `With*` options as the future stable frozen
  constructor contract.

## Primer And Boundary State

- Primer: `docs/modules/x/data/cache/README.md`
- Manifest: `x/data/cache/module.yaml`
- Boundary state: topology-heavy and provider-specific cache behavior remains
  extension-owned and outside stable `store/cache`.

## Why No `beta` Candidate Is Selected Yet

`x/data/cache` still mixes distributed topology, in-process ranked-data behavior, and
provider adapter contracts. Those surfaces need separate release-history
evidence before a single module-level compatibility promise is credible.

## Candidate Surface Inventory

| Surface | Package | Current decision | Why | Next blocker |
| --- | --- | --- | --- | --- |
| Distributed cache | `x/data/cache/distributed` | Experimental | Replication, failover, post-close lifecycle, fail-closed paths, caller byte ownership on reads and writes, primary-first synchronous mutation ordering, separately named sync fanout concurrency, non-mutating constructor config normalization, bounded weighted hash-ring placement, and health-probe fallback/callback isolation guidance are covered; partial synchronous write outcomes are documented; failover retry tuning is configurable; async scheduling has bounded workers, a bounded queue, metrics, close-time drops, and optional drop callbacks, but async replication remains best-effort without durable repair | Record exported API snapshots and decide whether best-effort async failure visibility is stable enough |
| Leaderboard cache | `x/data/cache/leaderboard` | Possible beta candidate after inventory | In-process sorted-set behavior has focused correctness, lifecycle, limits, explicit no-expiration TTL, instance-owned skiplist randomness, Plumego-local missing-key, validation, clean exported metrics snapshots, deadlock-safe approximate metrics, eventually-expired cleanup semantics, and score-range baseline coverage | Snapshot the exported sorted-set API and decide whether the bounded in-process range baseline is acceptable |
| Redis adapter | `x/data/cache/redis` | Experimental | New option-based call sites have constructor-owned behavior, a validation-capable constructor, adapter byte ownership, capability reporting, cache-miss mapping tests, compatibility-field boundaries, frozen clear policy tests, effective custom-option validation, and a dependency-free compatibility matrix, but no concrete driver integration evidence is recorded | Validate at least one real Redis driver binding outside the dependency-free adapter package |

Do not promote `x/data/cache` as a root module from this inventory. Promotion work
should select one surface, snapshot only that surface, and then prove release
stability with `internal/checks/extension-release-evidence`.

## Required Release Evidence

Missing. No selected `x/data/cache` surface has two consecutive minor release refs
with unchanged exported API.

Release refs:

- none recorded

## API Snapshot Evidence

Current-head snapshots are recorded for the candidate surfaces below. They are
useful development baselines, but they are not release evidence and do not
clear `api_snapshot_missing` by themselves.

- `docs/extension-evidence/snapshots/x-cache/x-cache-distributed-head.snapshot`
- `docs/extension-evidence/snapshots/x-cache/x-cache-leaderboard-head.snapshot`
- `docs/extension-evidence/snapshots/x-cache/x-cache-redis-head.snapshot`

## Release Evidence

Not recorded. `specs/extension-beta-evidence.yaml` does not name a selected
`x/data/cache` release candidate yet, and no real release-ref comparison has been
checked in.

Current state:

- Selected release candidate: none
- API snapshot comparison: current-head only
- Release-history comparison: not recorded

## Owner Sign-Off

Missing. No selected `x/data/cache` surface has owner sign-off recorded.

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
- Keep `x/data/cache/module.yaml` status as `experimental` until the promotion
  process in `docs/EXTENSION_STABILITY_POLICY.md` is complete.

## Ninth Stabilization Pass Validation

- `go test -race -timeout 60s ./x/data/cache/...`
- `go test -timeout 20s ./x/data/cache/...`
- `go vet ./x/data/cache/...`
- `go test -race -timeout 60s ./x/data/cache/distributed`
- `go test -race -timeout 60s ./x/data/cache/leaderboard`
- `go test -race -timeout 60s ./x/data/cache/redis`
- `go test -timeout 20s ./...`
- `go vet ./...`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/dependency-rules`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/agent-workflow`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/module-manifests`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/reference-layout`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/data/cache/distributed -out docs/extension-evidence/snapshots/x-cache/x-cache-distributed-head.snapshot`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/data/cache/leaderboard -out docs/extension-evidence/snapshots/x-cache/x-cache-leaderboard-head.snapshot`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/data/cache/redis -out docs/extension-evidence/snapshots/x-cache/x-cache-redis-head.snapshot`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-release-evidence -module ./x/data/cache/... -base HEAD -head HEAD`
  verified the release-evidence tool path only; it is not release-history
  evidence because no real release refs were supplied.

## Blockers

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

## Promotion Posture

Keep `x/data/cache` experimental.
