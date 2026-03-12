# Agent-First Repo Blueprint

This document defines the target repository shape for Plumego.

## Goals

- Keep the stable surface small and explicit.
- Make module ownership easy for agents to identify.
- Minimize search radius for routine feature work.
- Prevent extension capabilities from polluting the stable core.

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

Non-library areas stay out of import-path design:

- `cmd`
- `reference`
- `templates`
- `examples`
- `docs`
- `specs`
- `internal`

## Hard Rules

- Root package facade imports are forbidden.
- Stable packages must not depend on `x/*`.
- `core` is a kernel, not a feature catalog.
- `middleware` remains transport-only.
- Tenant-aware logic belongs in `x/tenant`, not in stable `middleware` or `store`.
- Reference apps define the canonical app layout; examples do not.

## Canonical Read Path

1. `docs/CANONICAL_STYLE_GUIDE.md`
2. `specs/repo.yaml`
3. `specs/dependency-rules.yaml`
4. `<module>/module.yaml`
5. module code

## Migration Direction

Near-term restructuring follows this order:

1. Add specs and module manifests.
2. Introduce `x/tenant` as the first extension boundary.
3. Remove the root package facade.
4. Move feature components out of `core/components`.
5. Replace broad category roots such as `net`, `utils`, and `validator`.
