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

Status: foundation in place

Goals:

- make `x/rest` the reusable home for resource-interface conventions
- preserve explicit route binding while maximizing CRUD reuse
- keep `x/rest` distinct from bootstrap and proxy topology concerns

Planned work:

- add route-level tests for `RegisterContextResourceRoutes(...)`
- add examples showing `ResourceSpec -> repository -> routes`
- continue moving query, pagination, hooks, and transformer behavior toward spec-driven configuration
- define clearer reusable error and response conventions where needed
- document recommended layering between handlers, repositories, and `x/rest` controllers

Exit criteria:

- resource APIs can be standardized without inventing per-service scaffolding
- `x/rest` stays clearly distinct from `reference/standard-service` and `x/gateway`
- spec-driven behavior is covered by focused tests and examples

## Phase 4: Shrink Stable-Root Migration Debt

Status: not started

Goals:

- keep stable roots narrow and durable
- move topology-heavy or feature-heavy logic out of stable packages
- align package placement with the architecture blueprint

Planned work:

- review `store/cache/distributed` and `store/cache/redis` placement against target stable-root rules
- continue checking for observability or protocol catch-all behavior in stable roots
- ensure transport health handlers remain outside `health`
- keep tenant-aware behavior out of stable `middleware` and `store`
- shrink `specs/check-baseline/*` entries instead of letting them grow

Exit criteria:

- stable roots read as long-lived primitives rather than convenience catalogs
- migration debt files shrink release over release
- architecture checks block new drift instead of merely documenting it

## Phase 5: Reference and Scaffold System

Status: not started

Goals:

- make the canonical app path easy to copy without reintroducing hidden patterns
- ensure scaffolds follow the same rules as the reference app

Planned work:

- add a minimal scaffold generator or template set based on `reference/standard-service`
- add reference variants outside the canonical path for `x/messaging`, `x/gateway`, `x/websocket`, and `x/webhook`
- keep feature references clearly marked as non-canonical
- add checks that canonical references do not drift into `x/*`

Exit criteria:

- new projects start from the same explicit structure every time
- feature demos do not pollute the canonical learning path
- scaffold output stays aligned with docs and specs

## Phase 6: Release Readiness Toward v1

Status: planned

Goals:

- turn the architecture cleanup into a durable release baseline
- make quality gates strict enough for stable public adoption

Planned work:

- close remaining migration debt tracked in `specs/check-baseline/*`
- audit public APIs for clarity before final v1 freeze
- expand race, integration, and negative-path coverage for critical roots
- run release-readiness reviews against documentation, examples, env defaults, and quality gates
- define a formal deprecation policy for future extension evolution

Exit criteria:

- public docs describe only the supported explicit APIs
- quality gates are green without new temporary exceptions
- maintainers can describe the supported architecture without caveats

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
