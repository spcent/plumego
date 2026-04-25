# Agent-First Repo Blueprint

This document defines the target repository shape for Plumego.

## Goals

- Keep the stable surface small and explicit.
- Make module ownership easy for agents to identify.
- Minimize search radius for routine feature work.
- Prevent extension capabilities from polluting the stable core.
- Establish one canonical implementation path for new work.

## Repository Shape

Stable packages stay at the repository root:

- `core`
- `router`
- `contract`
- `middleware`
- `security`
- `store`
- `health`
- `log`
- `metrics`

Optional or fast-moving capabilities live under `x/`:

- `x/ai`
- `x/cache`
- `x/data`
- `x/devtools`
- `x/discovery`
- `x/fileapi`
- `x/frontend`
- `x/gateway`
- `x/ipc`
- `x/messaging`
- `x/mq`
- `x/observability`
- `x/ops`
- `x/pubsub`
- `x/resilience`
- `x/rest`
- `x/scheduler`
- `x/tenant`
- `x/webhook`
- `x/websocket`

Non-library areas stay out of import-path design:

- `cmd`
- `reference`
- `docs`
- `specs`
- `internal`
- `tasks`

## Repository Control Surfaces

The repository uses four different top-level surfaces on purpose:

- `docs/`: human-readable explanation, architecture decisions, module primers, and roadmap
- `specs/`: machine-readable repository rules, ownership data, dependency constraints, and change recipes
- `tasks/`: repo-native execution cards that agents can consume as a work queue
- `reference/`: canonical application wiring and runnable reference examples

This split is intentional and should remain stable:

- do not move `specs/` under `docs/`; that would blur executable rules with explanatory prose
- keep machine-readable repository contracts at the top level so discovery and checks stay simple
- keep task cards outside `docs/` when they are meant to drive execution rather than serve as archival prose
- keep `reference/standard-service` as the only canonical application layout; feature demos must not replace it

## Agent Control Plane

Agents should treat the repository control plane as a layered contract:

