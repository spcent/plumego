# Plumego Roadmap

This roadmap tracks live repository direction.

It is intentionally shorter than the earlier migration-era version: completed
foundation work is summarized only where it still informs current priorities.

## Current Baseline

Plumego already has the following in place:

- stable roots with explicit boundaries: `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`
- extension discovery and task-routing metadata under `specs/*`
- a single canonical application layout in `reference/standard-service`
- milestone, plan, card, and verify workflow assets under `tasks/*`
- repo-wide quality gates in `Makefile` and `.github/workflows/quality-gates.yml`
- stable-root compatibility policy in `docs/DEPRECATION.md`
- stable-root exported API baseline snapshots under `docs/stable-api/snapshots`
- release evidence checklist under `docs/release/PRE_V1_RELEASE_CHECKLIST.md`
- beta promotion checklist and card template under `docs/release/PROMOTION_CARD_TEMPLATE.md`
- `x/rest`, `x/websocket`, `x/gateway`, and `x/observability` promoted to `beta`
  at v0.1.0–v0.2.0 with release-backed API snapshots and owner sign-off on record
- `v1.0.0` tagged on May 15, 2026; release notes and evidence in `docs/release/v1.0.0.md`

The next stages are about hardening extensions, improving examples, and keeping
docs, manifests, specs, and references aligned.

Release status is tied to repository tags and verifiable gate output.

## Roadmap Principles

- Keep `core` as a kernel, not a feature catalog.
- Preserve `net/http` compatibility and explicit control flow.
- Keep one canonical entrypoint per capability family.
- Let `specs/*` and module manifests carry machine-readable authority.
- Document only implemented behavior; remove stale drafts instead of preserving them in the active docs surface.

## Foundation Status

These phase labels remain because older cards and docs still reference them:

- Phase 1: agent workflow control plane — complete
- Phase 2: extension taxonomy convergence — complete
- Phase 3: `x/rest` reusable resource-interface hardening — substantially complete
- Phase 4: stable-root migration debt reduction — complete
- Phase 5: reference and scaffold system — complete
- Phase 6: release-readiness baseline — substantially complete
- Phase 7: CLI code-generation quality — complete

## Phase 8: `x/ai` Stabilisation Path

Status: substantially complete

Current state:

- `x/ai` remains experimental at the module level.
- `x/ai/module.yaml` already distinguishes stable subpackage tiers
  (`provider`, `session`, `streaming`, `tool`) from experimental ones
  (`orchestration`, `semanticcache`, `marketplace`, `distributed`, `resilience`).
- new AI service work should start with the stable-tier path:
  `provider` for provider contracts, `session` for state, `tool` for explicit
  tool policy and invocation, and `streaming` only when streaming coordination is
  required
- stable-tier packages (`provider`, `session`, `tool`) now have deepened contract
  tests covering Manager delegation, routing, session lifecycle, auto-trim, and
  builtin tool execution
- provider adapters (`ClaudeProvider`, `OpenAIProvider`) tested offline via
  `httptest.NewServer` covering constructor defaults, option overrides, successful
  completion, and API error paths
- `x/ai/marketplace` Manager contract tests cover PublishAgent, GetAgent,
  SearchAgents, ListAgentVersions, RateAgent, IsAgentInstalled, and the
  InstallAgent error path, ListInstalledAgents, and install/uninstall
  round-trips; the previous `UpdateDownloadCount` deadlock and installation
  metadata round-trip bug are fixed
- `docs/modules/x/ai/README.md` updated to list current test coverage

Next work:

- keep orchestration, semantic cache, marketplace, and distributed workflows explicitly experimental until their contracts settle
- create follow-up beta-evaluation cards per stable-tier subpackage only after
  release-history evidence proves two consecutive minor releases without
  exported API changes
- expand streaming contract tests if streaming primitives evolve

Non-goals:

- do not promote `x/ai` to a stable root package
- do not add hidden provider globals or implicit registration
- do not push transport or bootstrap concerns into `x/ai`

## Phase 9: `x/tenant` Production Readiness

Status: substantially complete

Current state:

