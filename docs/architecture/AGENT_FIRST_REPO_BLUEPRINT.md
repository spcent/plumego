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
- `templates`
- `examples`
- `docs`
- `specs`
- `internal`

## Canonical Implementation Path

Agents should treat these as the default read and write path:

1. `docs/CANONICAL_STYLE_GUIDE.md`
2. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
3. `specs/repo.yaml`
4. `specs/dependency-rules.yaml`
5. `<module>/module.yaml`
6. `reference/standard-service`

Rules:

- `reference/standard-service` is the only canonical application layout
- `templates/standard-service` must mirror the reference layout for scaffolding
- `examples/` demonstrate capabilities only; they do not define architecture
- new docs and generators should point to `reference/standard-service`, not to `examples/`

## Hard Rules

- Root package facade imports are forbidden.
- Stable packages must not depend on `x/*`.
- `core` is a kernel, not a feature catalog.
- `middleware` remains transport-only.
- Tenant-aware logic belongs in `x/tenant`, not in stable `middleware` or `store`.
- Reference apps define the canonical app layout; examples do not.
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

Avoid growing broad buckets such as:

- `middleware/tenant`
- `middleware/observability` as a catch-all feature catalog
- `contract/protocol` as a cross-protocol family root
- `health` HTTP handler packages
- `store/db/rw` or `store/db/sharding` as stable-layer defaults

## Canonical Read Path

1. `docs/CANONICAL_STYLE_GUIDE.md`
2. `specs/repo.yaml`
3. `specs/dependency-rules.yaml`
4. `<module>/module.yaml`
5. module code

## Migration Direction

Near-term restructuring follows this order:

1. Add specs and module manifests.
2. Freeze the canonical app path: reference app plus matching template root.
3. Introduce and harden `x/tenant` as the first extension boundary.
4. Remove the root package facade.
5. Move feature catalogs and topology-heavy capabilities out of stable roots.
6. Replace broad category roots such as `net`, `utils`, `validator`, `rest`, and `pubsub`.
