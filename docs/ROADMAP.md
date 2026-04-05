# Plumego Roadmap

This roadmap reflects the current repository direction after the agent-first restructuring work. It is architecture-first and optimized for low-regret iteration.

## Current Position

Plumego now has:

- stable roots with narrow responsibilities
- a canonical application path in `reference/standard-service`
- explicit `x/*` extension discovery rules
- machine-readable workflow metadata under `specs/*`
- removal of compatibility-style component APIs from `core` and extension entrypoints
- a formal deprecation policy for stable roots
- a full scaffold and code generation CLI in `cmd/plumego`
- deep AI gateway capabilities in `x/ai` (experimental)
- multi-tenancy middleware in `x/tenant` (experimental)

The next stages harden and grow the ecosystem rather than restructure it.

## Roadmap Principles

- Keep `core` as a kernel, not a feature catalog.
- Preserve `net/http` compatibility.
- Prefer explicit app-local wiring over hidden registration.
- Maintain one canonical entrypoint per capability family.
- Bias toward machine-readable repo rules so agents can work with low ambiguity.
- Reduce stable-root migration debt before adding new broad capabilities.

---

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
- expand `specs/task-routing.yaml` to cover more task families at finer granularity
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

Status: complete

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

Completed in this phase:

- added `canonical` template to `plumego new` that mirrors `reference/standard-service` exactly
- added `reference/with-messaging` demo: in-process broker wired into the standard-service shape
- added `reference/with-gateway` demo: reverse proxy wired into the standard-service shape
- added `reference/with-websocket` demo: WebSocket server wired into the standard-service shape
- added `reference/with-webhook` demo: inbound webhook receiver wired into the standard-service shape
- each feature reference is clearly marked non-canonical in its README and doc comment
- added `FindReferenceXImports` check in `internal/checks/checkutil` to detect x/* drift in `reference/standard-service`
- enhanced `internal/checks/reference-layout` to enforce the drift check and require feature reference paths
- updated `specs/task-routing.yaml` with a `scaffold` task entrypoint

Exit criteria:

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

---

## Phase 7: CLI Code Generation Quality

Status: complete

Goals:

- eliminate bare `// TODO` placeholders from `cmd/plumego generate` output
- make scaffold-generated code compile and run without manual editing
- align `plumego new` templates with the current `reference/standard-service` structure

Background:

`cmd/plumego/internal/codegen` and `cmd/plumego/internal/scaffold` previously emitted
`// TODO: define service methods` and stub values such as `ID: "TODO"` in generated
handlers, repositories, and services. This confused first-time users and prevented
generated projects from passing `go build ./...` without manual changes.

Completed work:

- replaced handler generation templates with minimal compilable implementations; each
  handler calls its injected service rather than returning stub values
- replaced repository generation with an interface plus in-memory implementation skeleton
- replaced service generation with method signatures that compile without editing
- completed missing files for the `fullstack` and `microservice` scaffold templates
- added focused tests in `codegen` and `scaffold` packages verifying that every generated
  Go file parses cleanly and contains no bare `// TODO` lines
- removed the `generate component` subcommand; `generate handler` is the single entry
  point for generating handler files

Non-goals:

- do not change the structural conventions of generated code (canonical style is preserved)
- do not introduce external template-engine dependencies

Exit criteria met:

- any template produced by `plumego new <name>` passes `go build ./...` without edits
- `plumego generate handler <Name>` produces a handler file with zero compile errors
- no generated file contains a bare `// TODO: ...` placeholder line

---

## Phase 8: x/ai Extension Stabilisation Path

Status: planned

Goals:

- define a clear promotion path for `x/ai` sub-packages from experimental to stable
- converge the AI provider interface to cover the existing Claude and OpenAI adapters
- fill integration-test gaps in `orchestration`, `semanticcache`, and `multimodal`

Background:

`x/ai` already has substantial depth: provider abstraction, Claude and OpenAI adapters,
multi-step orchestration workflows, semantic cache with vector storage, multimodal input,
SSE streaming output, AI-specific rate limiting and circuit breaking, and a tool-use
framework. The package is still marked experimental and contains internal development
draft documents (e.g. `SPRINT2_ENHANCEMENTS.md`) indicating some features are still
iterating.

Planned work:

- identify which `x/ai` sub-packages are candidates for stable API promotion
  (candidates: `provider`, `session`, `streaming`, `tool`)
- identify which sub-packages should remain experimental
  (candidates: `orchestration`, `semanticcache`, `marketplace`, `distributed`)
- remove internal development drafts such as `SPRINT2_ENHANCEMENTS.md`; migrate
  relevant content to formal docs or CHANGELOG
- add a mock implementation of `provider.Provider` for use in downstream tests
- add integration tests for multi-step agent scenarios in `orchestration.Workflow`
- add end-to-end tests for `semanticcache` covering offline vector retrieval and
  cache hit/miss paths
- document the stability tier of each sub-package in `docs/modules/x-ai`
- align the provider registration pattern in `x/ai/marketplace` with the
  `x/ai/provider` interface

Non-goals:

- do not promote `x/ai` to a stable root package
- do not implement business-level prompt flow logic inside `x/ai`
- do not introduce dependencies on `x/tenant` or the core bootstrap layer

Exit criteria:

- the public interfaces of `x/ai/provider`, `x/ai/session`, `x/ai/streaming`, and
  `x/ai/tool` have documented stability tiers
- no internal development draft documents remain in package directories
- the AI provider mock is directly usable by downstream test suites
- all tests pass under `-race`

---

## Phase 9: x/tenant Multi-Tenancy Production Readiness

Status: planned

Goals:

- advance `x/tenant` from experimental to a production-reference implementation
- provide production-grade JWT-based tenant ID extraction (replacing simple header reads)
- add a persistent distributed quota storage adapter for multi-instance deployments

Background:

`x/tenant` already has a complete multi-tenancy middleware stack
(resolve → ratelimit → quota → policy), but has the following production gaps:
- default tenant resolution reads `X-Tenant-ID` from a request header; no JWT claim
  extraction path exists
- quota storage defaults to an in-memory implementation (`InMemoryQuotaStore`) that
  cannot share state across instances
- the SQL schema and migration scripts for `DBTenantConfigManager` are undocumented
- `TenantDB` query interception supports only single-table `WHERE tenant_id = ?`
  and does not transparently inject tenant filtering into JOIN queries

Planned work:

- add a JWT claim extraction adapter in `x/tenant/resolve`
- add a Redis-backed distributed quota storage adapter in `x/tenant/store`
- publish the `x/tenant/config` database schema as SQL migration files under
  `docs/migrations/`
- document the query interception scope of `TenantDB` and explicitly note unsupported
  SQL patterns in the package README
- add end-to-end integration tests for `x/tenant` covering quota exhaustion and
  `Retry-After` header validation
- add a production deployment guidance section to
  `docs/architecture/X_TENANT_BLUEPRINT.md`

Non-goals:

- do not implement tenant CRUD business APIs inside `x/tenant`; that belongs to each
  application's domain layer
- do not introduce any tenant-aware code into stable root packages

Exit criteria:

- the JWT tenant resolution path has a complete example and test coverage
- the Redis quota storage adapter passes concurrent stress tests
- `docs/migrations/` contains the tenant-related schema files
- the supported SQL scope of `TenantDB` is documented in the package README

---

## Phase 10: x/discovery Backend Expansion

Status: planned

Goals:

- implement the service discovery backends annotated as "future" in `x/discovery`
- keep the explicit integration path with the `x/gateway` reverse proxy

Background:

`x/discovery` currently has two backend implementations:
- `discovery.NewStatic(...)` — static configuration
- `discovery.NewConsul(...)` — Consul

Kubernetes and etcd are named as future backends in package comments but are not yet
implemented. The `x/gateway` reverse proxy depends on the `discovery.Provider`
interface, which is the only coupling point between the two packages.

Planned work:

- implement a Kubernetes backend for `x/discovery` using the Endpoints API or
  EndpointSlices
- implement an etcd backend for `x/discovery` using the etcd v3 client interface
  with explicit dependency injection
- add independent integration tests for each new backend (using Docker or
  test containers)
- update the backend selection guide in `docs/modules/x-discovery`
- ensure all new backend dependencies flow through explicit constructors; no global
  `init`-style registration

Non-goals:

- do not implement a ZooKeeper or other registry backend (Kubernetes and etcd are the
  priority)
- do not place discovery logic inside stable root packages

Exit criteria:

- the Kubernetes and etcd backends each have a complete example and mock-based tests
- all `x/discovery` backends satisfy the same interface contract
- the integration path with `x/gateway` has a reference demo

---

## Phase 11: x/data Layer Stabilisation

Status: planned

Goals:

- clarify the stabilisation priority of the three `x/data` sub-packages
  (`file`, `rw`, `sharding`)
- provide clear production usage guidance for the read/write split (`rw`) and
  sharding (`sharding`) adapters
- register `x/fileapi` in the extension taxonomy

Background:

`x/data` contains:
- `x/data/file`: tenant-aware file storage implementation
- `x/data/rw`: read/write split database adapter
- `x/data/sharding`: shard routing with JSON-configured cluster and load balancing

`x/fileapi` is a standalone HTTP file upload/download handler that currently does not
appear in `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md` or the extension paths in
`specs/repo.yaml`, making it an architectural orphan.

Planned work:

- add `x/fileapi` to the `extension.paths` list in `specs/repo.yaml`
- add `x/fileapi` to `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- document the usage scope and limitations of `x/data/rw` (read-replica lag
  handling, failover strategy)
- add sharding strategy documentation and configuration examples for
  `x/data/sharding`
- verify that `x/data` and `x/fileapi` do not depend on anything outside the stable
  root packages
- add end-to-end handler tests for `x/fileapi` covering multipart upload, download,
  and tenant isolation

Non-goals:

- do not promote `x/data` to a stable root package
- do not implement business-level data access logic inside `x/data`

Exit criteria:

- `x/fileapi` appears in the architecture blueprint and the `specs/repo.yaml`
  extension list
- `x/data/rw` and `x/data/sharding` each have usage documentation
- `x/fileapi` handler tests cover upload, download, and tenant isolation paths

---

## Phase 12: x/observability and x/gateway Test Coverage

Status: planned

Goals:

- raise `x/observability` and `x/gateway` test coverage from minimal to
  feature-level
- ensure `x/gateway` circuit breaking, load balancing, and caching behaviour are
  protected by explicit tests

Background:

Current test file counts:
- `x/observability`: 1 test file (basic config coverage only)
- `x/gateway`: 4 test files (primarily routing and proxy forwarding)

`x/gateway` has sub-packages `cache`, `protocol`, and `protocolmw` whose behaviour
is mostly untested at the sub-package level. The Prometheus and OpenTelemetry
integration paths in `x/observability` have almost no test coverage.

Planned work:

- add tests for Prometheus metric registration and export paths in `x/observability`
- add span attribute tests for the OpenTelemetry tracer hook in `x/observability`
- add independent tests for cache hit, miss, and eviction in `x/gateway/cache`
- add distribution uniformity tests for round-robin and weighted load balancing in
  `x/gateway`
- add open/half-open state transition tests for the circuit breaker in `x/gateway`
- add configuration-level tests for TLS passthrough and protocol adapters in
  `x/gateway`

Non-goals:

- do not introduce external service dependencies; all tests use `httptest` or mocks
- do not change any stable public API

Exit criteria:

- `x/observability` has at least 4 test files covering both metrics and tracing paths
- load balancing and circuit breaking behaviour in `x/gateway` each have explicit tests
- all new tests pass under `-race`

---

## Phase 13: Developer Experience and Documentation Sync

Status: planned

Goals:

- let a new user go from `plumego new` to a running service without reading broad prose
- keep `README_CN.md` content synchronized with `README.md`
- add runnable minimal examples for `x/ai`, `x/data`, and `x/tenant`

Background:

- `README_CN.md` risks falling behind `README.md` now that the scaffold (Phase 5) and
  v1 API freeze (Phase 6) work is complete
- `x/ai` has substantial functionality but lacks a single end-to-end runnable minimal
  example (query one provider and print the response)
- `README.md` references `examples/multi-tenant-saas/` but the actual path is under
  `reference/`; the link needs correcting

Planned work:

- synchronize the following sections of `README_CN.md` with `README.md`:
  Agent-First Workflow, v1 Support Matrix, Development Server with Dashboard,
  Configuration Reference table
- add a `reference/with-ai` demo: a single provider completing one chat completion
  request using a mock backend
- correct the `examples/multi-tenant-saas/` path reference in `README.md`
- add a sub-package selection guide to `docs/modules/x-ai` explaining when to use
  `provider` versus `orchestration`
- confirm that `env.example` contains all AI provider environment variables
  (`OPENAI_API_KEY`, `ANTHROPIC_API_KEY`)

Non-goals:

- do not introduce live network dependencies in reference examples; AI provider calls
  use a mock
- do not change the overall structure of `README.md`

Exit criteria:

- `README_CN.md` is content-equivalent to `README.md` across all sections
- `env.example` covers all environment variables required by `x/*` extensions
- `reference/with-ai` exists and `go run .` succeeds using a mock provider
- all `examples/` path references in `README.md` point to directories that exist

---

## Phase 14: Extension Layer API Freeze Candidate Evaluation

Status: planned (depends on Phases 7–13)

Goals:

- evaluate the stability of the highest-usage `x/*` packages
- determine which `x/*` packages can be promoted to GA in a future major version
- update `docs/DEPRECATION.md` to cover the extension layer promotion process

Background:

All `x/*` packages are currently `status: experimental` with no compatibility
guarantee. As `x/tenant`, `x/rest`, `x/ai`, and related packages mature, a formal
evaluation and promotion mechanism is needed rather than keeping everything permanently
experimental.

Planned work:

- define a three-stage promotion standard for `x/*` packages:
  - experimental: API may change, no compatibility guarantee
  - stable-candidate: API frozen for at least 2 minor releases, no breaking changes,
    full test coverage
  - GA: enters the major-version compatibility commitment
- add a `stability_track` field to each extension package's `module.yaml`
- run a first-round evaluation of: `x/rest`, `x/websocket`, `x/webhook`,
  `x/scheduler`
- add an extension layer promotion section to `docs/DEPRECATION.md`

Non-goals:

- do not actually promote any package's compatibility level in this phase; only
  complete the evaluation framework
- do not modify the existing compatibility commitment for stable root packages

Exit criteria:

- the promotion standard document exists and is agent-discoverable
- each candidate package's `module.yaml` records its current stability track
- the extension layer section of `docs/DEPRECATION.md` is complete

---

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
7. Polish CLI code generation quality (Phase 7).
8. Stabilize `x/ai` sub-package tiers and clean internal drafts (Phase 8).
9. Advance `x/tenant` toward production readiness (Phase 9).
10. Add Kubernetes and etcd discovery backends (Phase 10).
11. Place `x/fileapi` in taxonomy; document `x/data` data patterns (Phase 11).
12. Harden `x/observability` and `x/gateway` test coverage (Phase 12).
13. Sync docs, add `with-ai` reference demo (Phase 13).
14. Run first x/* stability evaluation cycle (Phase 14).

## What Not to Do

- do not reintroduce component-style compatibility APIs
- do not add new broad legacy roots for short-term convenience
- do not let feature demos become the canonical app path
- do not push tenant or topology-heavy logic back into stable roots
- do not let docs get ahead of code behavior
- do not mark `x/*` packages as GA without completing the Phase 14 evaluation framework
- do not add Kubernetes or cloud-provider dependencies to stable root packages