- `x/tenant` already has dedicated families for `resolve`, `policy`, `quota`, `ratelimit`, `config`, `session`, and tenant-aware `store` adapters
- tenant-aware logic is intentionally excluded from stable `middleware` and stable `store`
- runnable offline examples now cover principal-first and custom-extractor tenant resolution flows
- tenant-aware `store/db` docs and tests now spell out the supported query-scoping subset and fail-closed misconfiguration behavior
- quota, policy, and rate-limit coverage now includes `Retry-After`, canonical deny responses, and tenant-scoped isolation checks
- `x/tenant/integration_test.go` covers the end-to-end resolve → policy → quota → ratelimit chain with tenant isolation verification
- `docs/modules/x/tenant/README.md` updated to document integration test coverage

Next work:

- add broader production-oriented resolution examples only when additional transport patterns are exercised in code
- extend `docs/architecture/X_TENANT_BLUEPRINT.md` only as implemented behavior changes land

Non-goals:

- do not move tenant concerns into stable roots
- do not turn `x/tenant` into application-specific tenant CRUD or onboarding logic

## Phase 10: `x/gateway/discovery` Backend Expansion

Status: substantially complete

Current state:

- `x/gateway/discovery` exposes static, Consul, Kubernetes, and etcd backends
- Kubernetes backend uses the Endpoints API with in-cluster auto-detection
- etcd backend uses the v3 HTTP gateway with explicit registration and health management
- all four backends implement the `Discovery` interface via explicit constructors
- `docs/modules/x/gateway/README.md` documents `x/gateway/discovery` backend selection guidance and standard validation

Next work:

- keep discovery concerns out of stable roots and out of bootstrap defaults
- expand backends only when additional infrastructure patterns are exercised in code

Non-goals:

- do not add hidden background registration
- do not couple discovery to gateway-only transport policy

## Phase 11: `x/data` and `x/fileapi` Hardening

Status: substantially complete

Current state:

- `x/fileapi` handler tests cover upload, download, cross-tenant isolation, info, delete, list, and signed URL paths
- `x/data/file` helper and metadata tests cover `isPathSafe`, `mimeToExt`, `extToMime`, round-trip behavior, nil metadata database guards, metadata scan unmarshalling, and configurable metadata clocks
- `x/data/rw`, `x/data/sharding`, `x/data/idempotency`, and `x/data/kvengine` all have tests
- `docs/modules/x/data/README.md` documents failover, read-after-write, sharding strategies, and boundary rules
- `docs/modules/x/fileapi/README.md` documents transport boundary and delegation to `x/data/file`
- `store/file` boundary is documented in `docs/modules/store/README.md`

Next work:

- extend `x/data/file` metadata coverage only when new SQL dialects, migration paths, or persistence semantics land
- extend `x/fileapi` example tests if new transport patterns land

## Phase 12: `x/observability` and `x/gateway` Test Depth

Status: substantially complete

Current state:

- `x/observability`: `PrometheusExporter.Handler()` output format, Content-Type header, and empty-collector behaviour now tested alongside existing collector and configuration tests
- `x/gateway`: `newBackendCircuitBreaker` default-config and Trip/Reset paths now tested; `entrypoints.go` functions (`NewGateway`, `NewGatewayBackendPool`, `NewGatewayProtocolRegistry`, `RegisterRoute`, `RegisterProxy`) all have tests; balancer, backend, health, proxy, rewrite, and transform packages were already well covered

Next work:

- do not introduce new stable-root API surface just to support tests
- do not add external-service requirements to the default test loop

## Phase 13: Docs and Onboarding Sync

Status: substantially complete (2026-05-20)

Work completed:

- the repository now has four distinct control surfaces: `docs/`, `specs/`, `tasks/`, and `reference/`
- `README.md` and `README_CN.md` are structurally aligned
- `docs/getting-started.md` matches the actual API surface
- `env.example` now includes `AUTH_TOKEN` (used by `x/observability/ops` but previously missing)
- module primers for `x/tenant`, `x/ai`, and `middleware` updated with current test coverage
- user-facing scenario entrypoint maps now identify the first reads for REST API,
  multi-tenant API, edge gateway, realtime, AI, and observability work without
  treating `x/*` packages as application bootstrap paths
- `docs/AGENT_FIRST.md` added as external-facing agent-first design overview
- `docs/README.md` updated to distinguish `agent-first.md` (internal operating reference)
  from `AGENT_FIRST.md` (external design overview)
- website module primers added for `x/rpc`, `x/openapi`, and `x/validate`
- `x/tenant` and `x/frontend` stability labels corrected to Beta across website and sidebar

