# Plumego Roadmap

This roadmap reflects the current repository direction after the agent-first restructuring work. It is architecture-first and optimized for low-regret iteration.

## Current Position

Plumego now has:

- stable roots with narrow responsibilities
- a canonical application path in `reference/standard-service`
- explicit `x/*` extension discovery rules
- machine-readable workflow metadata under `specs/*`
- removal of compatibility-style component APIs from `core` and extension entrypoints

The next stages should harden this shape instead of adding new generic abstraction layers.

## Roadmap Principles

- Keep `core` as a kernel, not a feature catalog.
- Preserve `net/http` compatibility.
- Prefer explicit app-local wiring over hidden registration.
- Maintain one canonical entrypoint per capability family.
- Bias toward machine-readable repo rules so agents can work with low ambiguity.
- Reduce stable-root migration debt before adding new broad capabilities.

## Phase 1: Finish the Agent Workflow Control Plane

Status: complete

Goals:

- make task discovery deterministic
- make module ownership and validation mechanically discoverable
- make repo-native task slicing part of normal development

Planned work:

- add `tasks/cards/*.md` for reversible work items
- add a checker that verifies every manifest `doc_paths` target exists
- add a checker that verifies `specs/change-recipes/*` remain discoverable from canonical repo metadata
- expand `specs/agent-entrypoints.yaml` to cover more task families at finer granularity
- keep `docs/modules/*` synchronized with module manifests

Exit criteria:

- agents can choose a start path without reading broad prose first
- manifests, primers, ownership data, and recipes stay in sync under checks
- new contributors can follow the same workflow without tribal knowledge

Execution surface:

- `tasks/cards/` is the repo-native queue for reversible work items
- task cards should stay small enough for one focused implementation pass
- task cards are workflow assets, not archival prose

## Phase 2: Complete Extension Taxonomy Convergence

Status: substantially complete

Goals:

- remove ambiguity between sibling extension packages
- keep one obvious app-facing surface per capability family
- keep primitive packages available without turning them into discovery traps

Planned work:

- keep `x/messaging` as the only app-facing messaging family entrypoint
- keep `x/mq`, `x/pubsub`, `x/webhook`, and `x/scheduler` explicitly subordinate
- keep `x/gateway` and `x/rest` as non-overlapping entrypoints
- keep `x/ops` and `x/observability` as non-overlapping discovery roots
- keep `x/frontend`, `x/devtools`, and `x/discovery` clearly documented as secondary capability roots, not bootstrap surfaces

Completed in this phase:

- converged messaging-family discovery onto `x/messaging`
- clarified subordinate status for `x/mq`, `x/pubsub`, and `x/webhook`
- separated `x/gateway` from `x/rest` at the metadata and primer level
- separated `x/ops` from `x/observability` at the metadata and primer level
- completed secondary discovery guidance for `x/frontend`, `x/devtools`, and `x/discovery`

Remaining work:

- keep `x/scheduler` explicitly subordinate in family-level documentation
- watch for taxonomy drift as new extensions or adapters are added
- add stronger checks if future ambiguity reappears in manifests or top-level docs

Exit criteria:

- app-facing docs never send users to primitive sibling packages first
- capability family entrypoints are consistent across docs, manifests, and examples
- extension naming no longer forces agents to guess where to start

## Phase 3: Harden `x/rest` as the Reusable Resource Interface Layer

Status: substantially complete

Goals:

- make `x/rest` the reusable home for resource-interface conventions
- preserve explicit route binding while maximizing CRUD reuse
- keep `x/rest` distinct from bootstrap and proxy topology concerns

Planned work:

- keep route-level tests around `RegisterContextResourceRoutes(...)` current as the public registration surface evolves
- keep examples showing `ResourceSpec -> repository -> routes` current as the canonical reuse path
- continue moving query, pagination, hooks, and transformer behavior toward spec-driven configuration
- keep reusable error and response conventions aligned with `contract`
- keep layering guidance between handlers, repositories, and `x/rest` controllers explicit

Completed in this phase:

- locked down the public route registration surface with focused `x/rest` route tests
- added a canonical `x/rest` example for `ResourceSpec -> NewDBResource -> RegisterContextResourceRoutes(...)`
- tightened `ApplyResourceSpec(...)` so controller defaults flow through one orchestration path
- clarified `x/rest` response guidance and the layering boundary between app wiring, controllers, repositories, and domain logic

Remaining work:

- deepen spec-driven orchestration only when new duplication or ambiguity appears
- add further examples if new reusable controller patterns emerge
- keep `x/rest` guidance synchronized as the public API evolves

Guidance constraints:

- keep response and error conventions aligned with `contract`
- do not turn `x/rest` into a bootstrap or transport-contract replacement layer

Execution approach:

- land route-level tests first to lock the public registration surface
- add examples before making deeper orchestration changes
- keep each `x/rest` hardening step reversible and independently testable

Exit criteria:

- resource APIs can be standardized without inventing per-service scaffolding
- `x/rest` stays clearly distinct from `reference/standard-service` and `x/gateway`
- spec-driven behavior is covered by focused tests and examples

## Phase 4: Shrink Stable-Root Migration Debt

Status: complete

Goals:

