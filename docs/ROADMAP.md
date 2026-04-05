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

The next stages are about hardening extensions, improving examples, and keeping
docs, manifests, specs, and references aligned.

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

Status: in progress

Current state:

- `x/ai` remains experimental at the module level.
- `x/ai/module.yaml` already distinguishes stable subpackage tiers
  (`provider`, `session`, `streaming`, `tool`) from experimental ones
  (`orchestration`, `semanticcache`, `marketplace`, `distributed`, `resilience`).

Next work:

- keep `docs/modules/x-ai/README.md` aligned with the manifest-defined stability tiers
- add runnable examples that exercise provider, session, streaming, and tool composition without live network calls
- deepen tests around the stable-tier package contracts before any broader stability claims
- keep orchestration, semantic cache, marketplace, and distributed workflows explicitly experimental until their contracts settle

Non-goals:

- do not promote `x/ai` to a stable root package
- do not add hidden provider globals or implicit registration
- do not push transport or bootstrap concerns into `x/ai`

## Phase 9: `x/tenant` Production Readiness

Status: in progress

Current state:

- `x/tenant` already has dedicated families for `resolve`, `policy`, `quota`, `ratelimit`, `config`, `session`, and tenant-aware `store` adapters
- tenant-aware logic is intentionally excluded from stable `middleware` and stable `store`
- runnable offline examples now cover principal-first and custom-extractor tenant resolution flows
- tenant-aware `store/db` docs and tests now spell out the supported query-scoping subset and fail-closed misconfiguration behavior
- quota, policy, and rate-limit coverage now includes `Retry-After`, canonical deny responses, and tenant-scoped isolation checks

Next work:

- add broader production-oriented resolution examples only when additional transport patterns are exercised in code
- expand higher-level end-to-end coverage across combined resolution, policy, quota, and rate-limit middleware chains
- extend `docs/architecture/X_TENANT_BLUEPRINT.md` only as implemented behavior changes land

Non-goals:

- do not move tenant concerns into stable roots
- do not turn `x/tenant` into application-specific tenant CRUD or onboarding logic

## Phase 10: `x/discovery` Backend Expansion

Status: planned

Current state:

- `x/discovery` currently exposes static and Consul-backed discovery
- integration with `x/gateway` is through explicit provider interfaces

Next work:

- add Kubernetes and etcd backends only through explicit constructors
- add backend-specific tests and selection guidance in `docs/modules/x-discovery/README.md`
- keep discovery concerns out of stable roots and out of bootstrap defaults

Non-goals:

- do not add hidden background registration
- do not couple discovery to gateway-only transport policy

## Phase 11: `x/data` and `x/fileapi` Hardening

Status: planned

Current state:

- `x/fileapi` is already part of the extension taxonomy and architecture blueprint
- `x/data/file`, `x/data/rw`, and `x/data/sharding` exist and need clearer production guidance

Next work:

- document failover, read-after-write, and health expectations for `x/data/rw`
- document sharding strategy selection, routing limits, and configuration examples for `x/data/sharding`
- keep `x/fileapi`, `x/data/file`, and `store/file` boundary docs aligned
- add focused tests and examples for upload, download, metadata, and tenant-isolation paths as behavior changes land

Non-goals:

- do not promote `x/data` to a stable root
- do not collapse transport, storage, and topology responsibilities into one package

## Phase 12: `x/observability` and `x/gateway` Test Depth

Status: planned

Current state:

- both modules exist, but coverage depth is still uneven across important subpackages

Next work:

- raise coverage around tracing and metrics export paths in `x/observability`
- add explicit tests for cache, load-balancing, circuit-breaking, and protocol-adapter behavior in `x/gateway`
- keep test dependencies local, explicit, and fast enough for routine iteration

Non-goals:

- do not introduce new stable-root API surface just to support tests
- do not add external-service requirements to the default test loop

## Phase 13: Docs and Onboarding Sync

Status: in progress

Current state:

- the repository now has four distinct control surfaces: `docs/`, `specs/`, `tasks/`, and `reference/`
- onboarding docs must stay aligned with the current `Makefile`, manifests, and reference app

Next work:

- keep `README.md` and `README_CN.md` aligned in scope and meaning
- keep `docs/getting-started.md`, `docs/README.md`, and module primers aligned with `reference/standard-service`
- trim stale historical drafts instead of leaving them in the active docs surface
- update `env.example` only when implemented configuration or provider behavior changes

Non-goals:

- do not document planned features as if they already exist
- do not let workflow docs drift away from the live `Makefile` and CI setup

## Phase 14: Extension Stability Evaluation

Status: planned

Current state:

- extension modules are still treated as experimental by default
- stability guidance exists only in a few places today, such as `x/ai/module.yaml`

Next work:

- define repository-wide criteria for any future `stable-candidate` or GA extension track
- record extension stability track in manifests only if the repo adopts a shared policy
- evaluate likely candidates such as `x/rest`, `x/websocket`, `x/webhook`, and `x/scheduler` against that policy
- extend `docs/DEPRECATION.md` only after the policy is concrete

Non-goals:

- do not silently upgrade extension guarantees without explicit policy and tests
- do not weaken the stable-root compatibility promise

## Cross-Cutting Workstreams

### Documentation

- keep `README.md`, `README_CN.md`, `AGENTS.md`, `docs/*`, and module primers synchronized
- let `specs/*` and manifests carry authority; let prose explain intent and usage
- remove stale placeholders and superseded drafts from the active docs surface

### Testing

- preserve the required validation order in `AGENTS.md`
- keep targeted tests next to changed behavior
- bias toward negative-path coverage for security, tenant, routing, and gateway work

### Tooling

- prefer checks that reduce ambiguity or drift
- keep scaffolds and reference apps aligned with the canonical style
- keep docs examples compatible with the current API surface

## Suggested Execution Order

1. Keep Phase 13 docs and onboarding sync continuous.
2. Harden `x/ai` docs, examples, and stable-tier tests.
3. Advance `x/tenant` production readiness.
4. Clarify `x/data` and `x/fileapi` operational guidance.
5. Expand `x/discovery` backends only when explicit adapters are ready.
6. Raise `x/observability` and `x/gateway` test depth.
7. Define extension stability criteria before any promotion discussion.

## What Not to Do

- do not reintroduce component-style compatibility APIs
- do not add new broad legacy roots for short-term convenience
- do not let feature demos replace the canonical app path
- do not move tenant or topology-heavy logic back into stable roots
- do not leave stale historical drafts inside the active docs surface
- do not mark `x/*` packages as GA without explicit policy, tests, and docs
