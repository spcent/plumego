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

Status: substantially complete

Current state:

- `x/ai` remains experimental at the module level.
- `x/ai/module.yaml` already distinguishes stable subpackage tiers
  (`provider`, `session`, `streaming`, `tool`) from experimental ones
  (`orchestration`, `semanticcache`, `marketplace`, `distributed`, `resilience`).
- stable-tier packages (`provider`, `session`, `tool`) now have deepened contract
  tests covering Manager delegation, routing, session lifecycle, auto-trim, and
  builtin tool execution
- `docs/modules/x-ai/README.md` updated to list current test coverage

Next work:

- keep orchestration, semantic cache, marketplace, and distributed workflows explicitly experimental until their contracts settle
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
- `docs/modules/x-tenant/README.md` updated to document integration test coverage

Next work:

- add broader production-oriented resolution examples only when additional transport patterns are exercised in code
- extend `docs/architecture/X_TENANT_BLUEPRINT.md` only as implemented behavior changes land

Non-goals:

- do not move tenant concerns into stable roots
- do not turn `x/tenant` into application-specific tenant CRUD or onboarding logic

## Phase 10: `x/discovery` Backend Expansion

Status: substantially complete

Current state:

- `x/discovery` exposes static, Consul, Kubernetes, and etcd backends
- Kubernetes backend uses the Endpoints API with in-cluster auto-detection
- etcd backend uses the v3 HTTP gateway with explicit registration and health management
- all four backends implement the `Discovery` interface via explicit constructors
- `docs/modules/x-discovery/README.md` documents backend selection guidance and standard validation

Next work:

- keep discovery concerns out of stable roots and out of bootstrap defaults
- expand backends only when additional infrastructure patterns are exercised in code

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
- `README.md` and `README_CN.md` are structurally aligned
- `docs/getting-started.md` matches the actual API surface
- `env.example` now includes `AUTH_TOKEN` (used by `x/ops` but previously missing)
- module primers for `x/tenant`, `x/ai`, and `middleware` updated with current test coverage

Next work:

- keep `README.md` and `README_CN.md` aligned in scope and meaning as features land
- keep `docs/getting-started.md` and module primers aligned with `reference/standard-service`
- trim stale historical drafts instead of leaving them in the active docs surface

Non-goals:

- do not document planned features as if they already exist
- do not let workflow docs drift away from the live `Makefile` and CI setup

## Phase 14: Extension Stability Evaluation

Status: in progress

Current state:

- `docs/EXTENSION_STABILITY_POLICY.md` defines the `experimental` → `beta` → `ga`
  criteria, promotion process, and current candidate assessment
- the `status` enum in `specs/module-manifest.schema.yaml` already supports
  `experimental`, `beta`, and `ga`
- no extension has been promoted yet; policy is now in place

Next work:

- apply the policy to `x/rest` and `x/websocket` first (highest maturity)
- fill the gaps listed in the policy's current-evaluation table before promoting
- promote `status` in `module.yaml` only after all criteria are verified
- extend `docs/DEPRECATION.md` with a cross-reference when the first `beta`
  promotion lands

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
