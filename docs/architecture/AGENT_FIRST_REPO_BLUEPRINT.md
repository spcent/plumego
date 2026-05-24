# Agent-First Repo Blueprint

This document defines the target repository shape for Plumego.

## 1. Goals

- Keep the stable surface small and explicit.
- Make ownership easy to identify.
- Minimize search radius for routine work.
- Prevent extension capabilities from leaking into the stable core.
- Preserve one canonical implementation path for new work.

## 2. Repository Shape

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
- `x/data`
- `x/fileapi`
- `x/frontend`
- `x/gateway`
- `x/messaging`
- `x/observability`
- `x/openapi`
- `x/resilience`
- `x/rest`
- `x/rpc`
- `x/tenant`
- `x/validate`
- `x/websocket`

Non-library areas stay out of import-path design:

- `cmd`
- `reference`
- `docs`
- `specs`
- `internal`
- `tasks`

## 3. Control Surfaces

The repository intentionally keeps four top-level control surfaces:

- `docs/`: human-readable explanation, architecture, primers, and roadmap
- `specs/`: machine-readable routing, taxonomy, dependency rules, and recipes
- `tasks/`: bounded execution cards and milestones
- `reference/`: canonical application wiring and scenario apps

Keep this split stable:

- Do not move `specs/` under `docs/`
- Keep machine-readable contracts at the top level
- Keep task cards outside `docs/`
- Keep `reference/standard-service` as the canonical application layout

## 4. Agent Read Path

Default bounded read path:

1. `AGENTS.md`
2. matching `specs/task-routing.yaml` entry
3. that entry's `start_with` files
4. `<module>/module.yaml` before editing module behavior
5. targeted style, architecture, or reference docs only when preflight requires them

Use the full architecture and spec set only for control-plane, boundary, or
release work.

## 5. Canonical Defaults

- `reference/standard-service` is the only canonical application layout.
- Extension or scenario apps must not redefine the default app shape.
- Each extension family should expose one canonical discovery entrypoint.
- Constructor policy follows `docs/CANONICAL_STYLE_GUIDE.md`.
- New fallible constructors return errors; existing panic convenience wrappers are compatibility paths only.

## 6. Task Contract Defaults

When humans do not provide an explicit task contract, agents should assume:

- one primary module per change
- no stable public API changes
- no new dependencies
- focused tests for behavior changes
- docs sync only for implemented behavior changes

Broad work should be split into an analysis pass and an implementation pass.

## 7. Hard Rules

- Root package facade imports are forbidden.
- Stable packages must not depend on `x/*`.
- `core` is a kernel, not a feature catalog.
- `middleware` remains transport-only.
- Tenant-aware logic belongs in `x/tenant`, not stable `middleware` or `store`.
- `health` owns models and readiness state, not HTTP endpoint ownership.
- `contract` owns transport contracts, not protocol gateway families.
- `store` owns primitives; topology-heavy data features belong in extensions.
- Reference apps define the canonical app layout.

## 8. Target Stable-Layer Boundaries

- `core`: app lifecycle, route attachment, middleware attachment, startup, and shutdown
- `router`: matching, params, groups, and reverse routing
- `contract`: error model, response helpers, request metadata, and context helpers
- `middleware`: narrow transport middleware only
- `security`: auth, headers, input safety, and abuse-guard primitives
- `store`: base persistence primitives only
- `health`, `log`, `metrics`: support contracts and base implementations only

Observability split:

- Stable `middleware/*` owns transport observability primitives
- `x/observability` owns broader adapter and export wiring

## 9. Machine-Readable Workflow

The repository should expose enough metadata for an agent to route work without
guessing.

Required metadata lives under `specs/`:

- `specs/task-routing.yaml`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/change-recipes/*.yaml`

Each `<module>/module.yaml` should declare:

- summary and strict boundary
- responsibilities and non-goals
- allowed and forbidden imports
- test commands
- review checklist
- agent hints
- doc paths

Stable roots should declare a non-empty `strict_boundary`.

## 10. Execution Loop

1. Classify the owning layer and module.
2. Select the smallest matching context package.
3. Declare intended scope and touched files.
4. Implement the smallest coherent change.
5. Run module validation first, then boundary and repo checks.
6. Report residual risks and docs-sync impact.

If the owning module or context package remains unclear, stop in analysis mode.

## 11. Extension Discovery Defaults

- Start messaging work in `x/messaging`; open `x/messaging/mq`, `x/messaging/pubsub`, or `x/messaging/webhook` only when the task is already narrow.
- Start gateway and edge transport work in `x/gateway`; treat `x/gateway/ipc` as a narrow primitive.
- Start CRUD-standardization work in `x/rest`; keep bootstrap shape in `reference/standard-service`.
- Start broader observability adapter and export work in `x/observability`; use `x/observability/ops` only for protected admin surfaces.
- Start reusable resilience primitives in `x/resilience`.
- Start file upload and download transport work in `x/fileapi`; keep tenant-aware storage in `x/data/file`.
- Start frontend asset-serving work in `x/frontend`, not in the canonical app path.
- Start local debug and developer-only route work in `x/observability/devtools`.
- Start service discovery work in `x/gateway/discovery`.
- Start transport observability work in stable `middleware/*`; use `x/observability` only for higher-level adapter or export wiring.
- Start RPC transport work in `x/rpc`; narrow into `x/rpc/server`, `x/rpc/client`, or `x/rpc/gateway` only when the transport role is already known.
- Start OpenAPI document generation work in `x/openapi`.
- Start explicit request binding and validation bridge work in `x/validate`.