1. `AGENTS.md`: operating rules, task contracts, validation order, milestone protocol
2. `docs/CODEX_WORKFLOW.md`: repeatable prompting and execution patterns
3. `docs/CANONICAL_STYLE_GUIDE.md`: canonical implementation path and code shape
4. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`: repository shape and ownership model
5. `specs/*.yaml`: machine-readable routing, dependency, taxonomy, and manifest rules
6. `<module>/module.yaml`: module-local responsibilities, review checklist, and validation commands
7. `reference/standard-service`: canonical application wiring

The intent is not just documentation. Each layer should make it harder for an
agent to guess incorrectly about module ownership, allowed imports, or the
canonical way to implement a change.

## Canonical Implementation Path

Agents should treat these as the default read and write path:

1. `AGENTS.md`
2. `docs/CODEX_WORKFLOW.md`
3. `docs/CANONICAL_STYLE_GUIDE.md`
4. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
5. `specs/repo.yaml`
6. `specs/task-routing.yaml`
7. `specs/extension-taxonomy.yaml`
8. `specs/package-hotspots.yaml`
9. `specs/dependency-rules.yaml`
10. `<module>/module.yaml`
11. `reference/standard-service`

Rules:

- `reference/standard-service` is the only canonical application layout
- `reference/standard-service` must depend only on stable root packages and the standard library
- extension or feature demos must live outside `reference/standard-service`
- each extension family must publish one canonical discovery entrypoint

## Task Contract Defaults

When humans do not provide an explicit task contract, agents should default to:

- one primary module per change
- no stable public API changes
- no new dependencies
- focused tests for behavior changes
- docs sync only for implemented behavior changes

Large work should be split into an analysis pass and an implementation pass.
Analysis identifies module ownership, in-scope paths, out-of-scope paths,
likely touched files, and validation commands before coding begins.

## Hard Rules

- Root package facade imports are forbidden.
- Stable packages must not depend on `x/*`.
- `core` is a kernel, not a feature catalog.
- `middleware` remains transport-only.
- Tenant-aware logic belongs in `x/tenant`, not in stable `middleware` or `store`.
- Stable `middleware` must not grow tenant resolution, tenant policy, or tenant quota behavior.
- Stable `store` must not grow tenant-aware adapters or tenant-specific storage policy.
- Reference apps define the canonical app layout.
- `health` keeps models and readiness state, not HTTP endpoint ownership.
- `contract` keeps transport contracts, not protocol gateway families.
- `store` stable layer keeps primitives; topology-heavy data features move to extensions.

## Target Stable-Layer Boundaries

- `core`: app lifecycle, route attachment, middleware attachment, server startup and shutdown
- `router`: matching, params, groups, reverse routing
- `contract`: error model, response helpers, request metadata helpers
- `middleware`: narrow transport middleware packages only
- `security`: auth, headers, input safety, abuse guard primitives
- `store`: base persistence primitives only
- `health`, `log`, `metrics`: support contracts and base implementations only

Within observability concerns:

- stable `middleware/*` owns transport observability primitives
- `x/observability` owns broader adapter and export wiring

Health HTTP exposure belongs in reference apps or owning extensions, not in the stable `health` root.

Avoid growing broad buckets such as:

- `middleware/tenant`
- `middleware/observability` as a catch-all feature catalog
- `contract/protocol` as a cross-protocol family root
- `health` HTTP handler packages
- `store/db/rw` or `store/db/sharding` as stable-layer defaults

## Canonical Read Path

1. `AGENTS.md`
2. `docs/CODEX_WORKFLOW.md`
3. `docs/CANONICAL_STYLE_GUIDE.md`
4. `specs/repo.yaml`
5. `specs/task-routing.yaml`
6. `specs/extension-taxonomy.yaml`
7. `specs/package-hotspots.yaml`
8. `specs/dependency-rules.yaml`
9. `<module>/module.yaml`
10. module code

## Machine-Readable Agent Workflow

The repository should expose enough machine-readable metadata that an agent can
decide where to start, who owns a boundary, and what recipe to follow without
guessing.

Required metadata lives under `specs/`:

- `specs/task-routing.yaml`: task-to-entrypoint map and disallowed first reads
- `specs/extension-taxonomy.yaml`: canonical extension families and subordinate roots
- `specs/package-hotspots.yaml`: ambiguity hotspot packages and first-read paths
- `specs/change-recipes/*.yaml`: standard task recipes for common change shapes

Human-readable module primers live under `docs/modules/` and should mirror
manifest-declared `doc_paths`.

Module manifests are part of the control plane, not just package notes. Each
`<module>/module.yaml` should declare:

- summary and strict boundary
- responsibilities and non-goals
- allowed and forbidden imports
- test commands
- review checklist
- agent hints
- doc paths

Stable roots must declare a non-empty `strict_boundary` so agents and checks can
distinguish kernel, router, contract, middleware, and other stable roles
without inferring from package names alone.

## Execution Loop

The default agent loop is:

1. classify the owning layer and module
2. read the control plane in canonical order
3. declare intended scope and touched files
4. implement the smallest coherent change
5. run module validation first, then boundary and repo checks
6. report residual risks and doc sync needs

If step 1 or 2 leaves the owning module unclear, the agent should stop in
analysis mode instead of coding through ambiguity.

## Migration Direction

Near-term restructuring follows this order:

1. Add specs and module manifests.
2. Freeze the canonical app path: reference app plus matching template root.
3. Introduce and harden `x/tenant` as the first extension boundary.
4. Remove the root package facade.
5. Move feature catalogs and topology-heavy capabilities out of stable roots.
6. Replace broad legacy top-level roots such as `net`, `utils`, `validator`, `rest`, and `pubsub` with stable roots or explicit `x/*` families.

## Extension Discovery Defaults

Agents should prefer these entrypoints when multiple related `x/*` packages exist:

- Start messaging-related work in `x/messaging`; open `x/mq` or `x/pubsub` only when you already know the task is a queue primitive or broker primitive.
- Treat `x/webhook` as a messaging sub-capability by default; start directly in `x/webhook` only for narrow webhook verification or delivery mechanics.
- Start gateway and edge transport work in `x/gateway`; treat `x/ipc` as a narrow primitive.
- Start reusable resource-interface and CRUD-standardization work in `x/rest`; keep bootstrap shape in `reference/standard-service` and edge proxy topology in `x/gateway`.
- Start broader observability adapter and export work in `x/observability`; use `x/ops` only for protected admin endpoints and auth-gated diagnostics surfaces.
- Start reusable resilience primitives in `x/resilience`; do not park them under stable `security` or a feature package unless the behavior is truly feature-specific.
- Start app-facing file upload and download transport work in `x/fileapi`; keep tenant-aware storage implementations in `x/data/file` and pure storage interfaces in `store/file`.
- Start frontend asset-serving work in `x/frontend`, but do not let frontend helpers define the canonical app path.
- Start local debug and developer-only route work in `x/devtools`; do not treat debug surfaces as part of `core`.
- Start service discovery work in `x/discovery`; do not spread discovery concerns across stable roots.
- Start transport observability work in stable `middleware/*` packages; use `x/observability` only for higher-level adapter or export wiring.
- Do not start new app structure from `x/rest`; prefer `reference/standard-service` and explicit route binding.
- Treat `x/ipc` as a narrow transport helper, not the default home for general eventing or workflow features.