- keep stable roots narrow and durable
- move topology-heavy or feature-heavy logic out of stable packages
- align package placement with the architecture blueprint

Completed work:

- migrated `store/cache/distributed` → `x/cache/distributed` (consistent-hashing distributed cache)
- migrated `store/cache/redis` → `x/cache/redis` (Redis adapter)
- `store/cache` now contains only abstract types, interfaces, and in-memory implementations
- audited stable `store` topology debt; established rule that no new topology-heavy packages may be added under stable `store`
- confirmed transport health endpoints remain outside `health`; `health` owns models and readiness state only
- tightened boundary documentation blocking tenant leakage into stable `middleware` and `store`
- reduced observability catch-all drift in stable `middleware`; stable middleware owns transport primitives, `x/observability` owns adapters and export wiring
- all `specs/check-baseline` migration debt files removed; checks now enforce boundaries without open exceptions

Exit criteria met:

- stable roots read as long-lived primitives rather than convenience catalogs
- migration debt files have been removed; checks block new drift instead of documenting it
- architecture checks pass cleanly with no baseline suppression

## Phase 5: Reference and Scaffold System

Status: substantially complete

Goals:

- make the canonical app path easy to copy without reintroducing hidden patterns
- ensure scaffolds follow the same rules as the reference app

Completed work:

- `reference/standard-service` is the single canonical application layout demonstrating
  explicit route registration, constructor-based wiring, and stdlib-only dependencies
- `cmd/plumego new` scaffold command generates projects from templates (`minimal`, `api`,
  `fullstack`, `microservice`) each based on the canonical `reference/standard-service` structure
- `cmd/plumego generate` code generation command produces handlers, middleware, and related
  boilerplate aligned with the canonical style
- scaffold templates follow the same explicit wiring rules as the canonical reference; no hidden
  registration or global init patterns are introduced

Remaining work:

- add reference variants outside the canonical path for `x/messaging`, `x/gateway`,
  `x/websocket`, and `x/webhook` — feature demos should be clearly non-canonical
- add a check that canonical references do not drift into `x/*`

Exit criteria met:

- new projects start from the same explicit structure via `cmd/plumego new`
- scaffold output follows the canonical style and is aligned with docs and specs

Remaining exit criteria:

- feature demos do not pollute the canonical learning path (reference variants still needed)

## Phase 6: Release Readiness Toward v1

Status: substantially complete

Goals:

- turn the architecture cleanup into a durable release baseline
- make quality gates strict enough for stable public adoption

Completed work:

- audited all stable-root public API surfaces; removed implementation details that leaked into
  the exported API before v1 freeze:
  - `router`: unexported `CacheEntry`, `PatternCacheEntry`, `RouteMatcher`, `NewRouteMatcher`,
    `IsParameterized` — internal implementation details with no external callers
  - `metrics`: removed `MetricsMiddleware` and `MetricsHandler` — middleware helpers that
    violated the module boundary (use `middleware/httpmetrics.Middleware` instead)
- expanded negative-path coverage for critical roots:
  - `contract`: `WriteBindError` is now tested against all sentinel errors with HTTP status and
    error code assertions; field-level validation errors are also covered
  - `router`: frozen-router registration, duplicate route, param validation failure, unknown path,
    and double-slash path are all covered with negative assertions
- defined a formal deprecation policy in `docs/DEPRECATION.md` covering the compatibility
  promise, four-step deprecation process, extension package exemption, and governance rules
- all quality gates pass (`go test -race ./...`, `go vet ./...`, all `internal/checks/*`)

Remaining work:

- Phase 5 scaffold work is substantially complete; the remaining Phase 5 item (non-canonical
  extension reference variants) does not block the API freeze
- keep `x/*` extension packages aligned with stable-root changes as the canonical reference
  evolves

Exit criteria met:

- public docs describe only the supported explicit APIs
- quality gates are green without new temporary exceptions
- maintainers can describe the supported architecture without caveats

See `docs/DEPRECATION.md` for the formal extension evolution policy.

## Cross-Cutting Workstreams

### Documentation

- keep `README.md`, `README_CN.md`, `AGENTS.md`, and module primers synchronized
- ensure examples use explicit route registration and constructor-based wiring
- avoid documenting speculative abstractions before code exists

### Testing

- preserve required repo-wide gates in `AGENTS.md`
- keep targeted tests near changed behavior
- bias toward negative-path coverage for `security`, `tenant`, and `webhook`
- add smoke tests around canonical references and resource registration helpers

### Tooling

- keep checks fast enough for routine iteration
- extend repo checks only when they reduce ambiguity or drift
- prefer machine-readable rules over tribal knowledge

## Suggested Execution Order

1. Finish agent workflow checks and task cards.
2. Complete extension taxonomy convergence.
3. Deepen `x/rest` examples and route coverage.
4. Reduce stable-root migration debt.
5. Build canonical scaffolds and feature references.
6. Run formal release-readiness work toward v1.

## What Not to Do

- do not reintroduce component-style compatibility APIs
- do not add new broad legacy roots for short-term convenience
- do not let feature demos become the canonical app path
- do not push tenant or topology-heavy logic back into stable roots
- do not let docs get ahead of code behavior