Ongoing (continuous):

- keep `README.md` and `README_CN.md` aligned in scope and meaning as features land
- keep `docs/getting-started.md` and module primers aligned with `reference/standard-service`
- trim stale historical drafts instead of leaving them in the active docs surface

Non-goals:

- do not document planned features as if they already exist
- do not let workflow docs drift away from the live `Makefile` and CI setup

## Phase 14: Extension Stability Evaluation

Status: complete

Current state:

- `docs/EXTENSION_STABILITY_POLICY.md` defines the `experimental` → `beta` → `ga`
  criteria, promotion process, and current candidate assessment
- `specs/extension-beta-evidence.yaml` tracks the machine-readable beta
  evidence model for candidate modules, release refs, exported API snapshot refs,
  owner sign-off, and blockers
- `docs/EXTENSION_MATURITY.md` provides the human-readable `x/*` maturity
  dashboard for status, risk, recommended entrypoint, validation, and blockers
- the `status` enum in `specs/module-manifest.schema.yaml` supports
  `experimental`, `beta`, and `ga`
- `x/rest`, `x/websocket`, `x/gateway`, and `x/observability` promoted to `beta`
  at v0.1.0–v0.2.0: API unchanged across both refs, release-backed snapshots in
  `docs/extension-evidence/snapshots/`, owner sign-off recorded, all blockers
  cleared in `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/BETA_EVIDENCE_TEMPLATE.md` updated to reflect the
  current promotion workflow including `owner_signoff` ledger field and the
  correct ten-step promotion checklist
- `docs/release/PROMOTION_CARD_TEMPLATE.md` added for standardized promotion PRs
- `x/rest` blocking gaps filled: CRUD negative-path tests (`entrypoints_test.go`)
  cover all 7 not-implemented methods, `RegisterResourceRoutes` nil-arg/edge cases,
  `NewPaginationMeta` boundary values, and `QueryBuilder` input rejection; primer
  updated with boundary rules and full coverage section
- `x/websocket` blocking gaps filled: hub lifecycle negative-path tests
  (`hub_lifecycle_test.go`) cover `Stop` idempotency, `Shutdown` (empty, with
  connections, context cancellation), capacity errors (`ErrHubFull`, `ErrRoomFull`,
  `ErrHubStopped`), `RangeConns` iteration and early return, `BroadcastRoom`/
  `BroadcastAll` no-op after stop, `Leave`/`RemoveConn` non-member no-op; primer
  updated with boundary section and coverage section
- `x/tenant`, `x/frontend`, and `x/messaging` (app-facing service) promoted to
  `beta` with two consecutive release refs (`v1.0.0`, `v1.1.0`), API unchanged
  across both refs, release-backed snapshots recorded, owner sign-off complete,
  all blockers cleared in `specs/extension-beta-evidence.yaml`
- selected surfaces under `x/ai` (`provider`, `session`, `streaming`, `tool`)
  and `x/data` (`file`, `idempotency`) confirmed as beta surfaces with two
  release refs, owner sign-off, and evidence recorded; parent families remain
  experimental
- `specs/task-routing.yaml` `beta_families` updated to include `x/tenant`,
  `x/frontend`, and `x/messaging`
- `docs/DEPRECATION.md` beta extension table updated to include all promoted families

## Phase 16: Active Extension Evaluations

Status: in progress

Current state:

- task card 1500 (x/tenant GA evaluation) is in the active queue pending
  `v1.2.0` release evidence; `x/tenant` is beta — GA requires one additional
  release ref with no API changes and a second owner sign-off cycle
- task card 1501 (x/ai stable-tier subpackages beta evaluation) is in the active
  queue; each subpackage (`provider`, `session`, `streaming`, `tool`) requires
  its own per-subpackage evaluation after `v1.2.0` release evidence
- task card 1502 (x/openapi beta evaluation) is in the active queue; `x/openapi`
  requires cleanup of stale milestone-card language in `module.yaml` and two
  consecutive release refs before promotion

Next work:

- execute task card 1500 once `v1.2.0` release evidence is available
- execute task card 1501 per stable-tier subpackage after `v1.2.0` release evidence
- execute task card 1502 after `x/openapi` release-history evidence is complete
- clarify `x/data` and `x/fileapi` operational guidance as topology surfaces mature
- expand `x/gateway/discovery` backends only when explicit adapters are ready

