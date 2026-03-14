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

- `x/tenant`
- `x/ai`
- `x/websocket`
- `x/webhook`
- `x/scheduler`
- `x/frontend`
- `x/ops`
- `x/devtools`
- `x/messaging`
- `x/discovery`
- `x/gateway`
- `x/data`

Non-library areas stay out of import-path design:

- `cmd`
- `reference`
- `docs`
- `specs`
- `internal`
- `tasks`

## Repository Control Surfaces

The repository uses three different top-level control surfaces on purpose:

- `docs/`: human-readable explanation, architecture decisions, module primers, and roadmap
- `specs/`: machine-readable repository rules, ownership data, dependency constraints, and change recipes
- `tasks/`: repo-native execution cards that agents can consume as a work queue

This split is intentional and should remain stable:

- do not move `specs/` under `docs/`; that would blur executable rules with explanatory prose
- keep machine-readable repository contracts at the top level so discovery and checks stay simple
- keep task cards outside `docs/` when they are meant to drive execution rather than serve as archival prose

## Canonical Implementation Path

Agents should treat these as the default read and write path:

1. `docs/CANONICAL_STYLE_GUIDE.md`
2. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
3. `specs/repo.yaml`
4. `specs/agent-entrypoints.yaml`
5. `specs/dependency-rules.yaml`
6. `specs/ownership.yaml`
7. `<module>/module.yaml`
8. `reference/standard-service`

Rules:

- `reference/standard-service` is the only canonical application layout
- `reference/standard-service` must depend only on stable root packages and the standard library
- extension or feature demos must live outside `reference/standard-service`
- each extension family must publish one canonical discovery entrypoint

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

1. `docs/CANONICAL_STYLE_GUIDE.md`
2. `specs/repo.yaml`
3. `specs/agent-entrypoints.yaml`
4. `specs/dependency-rules.yaml`
5. `specs/ownership.yaml`
6. `<module>/module.yaml`
7. module code

## Machine-Readable Agent Workflow

The repository should expose enough machine-readable metadata that an agent can
decide where to start, who owns a boundary, and what recipe to follow without
guessing.

Required metadata lives under `specs/`:

- `specs/agent-entrypoints.yaml`: task-to-entrypoint map and disallowed first reads
- `specs/ownership.yaml`: owner, risk, and default validation per critical module
- `specs/change-recipes/*.yaml`: standard task recipes for common change shapes

Human-readable module primers live under `docs/modules/` and should mirror
manifest-declared `doc_paths`.

## Migration Direction

Near-term restructuring follows this order:

1. Add specs and module manifests.
2. Freeze the canonical app path: reference app plus matching template root.
3. Introduce and harden `x/tenant` as the first extension boundary.
4. Remove the root package facade.
5. Move feature catalogs and topology-heavy capabilities out of stable roots.
6. Replace broad category roots such as `net`, `utils`, `validator`, `rest`, and `pubsub`.

## Extension Discovery Defaults

Agents should prefer these entrypoints when multiple related `x/*` packages exist:

- Start messaging-related work in `x/messaging`; open `x/mq` or `x/pubsub` only when you already know the task is a queue primitive or broker primitive.
- Treat `x/webhook` as a messaging sub-capability by default; start directly in `x/webhook` only for narrow webhook verification or delivery mechanics.
- Start gateway and edge transport work in `x/gateway`; treat `x/ipc` as a narrow primitive.
- Start reusable resource-interface and CRUD-standardization work in `x/rest`; keep bootstrap shape in `reference/standard-service` and edge proxy topology in `x/gateway`.
- Start broader observability adapter and export work in `x/observability`; use `x/ops` only for protected admin endpoints and auth-gated diagnostics surfaces.
- Start frontend asset-serving work in `x/frontend`, but do not let frontend helpers define the canonical app path.
- Start local debug and developer-only route work in `x/devtools`; do not treat debug surfaces as part of `core`.
- Start service discovery work in `x/discovery`; do not spread discovery concerns across stable roots.
- Start transport observability work in stable `middleware/*` packages; use `x/observability` only for higher-level adapter or export wiring.
- Do not start new app structure from `x/rest`; prefer `reference/standard-service` and explicit route binding.
- Treat `x/ipc` as a narrow transport helper, not the default home for general eventing or workflow features.