Non-goals:

- do not promote extensions without two consecutive release refs and owner sign-off
- do not bundle multiple subpackage evaluations into a single promotion card

## Phase 15: Community Onboarding Improvements

Status: substantially complete (2026-05-20)

Work completed:

- `README.md` and `README_CN.md`: scenario routing decision table added
- `docs/benchmarks/README.md`: benchmark methodology, medians vs. Chi/Gin/Echo,
  net/http compatibility cost explanation
- `docs/troubleshooting.md`: 14-category onboarding troubleshooting guide
- `docs/getting-started_CN.md`: Chinese onboarding guide (complete)
- `docs/modules/x/rpc/README.md`: expanded from 23 lines to full primer
  (server lifecycle, client pool, interceptors, gateway, wiring diagram)
- `docs/modules/x/openapi/README.md`: expanded from 7 lines to full primer
  (CLI quick start, hint file format, Go API, serialisation, boundary rules)
- `docs/modules/x/validate/README.md`: expanded from 18 lines to full primer
  (BindJSON/Bind usage, Validator interface, adapter pattern, error shape)
- `docs/modules/x/resilience/README.md`: usage examples added (circuit breaker,
  keyed rate limiter, HTTP middleware integration)
- `docs/ADOPTION_PATH.md`: gRPC/RPC and OpenAPI entries added
- `specs/change-recipes/`: three new recipes (add-websocket-room, add-ai-tool,
  add-grpc-method)
- `internal/checks/cross-extension-deps`: new boundary checker validates
  module.yaml forbidden_imports against actual Go imports in x/* packages
- task cards 1500–1502 added to active queue for B1/B2/B3 evaluation

Non-goals:

- do not add Chinese versions of all module primers in this phase
- do not promote extensions without release-history evidence

## Cross-Cutting Workstreams

### Release Evidence

- keep `docs/stable-api/README.md` and checked-in snapshots aligned with stable
  root API surface; compare against the previous release before each new release
- use `docs/release/PRE_V1_RELEASE_CHECKLIST.md` as the base evidence checklist
  before tagging any release candidate
- keep extension beta blockers in `specs/extension-beta-evidence.yaml` until
  two concrete release refs, matching snapshots, and owner sign-off exist

### Documentation

- keep `README.md`, `README_CN.md`, `AGENTS.md`, `docs/*`, and module primers synchronized
- let `specs/*` and manifests carry authority; let prose explain intent and usage
- remove stale placeholders and superseded drafts from the active docs surface
- keep the first-user path aligned with `docs/ADOPTION_PATH.md`

### Testing

- preserve the required validation order in `AGENTS.md`
- keep targeted tests next to changed behavior
- bias toward negative-path coverage for security, tenant, routing, and gateway work

### Tooling

- prefer checks that reduce ambiguity or drift
- keep scaffolds and reference apps aligned with the canonical style
- enforce scaffold/reference drift through `docs/SCAFFOLD_REFERENCE_CONTRACT.md`
  and `cmd/plumego/internal/scaffold` tests
- keep docs examples compatible with the current API surface
- run `go run ./internal/checks/extension-maturity` when extension manifest
  status or risk changes so the dashboard stays aligned
- run `go run ./internal/checks/cross-extension-deps` when x/* module.yaml
  forbidden_imports or actual imports change to confirm boundary compliance

## Suggested Execution Order

1. Keep Phase 13 and Phase 15 docs and onboarding sync continuous.
2. Keep `docs/EXTENSION_MATURITY.md` aligned with module manifests and evidence records.
3. Execute task card 1500 (x/tenant GA evaluation) once v1.2.0 release evidence is available.
4. Execute task card 1501 (x/ai stable-tier subpackages beta evaluation) per-subpackage after release evidence.
5. Execute task card 1502 (x/openapi module.yaml cleanup and beta evaluation) once release evidence is available.
6. Clarify `x/data` and `x/fileapi` operational guidance.
7. Expand `x/gateway/discovery` backends only when explicit adapters are ready.

## What Not to Do

- do not reintroduce component-style compatibility APIs
- do not add new broad legacy roots for short-term convenience
- do not let scenario reference apps replace the canonical app path
- do not move tenant or topology-heavy logic back into stable roots
- do not leave stale historical drafts inside the active docs surface
- do not mark `x/*` packages as GA without explicit policy, tests, and docs
